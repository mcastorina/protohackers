package server

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

// Server is a struct providing a simple interface for managing TCP connections
// on port 1337.
type Server struct {
	listener net.Listener
	conns    chan net.Conn
	cancel   context.CancelFunc
	workers  sync.WaitGroup
}

// NewServer listens for TCP connections on port 1337 and handles cleaning up
// on an interrupt signal.
func NewServer() (*Server, error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", ":1337")
	if err != nil {
		return nil, err
	}
	log.Println("listening on tcp :1337")

	server := Server{
		listener: listener,
		conns:    make(chan net.Conn),
		cancel:   cancel,
	}
	server.workers.Add(1)

	go server.closeServerOnCancel(ctx)
	go func() {
		defer server.workers.Done()
		server.acceptConnections(ctx)
	}()

	return &server, nil
}

// Close shuts down the server.
func (s *Server) Close() {
	log.Println("shutting down")
	s.cancel()
	if err := s.listener.Close(); err != nil {
		log.Println(err)
	}
	s.workers.Wait()
	close(s.conns)
}

// Wait waits for all the server workers to finish.
func (s *Server) Wait() {
	s.workers.Wait()
}

// Handle spawns a goroutine to execute the handler, close the connection, and
// ensures graceful shutdown.
func (s *Server) Handle(conn net.Conn, todo func(conn net.Conn)) {
	s.workers.Add(1)
	go func() {
		defer s.workers.Done()
		todo(conn)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()
}

// Connections returns a channel of new connections.
func (s *Server) Connections() <-chan net.Conn {
	return s.conns
}

// Wait for the context to cancel, then call s.Close().
func (s *Server) closeServerOnCancel(ctx context.Context) {
	<-ctx.Done()
	s.Close()
}

// Accept new connections as long as the context isn't canceled and send them
// on the internal channel.
func (s *Server) acceptConnections(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		// Check if the context was canceled.
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("accepted new connection")
		select {
		case s.conns <- conn:
		case <-ctx.Done():
			return
		}
	}
}
