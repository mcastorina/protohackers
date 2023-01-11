package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"07-line-reversal/server"
)

// version is replaced at compile time using -X flag.
var version = "dev"

type Connection struct {
	addr net.Addr
}

func main() {
	sessionToConnection := make(map[int32]*Connection)

	server, err := server.NewUDPServer()
	if err != nil {
		log.Fatal(err)
	}
	_, debug := os.LookupEnv("DEBUG")
	log.Println("running version", version, "( debug =", debug, ")")

	for packet := range server.Packets() {
		request := string(packet.Data)
		if debug {
			request = strings.TrimSpace(request)
		}
		log.Println("request:", request)

		parts, err := parseLRCP(request)
		if err != nil {
			continue
		}
		fmt.Println(parts)

		if parts[0] == "connect" {
			// get session id from request
			sessionId, _ := strconv.Atoi(parts[1])

			// get or create connection
			_, ok := sessionToConnection[int32(sessionId)]
			if !ok {
				fmt.Println("not okay!")
				sessionToConnection[int32(sessionId)] = &Connection{
					addr: packet.Addr,
				}
			} else {
				fmt.Println("not not okay!")
			}
		}
	}
}

// parseLRCP parses a LRCP packet into its parts
func parseLRCP(data string) ([]string, error) {
	if !strings.HasPrefix(data, "/") || !strings.HasSuffix(data, "/") {
		return nil, errors.New("invalid packet format")
	}
	data = strings.TrimPrefix(strings.TrimSuffix(data, "/"), "/")
	parts := strings.SplitN(data, "/", 4)
	return parts, nil
}
