package main

import (
	"log"
	"net"

	"06-speed-daemon/client"
)

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	for conn := range server.Connections() {
		server.Handle(conn, func(conn net.Conn) {
			c := client.New(conn)
			println(c)
		})
	}
	server.Wait()
}
