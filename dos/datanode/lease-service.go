package datanode

/*

This file contains implementation for the lease service of the datanode
contais a ZMQ pub socket for publishing object updates as topics
@author Muhammad Rowaha<ashfaqrowaha@gmail.com>
@date 18/12/2024
*/

import (
	"fmt"
	"log"

	zmq "github.com/pebbe/zmq4"
)

type DataNodeLeaseService struct {
	publisher *zmq.Socket
}

func NewDataNodeLeaseService(binding string) *DataNodeLeaseService {
	publisher, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Fatalf("failed to create pub sock...\n%s", err.Error())
	}
	err = publisher.Bind(binding)
	if err != nil {
		log.Fatalf("failed to bind pub sock...\n%s", err.Error())
	}

	log.Printf("datanode leaser listening on addr %s\n", binding)
	return &DataNodeLeaseService{
		publisher,
	}
}

var (
	LEASE_DELETE = ":delete"
	LEASE_UPDATE = ":update"
	LEASE_CREATE = ":create"
)

func (leaser *DataNodeLeaseService) PushObjectCreate(objectName string, data []byte, sequence int) {
	topic := objectName
	stringified := string(data)
	message := fmt.Sprintf("%s%s %d %s", topic, LEASE_CREATE, sequence, stringified)
	leaser.publisher.Send(message, 0)
}

func (leaser *DataNodeLeaseService) PushObjectUpdate(objectName string, data []byte, sequence int) {
	topic := objectName
	stringified := string(data)
	message := fmt.Sprintf("%s%s %d %s", topic, LEASE_UPDATE, sequence, stringified)
	leaser.publisher.Send(message, 0)
}

func (leaser *DataNodeLeaseService) PushObjectDelete(objectName string, sequence int) {
	topic := objectName
	message := fmt.Sprintf("%s%s %d", topic, LEASE_DELETE, sequence)
	leaser.publisher.Send(message, 0)
}
