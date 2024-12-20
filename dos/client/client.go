package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	api "github.com/mrowaha/dos/api"
	"github.com/mrowaha/dos/datanode"
	zmq "github.com/pebbe/zmq4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DosClient struct {
	client api.NameServiceClient
	logger *log.Logger
}

func NewDosClient(conn *grpc.ClientConn) *DosClient {
	c := api.NewNameServiceClient(conn)
	logger := log.New(os.Stdout, "[client]", log.Ltime)
	return &DosClient{
		client: c,
		logger: logger,
	}
}

func (c *DosClient) Create(name string, data []byte) error {
	c.logger.Printf("creating object named %s", name)

	ctx := context.Background()
	_, err := c.client.CreateObject(ctx, &api.CreateObjectRequest{
		Meta: &api.RequestMeta{Ts: timestamppb.Now()},
		Name: name,
		Data: data,
	})
	if err != nil {
		c.logger.Printf("failed to create object...\n %s", err.Error())
	}
	return nil
}

func (c *DosClient) Delete(name string) error {
	c.logger.Printf("deleting object named %s", name)
	_, err := c.client.DeleteObject(context.TODO(), &api.DeleteObjectRequest{
		Name: name,
	})
	if err != nil {
		c.logger.Printf("failed to delete object...\n%s", err.Error())
	}
	return nil
}

func (c *DosClient) Update(name string, data []byte) error {
	c.logger.Printf("updating object named %s", name)
	_, err := c.client.UpdateObject(context.TODO(), &api.UpdateObjectReq{
		Name: name,
		Data: data,
	})
	if err != nil {
		c.logger.Printf("failed to update object...\n%s", err.Error())
	}
	return nil
}

// this function subscribes to an object name from a randomly selected datanode
func (c *DosClient) Lease(name string) error {
	c.logger.Printf("leasing object named %s", name)
	res, err := c.client.LeaseObject(context.TODO(), &api.LeaseObjectReq{
		Name: name,
	})
	if err != nil {
		c.logger.Printf("failed to lease object...\n%s\n", err.Error())
	}

	c.logger.Printf("lease options: %v\n", res)

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(res.Leasers))
	c.subscribe(name, res.Leasers[randomIndex])
	return nil
}

func (c *DosClient) subscribe(object string, leaseAddr string) error {
	sub, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		panic(err)
	}
	defer sub.Close()

	fmt.Println(leaseAddr)
	err = sub.Connect(leaseAddr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// topics for update and delete
	updateTopic := fmt.Sprintf("%s%s", object, datanode.LEASE_UPDATE)
	deleteTopic := fmt.Sprintf("%s%s", object, datanode.LEASE_DELETE)

	err = sub.SetSubscribe(updateTopic)
	if err != nil {
		log.Fatalf("failed to sub topic %s...\n%s\n", updateTopic, err.Error())
	}

	err = sub.SetSubscribe(deleteTopic)
	if err != nil {
		log.Fatalf("failed to sub topic %s...\n%s\n", deleteTopic, err.Error())
	}

	c.logger.Printf("listening to %s and %s", updateTopic, deleteTopic)

	for {
		msg, err := sub.RecvMessage(0)
		if err != nil {
			c.logger.Fatalf("failed to recv from leaser...\n%s\n", err.Error())
		}
		c.logger.Printf("recv message %s\n", msg)
	}

}
