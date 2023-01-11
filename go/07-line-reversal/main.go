package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"07-line-reversal/server"
)

// version is replaced at compile time using -X flag.
var version = "dev"

type Connection struct {
	addr net.Addr
}

func main() {
	sessionToConnection := make(map[string]*Connection)

	server, err := server.NewUDPServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("running version", version)

	for packet := range server.Packets() {
		request := string(packet.Data)
		log.Println("request:", request)

		parts := strings.SplitN(request, "/", 3)
		fmt.Println(parts)

		if len(parts) == 1 {
			panic("invalid request")
		}

		if parts[1] == "connect" {
			// get session id from request
			sessionId := parts[2] // TODO take off "/"

			// get or create connection
			// TODO use lock
			_, ok := sessionToConnection[sessionId]
			if !ok {
				fmt.Println("not okay!")
				sessionToConnection[sessionId] = &Connection{
					addr: packet.Addr,
				}
			} else {
				fmt.Println("not not okay!")
			}
		}
	}
}
