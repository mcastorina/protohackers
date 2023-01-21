package lrcp

import (
	"07-line-reversal/server/udp"

	"log"
	"net"
	"sync"

	"github.com/acomagu/bufpipe"
)

type (
	transport interface {
		Packets() <-chan udp.Packet
		WriteTo([]byte, net.Addr) error
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
		defer s.workers.Done()
		todo(conn)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()
}

func (s *Server) listen() {
	conns := make(map[uint32]*Conn)
	for packet := range s.server.Packets() {
		msg, err := parseMsg(packet.Data)
		if err != nil {
			continue
		}
		// Create connection if it doesn't exist.
		id := msg.SessionID()
		if _, ok := conns[id]; !ok {
			conns[id] = s.newConn(id)
		}
		// TODO: Send messages over a channel to the Conn.
		conn := conns[id]

		switch msg := msg.(type) {
		case connectMsg:
			conn.connect(packet.Addr, msg)
			s.conns <- conn
		case dataMsg:
			conn.data(msg)
		case ackMsg:
			conn.ack(msg)
		case closeMsg:
			_ = conn.Close()
			delete(conns, id)
		}
	}
}

func (s *Server) newConn(id uint32) *Conn {
	r, w := bufpipe.New(nil)
	return &Conn{
		id:       id,
		server:   s.server,
		appRead:  r,
		appWrite: w,
	}
}
