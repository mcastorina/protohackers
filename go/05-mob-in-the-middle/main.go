package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"regexp"
	"sync"
)

const upstreamBudgetChatServer = "chat.protohackers.com:16963"
const tonyBoguscoinAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	for conn := range server.Connections() {
		server.Handle(conn, func(conn net.Conn) {
			// Create upstream connection.
			upstream, err := net.Dial("tcp", upstreamBudgetChatServer)
			if err != nil {
				log.Println(err)
				return
			}
			defer upstream.Close()

			var wg sync.WaitGroup
			ctx, cancel := context.WithCancel(context.Background())

			wg.Add(2)
			go func() {
				defer wg.Done()
				scanner := bufio.NewScanner(conn)
				for scanner.Scan() {
					// Read message from user.
					msg := scanner.Text()

					// Filter/map user message.
					updatedMsg := updateMessage(msg)

					// Write to upstream.
					upstream.Write([]byte(updatedMsg + "\n"))
				}
				log.Println("user disconnected")
				cancel()
			}()
			go func() {
				defer wg.Done()
				upstreamReader := bufio.NewReader(upstream)
				for {
					// Try read message from upstream.
					msg, ok := readLine(ctx, upstreamReader)
					if !ok {
						break
					}

					// Filter/map user message.
					updatedMsg := updateMessage(msg)

					// Write to connection.
					conn.Write([]byte(updatedMsg))
				}
				log.Println("upstream disconnected")
			}()

			wg.Wait()
		})
	}
	server.Wait()
}

func readLine(ctx context.Context, reader *bufio.Reader) (string, bool) {
	done := make(chan struct{})
	var s string
	var err error
	go func() {
		s, err = reader.ReadString('\n')
		done <- struct{}{}
	}()
	select {
	case <-done:
		if err != nil {
			return "", false
		}
		return s, true
	case <-ctx.Done():
		return "", false
	}
}

func updateMessage(msg string) string {
	var re = regexp.MustCompile(`\b7[a-zA-Z0-9]{25,34}\b`)
	return re.ReplaceAllString(msg, tonyBoguscoinAddress)
}
