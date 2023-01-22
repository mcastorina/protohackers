package lrcp

import (
	"07-line-reversal/server/udp"

	"log"
	"net"
	"sync"
)

type (
	transport interface {
		Packets() <-chan udp.Packet
		WriteTo([]byte, net.Addr) (int, error)
		LocalAddr() net.Addr
	}

	Server struct {
		server  transport
		workers sync.WaitGroup
		conns   chan net.Conn
	}
)

func NewServer() (*Server, error) {
	udpServer, err := udp.NewServer()
	if err != nil {
		return nil, err
	}

	return NewServerTransport(udpServer)
}

func NewServerTransport(t transport) (*Server, error) {
	server := &Server{
		server: t,
		conns:  make(chan net.Conn),
	}

	server.workers.Add(1)
	go func() {
		defer server.workers.Done()
		server.listen()
	}()

	return server, nil
}

func (s *Server) Connections() <-chan net.Conn {
	return s.conns
}

func (s *Server) Handle(conn net.Conn, todo func(conn net.Conn)) {
	s.workers.Add(1)
	go func() {
		log.Println("connection opened")
		defer s.workers.Done()
		todo(conn)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
		log.Println("connection closed")
	}()
}

func (s *Server) listen() {
	defer close(s.conns)
	chs := make(map[uint32]chan<- lrcpMsg, 32)
	for packet := range s.server.Packets() {
		msg, err := parseMsg(packet.Data)
		if err != nil {
			continue
		}
		// Create connection if it doesn't exist.
		id := msg.SessionID()
		if _, ok := chs[id]; !ok {
			conn := NewConn(s.server, packet.Addr)
			conn.OnConnect(func() { s.conns <- conn })
			chs[id] = conn.MsgChan()
		}
		chs[id] <- msg
	}
}

// Wait waits for all the server workers to finish.
func (s *Server) Wait() {
	s.workers.Wait()
}
