package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

const upstreamBudgetChatServer = "chat.protohackers.com:16963"
const tonyBoguscoinAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	for conn := range server.Connections() {
		// Create upstream connection.
		upstream, err := net.Dial("tcp", upstreamBudgetChatServer)
		if err != nil {
			log.Println(err)
			continue
		}

		conn := conn
		// Asynchronously read from client and send to upstream.
		go func() {
			defer upstream.Close()
			clientReader := bufio.NewReader(conn)
			upstreamWriter := bufio.NewWriter(upstream)

			proxy(clientReader, upstreamWriter, replaceBoguscoins)
			log.Println("user disconnected")
		}()
		// Asynchronously read from upstream and send to client.
		go func() {
			defer conn.Close()
			upstreamReader := bufio.NewReader(upstream)
			clientWriter := bufio.NewWriter(conn)

			proxy(upstreamReader, clientWriter, replaceBoguscoins)
			log.Println("upstream disconnected")
		}()
	}
	server.Wait()
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

// replaceBoguscoins replaces all boguscoin addresses found in msg with Tony's
// boguscoin address.
func replaceBoguscoins(msg string) string {
	msg = strings.ReplaceAll(msg, " ", " 🧑‍🎄 ")
	var re = regexp.MustCompile(`(^|\s)(7[a-zA-Z0-9]{25,34})(\s|$)`)
	addressReplacedMsg := re.ReplaceAllString(msg, fmt.Sprintf("${1}%s${3}", tonyBoguscoinAddress))
	return strings.ReplaceAll(addressReplacedMsg, " 🧑‍🎄 ", " ")
}
