package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"07-line-reversal/server/lrcp"
)

// version is replaced at compile time using -X flag.
var version = "dev"

func main() {
	server, err := lrcp.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	_, debug := os.LookupEnv("DEBUG")
	log.Println("running version", version, "( debug =", debug, ")")

	for conn := range server.Connections() {
		server.Handle(conn, func(conn *lrcp.Conn) {
			proxy(bufio.NewReader(conn), bufio.NewWriter(conn), func(in string) string {
				reversed := make([]rune, len(in))
				for i, r := range strings.TrimSuffix(in, "\n") {
					reversed[len(reversed)-2-i] = r
				}
				reversed[len(reversed)-1] = '\n'
				return string(reversed)
			})
		})
	}
}

// proxy will read from reader, transform the message using mapper, then write
// to writer. This function runs until either read or write operation fails.
func proxy(reader *bufio.Reader, writer *bufio.Writer, mapper func(string) string) {
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		msg = mapper(msg)
		if _, err = writer.WriteString(msg); err != nil {
			return
		}
		if err = writer.Flush(); err != nil {
			return
		}
	}
}
