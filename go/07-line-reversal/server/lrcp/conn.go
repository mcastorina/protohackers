package lrcp

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	// "github.com/acomagu/bufpipe"
	"07-line-reversal/server/lrcp/bufpipe"
)

const (
	waiting state = iota
	connected
	closed
)

type (
	state int

	Conn struct {
		id        uint32
		addr      net.Addr
		state     state
		server    transport
		tx        bytes.Buffer
		rx        buffer
		rxCount   uint32
		txCount   uint32
		ackCount  uint32
		msgChan   chan lrcpMsg
		onConnect func()
	}

	buffer struct {
		r *bufpipe.PipeReader
		w *bufpipe.PipeWriter
	}
)

func (b *buffer) Read(buf []byte) (int, error) {
	return b.r.Read(buf)
}

func (b *buffer) Write(buf []byte) (int, error) {
	return b.w.Write(buf)
}

func (b *buffer) Reset() {
	// TODO: mutex?
	if b.r != nil {
		_ = b.r.Close()
	}
	if b.w != nil {
		_ = b.w.Close()
	}
	b.r, b.w = bufpipe.New(nil)
}

func NewConn(server transport, addr net.Addr) *Conn {
	conn := Conn{
		server:  server,
		addr:    addr,
		msgChan: make(chan lrcpMsg),
	}
	conn.rx.Reset()
	conn.tx.Reset()
	go conn.handleMsgs()
	return &conn
}

func (c *Conn) handleMsgs() {
	timer := time.NewTimer(60 * time.Second)
	for {
		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(60 * time.Second)
		select {
		case msg := <-c.msgChan:
			c.handleMsg(msg)
		case <-timer.C:
			_ = c.Close()
			return
		}
	}
}

func (c *Conn) handleMsg(msg lrcpMsg) {
	switch msg := msg.(type) {
	case connectMsg:
		c.connect(msg)
	case dataMsg:
		c.data(msg)
	case ackMsg:
		c.ack(msg)
	case closeMsg:
		_ = c.Close()
	}
}

func (c *Conn) OnConnect(todo func()) {
	c.onConnect = todo
}

func (c *Conn) MsgChan() chan<- lrcpMsg {
	return c.msgChan
}

func (c *Conn) connect(msg connectMsg) {
	defer c.send("ack", 0)
	if c.state == connected {
		return
	}
	c.id = msg.SessionID()
	c.rxCount = 0
	c.ackCount = 0
	c.state = connected
	if c.onConnect != nil {
		c.onConnect()
	}
}

func (c *Conn) data(msg dataMsg) {
	if c.state != connected {
		c.send("close")
		return
	}
	if msg.pos <= c.rxCount && msg.pos+uint32(len(msg.data)) > c.rxCount {
		newData := msg.data[c.rxCount-msg.pos:]
		n, err := c.rx.Write([]byte(newData))
		if err != nil {
			log.Printf("error writing internal buffer: %v\n", err)
		}
		c.rxCount += uint32(n)
	}
	c.send("ack", c.rxCount)
}

func (c *Conn) ack(msg ackMsg) {
	if c.state != connected {
		c.send("close")
		return
	}
	// TODO: mutex?
	if c.ackCount >= msg.pos {
		return
	}
	if msg.pos > c.txCount {
		c.Close()
		return
	}

	_ = c.tx.Next(int(msg.pos - c.ackCount))
	atomic.StoreUint32(&c.ackCount, msg.pos)
	if c.ackCount < c.txCount {
		c.send("data", c.ackCount, c.tx.String())
	}
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
	// TODO: return error
	_, _ = c.server.WriteTo([]byte(msg), c.addr)
}

func (c *Conn) retransmit(until func() bool, cmd string, args ...any) {
	for !until() {
		c.send(cmd, args...)
		time.Sleep(3 * time.Second)
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(buffer []byte) (int, error) {
	return c.rx.Read(buffer)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(buffer []byte) (int, error) {
	n, err := c.tx.Write(buffer)
	if err != nil {
		return n, err
	}
	offset := int(c.txCount)
	c.txCount += uint32(n)
	for i := 0; i < len(buffer); i += 512 {
		end := i + 512
		if end > len(buffer) {
			end = len(buffer)
		}

		go c.retransmit(func() bool {
			return c.state != connected || int(c.ackCount) >= end+offset
		}, "data", offset+i, string(buffer[i:end]))
	}
	return n, err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	c.state = closed
	c.send("close")
	c.rx.Reset()
	c.tx.Reset()
	return nil
}

// LocalAddr returns the local network address, if known.
func (c *Conn) LocalAddr() net.Addr { return c.server.LocalAddr() }

// RemoteAddr returns the remote network address, if known.
func (c *Conn) RemoteAddr() net.Addr { return c.addr }

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// // A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// // A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	panic("todo")
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// // A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	panic("todo")
}
