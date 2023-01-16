package lrcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"07-line-reversal/server/udp"
)

type (
	Server struct {
		server  *udp.Server
		workers sync.WaitGroup
		conns   chan *Conn
	}
	Conn struct {
		bufio.Reader
		bufio.Writer
		id     uint32
		addr   net.Addr
		server *udp.Server
	}
)

func NewServer() (*Server, error) {
	udpServer, err := udp.NewServer()
	if err != nil {
		return nil, err
	}

	server := &Server{
		server: udpServer,
		conns:  make(chan *Conn),
	}

	server.workers.Add(1)
	go func() {
		defer server.workers.Done()
		server.listen()
	}()

	return server, nil
}

func (s *Server) Connections() <-chan *Conn {
	return s.conns
}

func (s *Server) Handle(conn *Conn, todo func(conn *Conn)) {
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
			conns[id] = &Conn{id: id, server: s.server}
		}
		conn := conns[id]

		switch msg := msg.(type) {
		case connectMsg:
			conn.connect(packet.Addr, msg)
		case dataMsg:
			conn.data(msg)
		case ackMsg:
			conn.ack(msg)
		case closeMsg:
			_ = s.server.WriteTo(packet.Data, conn.addr)
			delete(conns, id)
		}
	}
}

func (c *Conn) Close() error {
	panic("todo")
}

func (c *Conn) connect(addr net.Addr, msg connectMsg) {
	c.addr = addr
	c.send("ack", 0)
}

func (c *Conn) data(msg dataMsg) {
	if !c.Open() {
		c.send("close", c.id)
		return
	}
}

func (c *Conn) ack(msg ackMsg) {
	panic("todo")
}

func (c *Conn) send(a ...any) {
	parts := make([]string, len(a))
	for i, a := range a {
		switch a := a.(type) {
		case string:
			parts[i] = escape(a)
		default:
			parts[i] = fmt.Sprintf("%v", a)
		}
	}
	msg := fmt.Sprintf("/%s/", strings.Join(parts, "/"))
	_ = c.server.WriteTo([]byte(msg), c.addr)
}

func (c *Conn) Open() bool {
	return c.addr != nil
}

func escape(s string) string {
	replacer := strings.NewReplacer("/", "\\/", "\\", "\\\\")
	return replacer.Replace(s)
}
