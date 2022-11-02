package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

// Deep inside Initrode Global's enterprise management framework lies a
// component that writes data to a server and expects to read the same data
// back. (Think of it as a kind of distributed system delay-line memory). We
// need you to write the server to echo the data back.

// Accept TCP connections.

// Whenever you receive data from a client, send it back unmodified.

// Make sure you don't mangle binary data, and that you can handle at least 5
// simultaneous clients.

// Once the client has finished sending data to you it shuts down its sending
// side. Once you've reached end-of-file on your receiving side, and sent back
// all the data you've received, close the socket so that the client knows
// you've finished. (This point trips up a lot of proxy software, such as
// ngrok; if you're using a proxy and you can't work out why you're failing the
// check, try hosting your server in the cloud instead).

// Your program will implement the TCP Echo Service from RFC 862.

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", ":1337")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening on :1337")

	go func() {
		<-ctx.Done()
		log.Println("shutting down")
		listener.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		// Check if the listener closed.
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Println("accepted new connection")
		wg.Add(1)
		go func() {
			defer wg.Done()
			echo(ctx, conn)
			log.Println("closing connection")
			if err := conn.Close(); err != nil {
				log.Println(err)
			}
		}()
	}
	wg.Wait()
}
