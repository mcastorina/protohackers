package lrcp

import (
	"07-line-reversal/server/udp"
	"bufio"
	"io"

	"net"
	"strings"
	"testing"

	"gotest.tools/assert"
)

type (
	testTransport struct {
		ch              chan udp.Packet
		wbuf            strings.Builder
		waitForResponse chan struct{}
	}
	testAddr struct{}
)

func (testAddr) Network() string { return "test" }
func (testAddr) String() string  { return "test" }

func (t *testTransport) Packets() <-chan udp.Packet {
	return t.ch
}
func (t *testTransport) WriteTo(data []byte, _ net.Addr) error {
	_, err := t.wbuf.Write(data)
	t.waitForResponse <- struct{}{}
	return err
}
func (t *testTransport) send(msg string) {
	t.ch <- udp.Packet{
		Data: []byte(msg),
		Addr: testAddr{},
	}
	<-t.waitForResponse
}

func TestTransport(t *testing.T) {
	transport := testTransport{
		ch:              make(chan udp.Packet),
		waitForResponse: make(chan struct{}, 5),
	}
	server, err := NewServerTransport(&transport)
	assert.NilError(t, err)

	// Connect.
	transport.send("/connect/123456/")
	assert.Equal(t, "/ack/123456/0/", transport.wbuf.String())
	transport.wbuf.Reset()

	conn := <-server.Connections()
	assert.Equal(t, true, conn.Open())

	// Send some data that we shouldn't do anything with.
	transport.send("/data/123456/1/oobar/")
	assert.Equal(t, "/ack/123456/0/", transport.wbuf.String())
	transport.wbuf.Reset()

	// Send initial data.
	transport.send("/data/123456/0/foobar/")
	assert.Equal(t, "/ack/123456/6/", transport.wbuf.String())
	transport.wbuf.Reset()

	// Send redundant data.
	transport.send("/data/123456/3/ba/")
	assert.Equal(t, "/ack/123456/6/", transport.wbuf.String())
	transport.wbuf.Reset()

	// Send partially redundant and new data.
	transport.send("/data/123456/5/rbazbuzz/")
	assert.Equal(t, "/ack/123456/13/", transport.wbuf.String())
	transport.wbuf.Reset()

	// Send all new data.
	transport.send("/data/123456/13/!!!/")
	assert.Equal(t, "/ack/123456/16/", transport.wbuf.String())
	transport.wbuf.Reset()

	// Close the connection.
	transport.send("/close/123456/")
	assert.Equal(t, "/close/123456/", transport.wbuf.String())
	transport.wbuf.Reset()
	assert.Equal(t, false, conn.Open())

	// Try to send data on a closed connection.
	transport.send("/data/123456/0/foobar/")
	assert.Equal(t, "/close/123456/", transport.wbuf.String())
	transport.wbuf.Reset()
}

func TestConnRead(t *testing.T) {
	transport := testTransport{
		ch:              make(chan udp.Packet),
		waitForResponse: make(chan struct{}, 5),
	}
	server, err := NewServerTransport(&transport)
	assert.NilError(t, err)

	// Connect.
	transport.send("/connect/123456/")
	assert.Equal(t, "/ack/123456/0/", transport.wbuf.String())

	conn := <-server.Connections()
	assert.Equal(t, true, conn.Open())
	reader := bufio.NewReader(conn)
	readLine := func() string {
		s, err := reader.ReadString('\n')
		assert.NilError(t, err)
		return s
	}

	// Send initial data.
	transport.send("/data/123456/0/foobar\n/")
	assert.Equal(t, "foobar\n", readLine())

	// Send data accross multiple packets.
	transport.send("/data/123456/7/hello /")
	transport.send("/data/123456/13/world\n/")
	transport.send("/data/123456/19/extra bits\n/")
	assert.Equal(t, "hello world\n", readLine())

	// Close the connection.
	conn.Close()
	assert.Equal(t, false, conn.Open())

	// Reading should result in 'extra bits' followed by EOF.
	assert.Equal(t, "extra bits\n", readLine())
	_, err = reader.ReadString('\n')
	assert.Equal(t, io.EOF, err)
}
