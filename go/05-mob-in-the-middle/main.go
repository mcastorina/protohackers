package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	for conn := range server.Connections() {
		server.Handle(conn, func(conn net.Conn) {
			scanner := bufio.NewScanner(conn)

			for scanner.Scan() {
				msg := scanner.Text()
				fmt.Printf("Incoming message %s\n", msg)

				// read message and edit accordingly

				// forward to chat.protohackers.com on port 16963

				// read back message

				// write message to connection
			}
		})
	}
	server.Wait()
}

func extractBoguscoin(msg string) (string, bool) {
	return "", false
}
