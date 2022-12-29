package main

import (
	"fmt"
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
			if camera, err := c.AsCamera(); err == nil {
				fmt.Println("camera", camera)
				for {
					if plateInfo, err := camera.ReadPlate(); err == nil {
						// TODO play with plateinfo logic
						fmt.Println(plateInfo)
					} else {
						// Return an error and disconnect.
						_ = camera.SendError(err)
						return
					}
				}
			} else if dispatcher, err := c.AsDispatcher(); err == nil {
				fmt.Println("dispatcher", dispatcher)
			} else {
				// Unrecognized device.
				return
			}
			fmt.Println(c)
		})
	}
	server.Wait()
}
