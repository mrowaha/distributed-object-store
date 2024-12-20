package namenode

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"sync"

	"github.com/mrowaha/dos/api"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrNotEnoughDataNodes      = errors.New("not enough nodes connected")
	ErrFailedObjectReplication = errors.New("failed to replicate object")
)

type DosNameNodeServer struct {
	api.UnimplementedNameServiceServer
	api.UnimplementedDataServiceServer
	api.UnimplementedGhostServiceServer
	logger  *log.Logger
	flatNS  *FlatNamespace
	config  *NameNodeConfig
	meta    *DataNodeMeta
	lock    sync.Mutex
	ghosts  GhostNodesMap
	lamport int32
}

func NewDosNameNodeServer(logFilePath string, flatNSPath string, opts ...ConfigFunc) (*DosNameNodeServer, error) {
	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	logger := log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	flatNS := NewFlatNamespace()
	err = flatNS.Load(flatNSPath)
	if err != nil {
		return nil, err
	}

	config := NewNameNodeConfig(opts...)

	meta := NewDataNodeMeta()

	return &DosNameNodeServer{
		logger:  logger,
		flatNS:  flatNS,
		config:  config,
		meta:    meta,
		lamport: 0,
		ghosts:  make(GhostNodesMap),
	}, nil
}

type Transaction func()

func (s *DosNameNodeServer) Transactional(fn Transaction) {
	s.lock.Lock()
	defer s.lock.Unlock()
	fn()
}

// this function will begin the name service itself and begin listening to clients
// it accepts a listener object so that the grpc server can connect to it
// this is a blocking procedure
func (s *DosNameNodeServer) Serve(listener *net.Listener) {
	grpcServer := grpc.NewServer()
	api.RegisterNameServiceServer(grpcServer, s)
	api.RegisterDataServiceServer(grpcServer, s)
	api.RegisterGhostServiceServer(grpcServer, s)
	if err := grpcServer.Serve(*listener); err != nil {
		log.Fatalf("failed to start name node service %v", err)
	}
}

/*
*
name node create object service will place a lock on the flat namespace
since accesses to the flat namespace are single threaded
*/
func (s *DosNameNodeServer) CreateObject(ctx context.Context, req *api.CreateObjectRequest) (*api.CreateObjectResponse, error) {
	s.logger.Printf("request to create object %s\n", req.Name)

	// if len(s.ghosts) < s.config.Tolerance && s.config.Tolerance != -1 {
	// 	return nil, ErrToleranceNotEnough
	// }

	var err error

	s.Transactional(func() {
		if s.flatNS.Exists(req.Name) {
			err = ErrObjectAlreadyExists
			return
		}
		successes := 0
		pickedEntries := make([]*MetaHeapEntry, 0)
		// for this to be successfull the number of successess should be equal to replication
		for i := 1; i <= s.config.Replication; i++ {
			if s.meta.Count() == 0 {
				err = ErrNotEnoughDataNodes
				return
			}

			nextMetaEntry := s.meta.Next()
			ackChan := make(chan interface{})
			messageTag := CreateMessageTag(req.Name, nextMetaEntry.Id)
			nextMetaEntry.ResChs[messageTag] = ackChan
			log.Printf("selected %s", nextMetaEntry.Id)
			select {
			case nextMetaEntry.CommandCh <- CommandNode{command: api.CommandNodeRes_CREATE, tag: messageTag,
				create: CreateCommand{
					Name: req.Name,
					Data: req.Data,
				}}:
				s.logger.Printf("create message tagged with %s\n", messageTag)
				ack, ok := <-ackChan
				if ok {
					// channel was not closed / cleanedup
					if ack.(bool) {
						s.logger.Printf("object %s replicated to %s\n", req.Name, nextMetaEntry.Id)
						successes++
					}
				}
				pickedEntries = append(pickedEntries, nextMetaEntry)

			case <-nextMetaEntry.ClosedCh:
				close(nextMetaEntry.CommandCh)
			}

			delete(nextMetaEntry.ResChs, messageTag)
		}
		for _, entry := range pickedEntries {
			// re-register these nodes
			s.meta.RegisterNode(entry)
		}

		if successes != s.config.Replication {
			err = ErrFailedObjectReplication
			return
		}

		s.flatNS.AddObject(req.Name)
		s.logger.Printf("added object [%s] to flatNS with OPEN\n", req.Name)
		for _, entry := range pickedEntries {
			s.flatNS.AddNode(req.Name, entry.Id)
		}
		s.logger.Printf("flatNS entries updated: %v\b", s.flatNS)

	})
	// check error after meta transactional
	if err != nil {
		return nil, err
	}
	s.BroadcastCommit(req.Name)
	return &api.CreateObjectResponse{
		Meta: &api.ResponseMeta{Ts: timestamppb.Now(), Status: api.ResponseMeta_CREATED},
	}, nil
}

/*
*
name node deletes object with the given name and initiates a deletion sequence
*/
func (s *DosNameNodeServer) DeleteObject(ctx context.Context, req *api.DeleteObjectRequest) (*api.DeleteObjectResponse, error) {
	s.logger.Printf("request to delete object %s", req.Name)

	var transactionErr error
	s.Transactional(func() {

		if !s.flatNS.Exists(req.Name) {
			transactionErr = ErrObjectDoestNotExist
			s.logger.Printf("object %s does not exist in ns %v\n", req.Name, s.flatNS.ns)
			return
		}

		if err := s.flatNS.DeleteObject(req.Name); err != nil {
			transactionErr = err
		}
	})

	if transactionErr != nil {
		return nil, transactionErr
	}
	s.BroadcastDelete(req.Name)
	s.logger.Printf("deleted object [%s] from flatNS", req.Name)
	return &api.DeleteObjectResponse{
		Meta: &api.ResponseMeta{Status: api.ResponseMeta_DELETED},
	}, transactionErr
}

func (s *DosNameNodeServer) UpdateObject(ctx context.Context, req *api.UpdateObjectReq) (*api.UpdateObjectRes, error) {
	s.logger.Printf("attempting request [update %s]\n", req.Name)
	var transactionErr error
	s.Transactional(func() {
		if !s.flatNS.Exists(req.Name) {
			transactionErr = ErrObjectDoestNotExist
			s.logger.Printf("object %s does not exist in ns %v\n", req.Name, s.flatNS.ns)
			return
		}
	})

	if transactionErr != nil {
		return nil, transactionErr
	}

	s.BroadcastUpdate(req.Name, req.Data)
	s.logger.Printf("updated object [%s]", req.Name)
	return &api.UpdateObjectRes{
		Meta: &api.ResponseMeta{
			Status: api.ResponseMeta_UPDATED,
		},
	}, nil
}

func (s *DosNameNodeServer) LeaseObject(ctx context.Context, req *api.LeaseObjectReq) (*api.LeaseObjectRes, error) {
	s.logger.Printf("attempting request [lease %s]\n", req.Name)

	var transactionErr error
	var res *api.LeaseObjectRes
	s.Transactional(func() {
		nodes, err := s.flatNS.Nodes(req.Name)
		if err != nil {
			transactionErr = err
			s.logger.Printf("failed to lease...\n%s\n", err.Error())
			return
		}

		// we have the nodes here. now we need to map to lease addresses
		leaseServices, _ := s.meta.LeaseServices(nodes)
		res = &api.LeaseObjectRes{
			Leasers: leaseServices,
		}
	})

	if transactionErr != nil {
		return nil, transactionErr
	}

	return res, nil
}
