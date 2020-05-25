package main

import (
	"cognologix.com/grpc-firewall-bypass/gnmi"
	"log"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
)

// TCP client and GRPC server
func main() {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8081", time.Second*5)
	if err != nil {
		log.Fatalf("error dialing: %s", err)
	}

	srvConn, err := yamux.Server(conn, yamux.DefaultConfig())
	if err != nil {
		log.Fatalf("couldn't create yamux server: %s", err)
	}

	// create a server instance
	s := gnmi.Server{}

	// create a gRPC server object
	grpcServer := grpc.NewServer()

	// attach the Ping service to the server
	gnmi.RegisterGNMIServer(grpcServer, &s)

	// start the gRPC erver
	log.Println("launching gRPC server over TCP connection...")
	if err := grpcServer.Serve(srvConn); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
