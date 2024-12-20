package namenode

import (
	"errors"
	"math/rand"
	"time"

	"github.com/mrowaha/dos/api"
	"google.golang.org/grpc"
)

var (
	ErrToleranceNotEnough = errors.New("not enough ghost nodes connected to maintain tolerance")
	ErrNoGhostNode        = errors.New("no ghost node exists in map")
)

type GhostNodeEntry struct {
	Id        string
	CommandCh chan<- map[string][]byte
	CloseCh   <-chan struct{}
}

type GhostNodesMap map[string]*GhostNodeEntry

func (s *DosNameNodeServer) randomGhost() *GhostNodeEntry {
	rand.Seed(time.Now().UnixNano())
	keys := make([]string, 0, len(s.ghosts))
	for key := range s.ghosts {
		keys = append(keys, key)
	}

	if len(keys) > 0 {
		randomKey := keys[rand.Intn(len(keys))]
		randomValue := s.ghosts[randomKey]
		return randomValue
	}
	return nil
}

func (s *DosNameNodeServer) InitiateSpawn(failedNode string) error {

	s.logger.Printf("initiating spawn protocol\n")

	// select a random ghost node
	randomGhostNode := s.randomGhost()
	if randomGhostNode == nil {
		return ErrNoGhostNode
	}

	var requiredObjects []string
	s.Transactional(func() {
		requiredObjects = s.flatNS.Objects(failedNode)
	})

	// now we have required objects
	s.logger.Printf("failing node %s had objects %v", failedNode, requiredObjects)

	aggregate := make(map[string][]byte, len(requiredObjects))
	resultsCh := s.BroadcastDistributedRead(requiredObjects)
	for result := range resultsCh {
		typedResult := result.([]*api.NodeHeartBeat_Object)
		for _, ob := range typedResult {
			if _, ok := aggregate[ob.Name]; !ok {
				aggregate[ob.Name] = ob.Data
			}
		}
	}

	// now we have a randomly selected ghost node
	randomGhostNode.CommandCh <- aggregate

	s.Transactional(func() {
		for _, object := range requiredObjects {
			s.flatNS.AddNode(object, randomGhostNode.Id)
		}
	})

	return nil
}

func (s *DosNameNodeServer) Spawn(req *api.SpawnWait, stream grpc.ServerStreamingServer[api.SpawnCommand]) error {
	ghostNodeId := req.Name

	commandCh := make(chan map[string][]byte)
	closech := make(chan struct{})
	s.ghosts[ghostNodeId] = &GhostNodeEntry{
		Id:        ghostNodeId,
		CommandCh: commandCh,
		CloseCh:   closech,
	}

	defer func() {
		closech <- struct{}{}
		close(closech)
		s.logger.Printf("[ghostnode %s] stream closed", ghostNodeId)
	}()

	s.logger.Printf("[ghostnode %s] connected", ghostNodeId)
	for {
		select {
		case req, ok := <-commandCh:
			if !ok {
				s.logger.Printf("[ghostnode %s] closed command ch", ghostNodeId)
				return nil
			}

			commands := make([]*api.CreateCommand, len(req))
			for k, v := range req {
				commands = append(commands, &api.CreateCommand{
					ObjectName: k,
					ObjectData: v,
				})
			}

			stream.Send(&api.SpawnCommand{
				Status:   api.SpawnCommand_DONE,
				Lamport:  s.lamport,
				Commands: commands,
			})
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}
