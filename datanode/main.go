package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	dos "github.com/mrowaha/dos/datanode"
)

var (
	port    int
	store   string
	process string
	lease   string
	lamport int
)

func main() {
	flag.IntVar(&port, "port", 50051, "port of name node service")
	flag.StringVar(&store, "store", "data.db", "store .db file name")
	flag.StringVar(&process, "process", "", "name of the process. required for unique queues in redis")
	flag.StringVar(&lease, "lease", "", "lease address")
	flag.IntVar(&lamport, "lamport", 0, "initial lamport")
	flag.Parse()

	if len(process) == 0 {
		log.Fatalf("required flag -process")
	}

	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if process == "alpha" {
		go func() {
			time.Sleep(10 * time.Second)
			os.Exit(1)
		}()
	}

	client := dos.NewDosDataNode(
		conn,
		NewRedisDataNodeQueue(),
		dos.WithDBFile(store),
		dos.WithLeaser(lease),
		dos.WithName(process),
		dos.WithLamport(lamport),
	)
	client.Register()
}
