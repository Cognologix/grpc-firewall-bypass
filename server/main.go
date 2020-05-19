package main

import (
	"context"
	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
	"cognologix.com/grpc-firewall-bypass/api"
	"log"
	"net"
	"time"
)

// TCP server and GRPC client

func main() {

	log.Println("launching tcp server...")

	// start tcp listener on all interfaces
	// note that each connection consumes a file descriptor
	// you may need to increase your fd limits if you have many concurrent clients
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}
	defer ln.Close()

	for {
		log.Println("waiting for incoming TCP connections...")
		// Accept blocks until there is an incoming TCP connection
		incoming, err := ln.Accept()
		if err != nil {
			log.Fatalf("couldn't accept %s", err)
		}

		incomingConn, err := yamux.Client(incoming, yamux.DefaultConfig())
		if err != nil {
			log.Fatalf("couldn't create yamux %s", err)
		}

		log.Println("starting a gRPC server over incoming TCP connection")

		var conn *grpc.ClientConn
		// gRPC dial over incoming net.Conn
		conn, err = grpc.Dial(":7777", grpc.WithInsecure(),
			grpc.WithDialer(func(target string, timeout time.Duration) (net.Conn, error) {
				return incomingConn.Open()
			}),
		)

		if err != nil {
			log.Fatalf("did not connect: %s", err)
		}

		// handle connection in goroutine so we can accept new TCP connections
		go handleConn(conn)
	}
}

func handleConn(conn *grpc.ClientConn) {
	defer conn.Close()
	c := api.NewPingClient(conn)
	for i := 0; i < 10; i++ {
		response, err := c.RunCommand(context.Background(), &api.CommandMessage{Command: "setConf"})
		if err != nil {
			log.Fatalf("error when calling RunCommand: %s", err)
		}
		log.Printf("response from server: %s", response.CommandResult)
	}

}
