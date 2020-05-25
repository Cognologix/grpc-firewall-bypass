package main

import (
	"cognologix.com/grpc-firewall-bypass/api"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/yamux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"time"
)

var m = make(map[string]api.PingClient)

func handleRequests() {
	log.Println("Starting REST API ...")
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/command/{serverAddress}/{commandString}", runCommand)
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func initGrpcServer() {
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

		var address net.Addr
		address = incomingConn.RemoteAddr()
		var stringAddress = address.String()
		log.Println("Remote Address :: " + stringAddress)
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

		//defer conn.Close()
		c := api.NewPingClient(conn)
		m[stringAddress] = c
		// handle connection in goroutine so we can accept new TCP connections
		//go handleConn(conn, stringAddress, "Get")
	}
}

func runCommand(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	command := vars["commandString"]
	serverAddress := vars["serverAddress"]
	var output = handleConn(m[serverAddress], serverAddress, command)
	fmt.Fprintf(w, output)
}

func handleConn(c api.PingClient, serverAddress string, command string ) string {
	//defer conn.Close()
	//c := api.NewPingClient(conn)

	response, err := c.RunCommand(context.Background(), &api.CommandMessage{Command: command})
	if err != nil {
		log.Fatalf("error when calling RunCommand: %s", err)
	}
	log.Printf("response from server address: %s , Command output is %s", serverAddress, response.CommandResult)
	return "From Device: \n\nServerAddress: " + serverAddress  +
		" \n\nCommand Sent to Device = " + command +
		" \n\nCommand Output = " + response.CommandResult




}

// TCP server and GRPC client
func main() {

	fmt.Println("Rest API v2.0 - Mux Routers")
	go initGrpcServer()

	handleRequests()

}

