package namenode

import (
	"errors"
	"sync/atomic"

	"google.golang.org/grpc"

	"github.com/mrowaha/dos/api"
)

var (
	ErrInvalidDataNodeUUID = errors.New("failed to parse data node to uuid")
)

type BroadcastEvent string

// these events are broadcasted to datanodes
const (
	COMMIT          BroadcastEvent = "commit"
	DELETE          BroadcastEvent = "delete"
	UPDATE          BroadcastEvent = "update"
	DISTRIBUTEDREAD BroadcastEvent = "distributed-read"
)

type CreateCommand struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

type CommitCommand struct {
	Lamport int32          `json:"-"`
	Type    BroadcastEvent `json:"type"`
	Name    string         `json:"name"`
}

type DeleteCommand struct {
	Lamport int32          `json:"-"`
	Name    string         `json:"name"`
	Type    BroadcastEvent `json:"type"`
}

type UpdateCommand struct {
	Lamport int32          `json:"-"`
	Name    string         `json:"name"`
	Data    []byte         `json:"data"`
	Type    BroadcastEvent `json:"type"`
}

type DistributedReadCommand struct {
	Lamport int32          `json:"-"`
	Objects []string       `json:"objects"`
	Type    BroadcastEvent `json:"type"`
}

type CommandNode struct {
	tag             string
	command         api.CommandNodeRes_Command
	commit          CommitCommand
	create          CreateCommand
	delete          DeleteCommand
	update          UpdateCommand
	distributedRead DistributedReadCommand
}

// this file contains definitions for the datanode service
func (s *DosNameNodeServer) RegisterNode(stream grpc.BidiStreamingServer[api.NodeHeartBeat, api.CommandNodeRes]) error {
	var dataNodeID string
	// block for first heartbeat
	req, _ := stream.Recv()
	dataNodeID = req.Id
	s.logger.Printf("registering data node %s", dataNodeID)

	reqChan := make(chan CommandNode)
	resChans := make(map[string]chan interface{})
	closedCh := make(chan struct{}, 1) // size of the channel is one
	// because we do not want to block when the stream is closed

	s.Transactional(func() {
		s.meta.RegisterNode(&MetaHeapEntry{
			Id:        dataNodeID,
			CommandCh: reqChan,
			ResChs:    resChans,
			ClosedCh:  closedCh,
			Size:      req.Size,
			Lease:     req.LeaserService,
		})
	})

	defer func() {
		closedCh <- struct{}{}
		for tag, ch := range resChans {
			s.logger.Printf("closing ch for tag %s\n", tag)
			close(ch)
		}

		s.Transactional(func() {
			s.meta.DeleteNode(dataNodeID)
			s.logger.Printf("[datanode %s] removed from heap", dataNodeID)
		})

		// begin protocol for fault tolerance via ghost nodes
		err := s.InitiateSpawn(dataNodeID)
		if err != nil {
			if err == ErrNoGhostNode {
				// s.logger.Fatalf("likely bug. execution should not reach")
			}
		}

		s.Transactional(func() {
			s.flatNS.RemoveNode(dataNodeID)
		})

	}()

	go func() {
		// heart beat
		for {
			select {
			case <-closedCh:
				return
			default:
				req, err := stream.Recv()
				if err != nil {
					return
				}

				if req.Type == api.NodeHeartBeat_ACK {
					if ch, ok := resChans[req.MessageTag]; ok {
						s.logger.Printf("ack message tagged %s", req.MessageTag)
						ch <- true
					} else {
						s.logger.Printf("ack error, message tag %s channel does not exist", req.MessageTag)
					}
				} else if req.Type == api.NodeHeartBeat_BEAT {
					s.Transactional(func() {
						s.meta.UpdateSize(req.Id, req.Size)
					})
				} else if req.Type == api.NodeHeartBeat_DISTRIUTED_READ {
					if ch, ok := resChans[req.MessageTag]; ok {
						s.logger.Printf("distributed read tagged %s result", req.MessageTag)
						ch <- req.ObjectData
					} else {
						s.logger.Printf("distributed read error, message tag %s channel does not exist", req.MessageTag)
					}
				}
			}
		}
	}()

	mux := NewDataNodeCommandMux(stream, s.logger)

	for {
		select {
		case req, ok := <-reqChan:
			if !ok {
				s.logger.Printf("[datanode %s] closed request ch", dataNodeID)
				return nil
			}
			go mux.Handle(&req)
		case <-stream.Context().Done():
			s.logger.Printf("[datanode %s] closed connection", dataNodeID)
			return stream.Context().Err()
		}
	}
}

func (s *DosNameNodeServer) BroadcastCommit(name string) {
	atomic.AddInt32(&s.lamport, 1)
	command := CommandNode{
		command: api.CommandNodeRes_COMMIT,
		commit:  CommitCommand{Lamport: s.lamport, Type: COMMIT, Name: name},
	}
	s.meta.ForEach(func(entry *MetaHeapEntry) {
		messageTag := CommitMessageTag(name, entry.Id)
		command.tag = messageTag
		ackChan := make(chan interface{})
		entry.ResChs[messageTag] = ackChan
		entry.CommandCh <- command
		success, ok := <-ackChan
		if ok {
			if !success.(bool) {
				s.logger.Printf("failed to broadcast commit to %s\n", entry.Id)
			} else {
				s.logger.Printf("broadcasted commit to %s\n", entry.Id)
			}
		}
		delete(entry.ResChs, messageTag)
	})
}

func (s *DosNameNodeServer) BroadcastDelete(name string) {
	atomic.AddInt32(&s.lamport, 1)
	command := CommandNode{
		command: api.CommandNodeRes_DELETE,
		delete: DeleteCommand{
			Lamport: s.lamport,
			Type:    DELETE,
			Name:    name,
		},
	}
	s.meta.ForEach(func(entry *MetaHeapEntry) {
		messageTag := DeleteMessageTag(name, entry.Id)
		command.tag = messageTag
		ackChan := make(chan interface{})
		entry.ResChs[messageTag] = ackChan
		entry.CommandCh <- command
		success, ok := <-ackChan
		if ok {
			if !success.(bool) {
				s.logger.Printf("failed to broadcast delete %s\n", entry.Id)
			} else {
				s.logger.Printf("broadcasted delete to %s\n", entry.Id)
			}
		}
		delete(entry.ResChs, messageTag)
	})
}

func (s *DosNameNodeServer) BroadcastUpdate(name string, data []byte) {
	atomic.AddInt32(&s.lamport, 1)
	command := CommandNode{
		command: api.CommandNodeRes_UPDATE,
		update: UpdateCommand{
			Lamport: s.lamport,
			Type:    UPDATE,
			Name:    name,
			Data:    data,
		},
	}
	s.meta.ForEach(func(entry *MetaHeapEntry) {
		messageTag := UpdateMessageTag(name, entry.Id)
		command.tag = messageTag
		ackChan := make(chan interface{})
		entry.ResChs[messageTag] = ackChan
		entry.CommandCh <- command
		success, ok := <-ackChan
		if ok {
			if !success.(bool) {
				s.logger.Printf("failed to broadcast update %s\n", entry.Id)
			} else {
				s.logger.Printf("broadcasted update to %s\n", entry.Id)
			}
		}
		delete(entry.ResChs, messageTag)
	})
}

func (s *DosNameNodeServer) BroadcastDistributedRead(objects []string) <-chan interface{} {
	atomic.AddInt32(&s.lamport, 1)
	command := CommandNode{
		command: api.CommandNodeRes_DISTRIBUTED_READ,
		distributedRead: DistributedReadCommand{
			Objects: objects,
			Lamport: s.lamport,
		},
	}

	resultsCh := make(chan interface{}, 5)
	defer close(resultsCh)
	s.meta.ForEach(func(entry *MetaHeapEntry) {
		messageTag := DistributedReadTag(objects, entry.Id)
		command.tag = messageTag
		ackChan := make(chan interface{})
		entry.ResChs[messageTag] = ackChan
		entry.CommandCh <- command
		reads, ok := <-ackChan
		if ok {
			resultsCh <- reads
		}
		delete(entry.ResChs, messageTag)
	})

	return resultsCh
}
