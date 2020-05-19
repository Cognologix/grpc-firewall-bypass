package api

import (
	"log"

	"golang.org/x/net/context"
)

// Server represents the gRPC server
type Server struct {
}

// SayHello generates response to a Ping request
func (s *Server) SayHello(ctx context.Context, in *PingMessage) (*PingMessage, error) {
	log.Printf("Receive message %s", in.Greeting)
	return &PingMessage{Greeting: "bar"}, nil
}

// SayHello generates response to a Ping request
func (s *Server) RunCommand(ctx context.Context, in *CommandMessage) (*CommandResponse, error) {
	log.Printf("Receive message %s", in.Command)
	return &CommandResponse{CommandResult: "Command Executed Successfully"}, nil
}
