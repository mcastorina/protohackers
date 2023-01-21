package lrcp

import (
	"07-line-reversal/server/udp"
	"bytes"
	"errors"
	"sync/atomic"
	"time"

	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/acomagu/bufpipe"
)

type (
	transport interface {
		Packets() <-chan udp.Packet
		WriteTo([]byte, net.Addr) error
	}

	Server struct {
		server  transport
		workers sync.WaitGroup
		conns   chan *Conn
	}
	Conn struct {
		// Buffered pipe transport -> application layer.
		appRead   *bufpipe.PipeReader
		appWrite  *bufpipe.PipeWriter
		appData   bytes.Buffer
		id        uint32
		addr      net.Addr
		server    transport
		readCount uint32
		ackCount  uint32
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

func (c *Conn) Close() error {
	if !c.Open() {
		return nil
	}
	c.send("close")
	c.addr = nil
	_ = c.appRead.Close()
	_ = c.appWrite.Close()
	return nil
}

func (c *Conn) connect(addr net.Addr, msg connectMsg) {
	c.addr = addr
	c.send("ack", 0)
}

func (c *Conn) data(msg dataMsg) {
	// If the session is not open: send close and stop.
	if !c.Open() {
		c.send("close")
		return
	}

	// If we received any new data, add it to the buffer.
	// 1. First check we have some overlap in the position we received.
	// 2. Then check to make sure the amount of data we received is more than
	//    our current read position.
	if msg.pos <= c.readCount && msg.pos+uint32(len(msg.data)) > c.readCount {
		newData := msg.data[c.readCount-msg.pos:]
		n, err := c.appWrite.Write([]byte(newData))
		if err != nil {
			log.Printf("error writing internal buffer: %v\n", err)
		}
		c.readCount += uint32(n)
	}

	// ACK with how much data we've read.
	c.send("ack", c.readCount)
}

func (c *Conn) ack(msg ackMsg) {
	if c.ackCount >= msg.pos {
		return
	}
	// Check that the ack makes sense: the number of acknowledged bytes is not
	// more than the bytes we have sent.
	toDrop := int(msg.pos - c.ackCount)
	if toDrop > c.appData.Len() {
		return
	}
	// We have confirmation that these bytes have been received, so we can drop
	// this data.
	_ = c.appData.Next(int(msg.pos - c.ackCount))
	atomic.StoreUint32(&c.ackCount, msg.pos)
}

func (c *Conn) send(cmd string, args ...any) {
	parts := make([]string, len(args)+2)
	parts[0] = cmd
	parts[1] = strconv.FormatInt(int64(c.id), 10)
	for i, arg := range args {
		switch arg := arg.(type) {
		case string:
			parts[i+2] = escape(arg)
		default:
			parts[i+2] = fmt.Sprintf("%v", arg)
		}
	}
	msg := fmt.Sprintf("/%s/", strings.Join(parts, "/"))
	_ = c.server.WriteTo([]byte(msg), c.addr)
}

func (c *Conn) Open() bool {
	return c.addr != nil
}

func (c *Conn) Read(buffer []byte) (int, error) {
	return c.appRead.Read(buffer)
}

func (c *Conn) Write(buffer []byte) (int, error) {
	if !c.Open() {
		return 0, errors.New("transport already closed")
	}
	for i := 0; i < len(buffer); i += 512 {
		end := i + 512
		if end > len(buffer) {
			end = len(buffer)
		}
		c.sendData(buffer[i:end])
	}
	return len(buffer), nil
}

func (c *Conn) sendData(data []byte) {
	pos := c.ackCount + uint32(c.appData.Len())
	msg := string(data)
	go func() {
		for c.Open() && c.ackCount < pos+uint32(len(data)) {
			c.send("data", pos, msg)
			time.Sleep(3 * time.Second)
		}
	}()
	c.appData.Write(data)
}
