package main

import (
	"log"
	"strings"

	"04-unusual-database/server"
)

// version is replaced at compile time using -X flag.
var version = "dev"

func main() {
	server, err := server.NewUDPServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("running version", version)

	db := make(map[string]string)

	for packet := range server.Packets() {
		request := string(packet.Data)
		log.Println("request:", request)

		key, value, found := strings.Cut(request, "=")
		if found {
			// Insert request.
			db[key] = value
			continue
		}
		// Retrieve request.
		response := version
		if request != "version" {
			response = db[request]
		}
		packet.Reply([]byte(request + "=" + response))
	}
}
