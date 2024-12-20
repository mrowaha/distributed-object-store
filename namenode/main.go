package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	dos "github.com/mrowaha/dos/namenode"
)

var (
	port      int
	logfile   string
	nsFile    string
	repl      int
	tolerance int
)

func main() {
	flag.IntVar(&port, "port", 50051, "port number for name service")
	flag.StringVar(&logfile, "logfile", "namenode.log", "log file path")
	flag.StringVar(&nsFile, "nsfile", "namenode-ns.txt", "flat namespace file pth")
	flag.IntVar(&repl, "repl", 2, "replication factor")
	flag.IntVar(&tolerance, "tol", 1, "tolerance factor")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("listening on port %d", port)
	service, err := dos.NewDosNameNodeServer(logfile, nsFile, dos.WithReplication(repl), dos.WithTolerance(tolerance))
	if err != nil {
		log.Fatalln(err.Error())
	}
	service.Serve(&listener)
}
