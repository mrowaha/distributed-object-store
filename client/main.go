package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	dos "github.com/mrowaha/dos/client"
)

var (
	port   int
	object string
	data   string
	cmd    int
)

func main() {
	flag.IntVar(&port, "port", 50051, "port of name node service")
	flag.StringVar(&object, "object", "test", "name of object")
	flag.StringVar(&data, "data", "test data", "data of object")
	flag.IntVar(&cmd, "cmd", 1, "Create = 1, Delete = 2, Update = 3, Lease = 4")
	flag.Parse()

	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := dos.NewDosClient(conn)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Add(1)
	go func() {
		client.Create(object, []byte(data))
		wg.Done()
	}()

	time.Sleep(2 * time.Second)
	go func() {
		client.Delete(object)
		wg.Done()
	}()

	wg.Wait()

	// if cmd == 1 {
	// 	client.Create(object, []byte(data))
	// } else if cmd == 2 {
	// 	client.Delete(object)
	// } else if cmd == 3 {
	// 	client.Update(object, []byte(data))
	// } else if cmd == 4 {
	// 	client.Lease(object)
	// }
}
