package main

import (
	"cognologix.com/grpc-firewall-bypass/gnmi"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hashicorp/yamux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var m = make(map[string]*grpc.ClientConn)

func handleRequests() {
	log.Println("Starting REST API ...")
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/command/{serverAddress}/{commandString}", restMethod)
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
		//c := api.NewPingClient(conn)
		stringIPAddress := strings.Split(stringAddress, ":")[0]
		m[stringIPAddress] = conn
		// handle connection in goroutine so we can accept new TCP connections
		//go handleConn(conn, stringAddress, "Get")
	}
}

func restMethod(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	command := vars["commandString"]
	serverAddress := vars["serverAddress"]
	if client, ok := m[serverAddress]; ok {
		var output = handleConn(client, serverAddress, command)
		fmt.Fprintf(w, output)
	} else {
		// log error
		fmt.Fprintf(w, "client not available")
	}
}

func handleConn(conn *grpc.ClientConn, serverAddress string, command string ) string {
	//defer conn.Close()
	//c := api.NewPingClient(conn)
	c := gnmi.NewGNMIClient(conn)
	log.Println("input data " + serverAddress + " command " + command)
	response, err := c.Get(context.Background(), sampleRequest())
	if err != nil {
		log.Fatalf("error when calling gnmi Get: %s", err)
	}
	log.Printf("response from server address: %s , Command output is %s", serverAddress, response)
	return "From Device: \n\nServerAddress: " + serverAddress  +
		" \n\nCommand Sent to Device = " + command +
		" \n\nCommand Output = " + response.String()

}

func sampleRequest() *gnmi.GetRequest {
	path := make([]*gnmi.Path, 2)
	path[0], _ = ParseGNMIElements([]string{"a", "b", "c"})
	path[1], _ = ParseGNMIElements([]string{"x", "y", "z"})
	request := &gnmi.GetRequest{
		Prefix: nil,
		Path:   path,
	}
	return request
}

/** following methods are copied from: https://github.com/aristanetworks/goarista/blob/master/gnmi/path.go
ParseGNMIElements builds up a gnmi path, from user-supplied text*/
func ParseGNMIElements(elms []string) (*gnmi.Path, error) {
	var parsed []*gnmi.PathElem
	for _, e := range elms {
		n, keys, err := parseElement(e)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, &gnmi.PathElem{Name: n, Key: keys})
	}
	return &gnmi.Path{
		Element: elms, // Backwards compatibility with pre-v0.4 gnmi
		Elem:    parsed,
	}, nil
}

/** parseElement parses a path element, according to the gNMI specification. See
// https://github.com/openconfig/reference/blame/master/rpc/gnmi/gnmi-path-conventions.md
//
// It returns the first string (the current element name), and an optional map of key name
// value pairs.*/
func parseElement(pathElement string) (string, map[string]string, error) {
	// First check if there are any keys, i.e. do we have at least one '[' in the element
	name, keyStart := findUnescaped(pathElement, '[')
	if keyStart < 0 {
		return name, nil, nil
	}
	// Error if there is no element name or if the "[" is at the beginning of the path element
	if len(name) == 0 {
		return "", nil, fmt.Errorf("failed to find element name in %q", pathElement)
	}
	// Look at the keys now.
	keys := make(map[string]string)
	keyPart := pathElement[keyStart:]
	for keyPart != "" {
		k, v, nextKey, err := parseKey(keyPart)
		if err != nil {
			return "", nil, err
		}
		keys[k] = v
		keyPart = nextKey
	}
	return name, keys, nil
}

func parseKey(s string) (string, string, string, error) {
	if s[0] != '[' {
		return "", "", "", fmt.Errorf("failed to find opening '[' in %q", s)
	}
	k, iEq := findUnescaped(s[1:], '=')
	if iEq < 0 {
		return "", "", "", fmt.Errorf("failed to find '=' in %q", s)
	}
	if k == "" {
		return "", "", "", fmt.Errorf("failed to find key name in %q", s)
	}
	rhs := s[1+iEq+1:]
	v, iClosBr := findUnescaped(rhs, ']')
	if iClosBr < 0 {
		return "", "", "", fmt.Errorf("failed to find ']' in %q", s)
	}
	if v == "" {
		return "", "", "", fmt.Errorf("failed to find key value in %q", s)
	}
	next := rhs[iClosBr+1:]
	return k, v, next, nil
}

/**findUnescaped will return the index of the first unescaped match of 'find', and the unescaped
// string leading up to it.*/
func findUnescaped(s string, find byte) (string, int) {
	// Take a fast track if there are no escape sequences
	if strings.IndexByte(s, '\\') == -1 {
		i := strings.IndexByte(s, find)
		if i < 0 {
			return s, -1
		}
		return s[:i], i
	}
	// Find the first match, taking care of escaped chars.
	var b strings.Builder
	var i int
	len := len(s)
	for i = 0; i < len; {
		ch := s[i]
		if ch == find {
			return b.String(), i
		} else if ch == '\\' && i < len-1 {
			i++
			ch = s[i]
		}
		b.WriteByte(ch)
		i++
	}
	return b.String(), -1
}

// TCP server and GRPC client
func main() {

	fmt.Println("Rest API v2.0 - Mux Routers")
	go initGrpcServer()

	handleRequests()

}

