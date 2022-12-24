package main

import (
	"log"
	"strings"
	"sync"

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

	db := Database{data: make(map[string]string)}

	for packet := range server.Packets() {
		request := string(packet.Data)
		log.Println("request:", request)

		key, value, found := strings.Cut(request, "=")
		if found {
			// Insert request.
			db.Insert(key, value)
			continue
		}
		// Retrieve request.
		response := version
		if request != "version" {
			response = db.Retrieve(request)
		}
		packet.Reply([]byte(request + "=" + response))
	}
}

type Database struct {
	data map[string]string
	lock sync.RWMutex
}

func (db *Database) Insert(key, value string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.data[key] = value
}

func (db *Database) Retrieve(key string) string {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.data[key]
}
