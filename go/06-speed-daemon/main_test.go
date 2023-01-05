//go:build integration

package main

import (
	"06-speed-daemon/client"
	"net"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func connectAsClient(t *testing.T) client.Client {
	conn, err := net.Dial("tcp", "127.0.0.1:1337")
	assert.NilError(t, err)
	return client.New(conn)
}

func mustRead[T uint8 | uint16 | uint32 | string](t *testing.T, f func() (T, error)) T {
	val, err := f()
	assert.NilError(t, err)
	return val
}

func TestMain(m *testing.M) {
	go main()
	time.Sleep(10 * time.Millisecond)
	os.Exit(m.Run())
}

func TestIntegration(t *testing.T) {
	camera1 := connectAsClient(t)
	camera2 := connectAsClient(t)
	dispatcher := connectAsClient(t)

	camera1.WriteAll(0x80, 0x00, 0x7b, 0x00, 0x08, 0x00, 0x3c)
	camera1.WriteAll(0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x00, 0x00)

	camera2.WriteAll(0x80, 0x00, 0x7b, 0x00, 0x09, 0x00, 0x3c)
	camera2.WriteAll(0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x00, 0x2d)

	dispatcher.WriteAll(0x81, 0x1, 0x0, 0x7b)
	// Ticket{plate: "UN1X", road: 123, mile1: 8, timestamp1: 0, mile2: 9, timestamp2: 45, speed: 8000}
	assert.Equal(t, uint8(0x21), mustRead(t, dispatcher.ReadU8))
	got := client.Ticket{
		Plate:      mustRead(t, dispatcher.ReadStr),
		Road:       mustRead(t, dispatcher.ReadU16),
		Mile1:      mustRead(t, dispatcher.ReadU16),
		Timestamp1: mustRead(t, dispatcher.ReadU32),
		Mile2:      mustRead(t, dispatcher.ReadU16),
		Timestamp2: mustRead(t, dispatcher.ReadU32),
		Speed:      mustRead(t, dispatcher.ReadU16),
	}
	assert.Equal(t, client.Ticket{
		Plate:      "UN1X",
		Road:       123,
		Mile1:      8,
		Timestamp1: 0,
		Mile2:      9,
		Timestamp2: 45,
		Speed:      8000,
	}, got)
}

func TestHeartbeat(t *testing.T) {
	client := connectAsClient(t)
	client.WriteAll(0x40, uint32(10))
	start := time.Now()
	assert.Equal(t, uint8(0x41), mustRead(t, client.ReadU8))
	assert.Assert(t, time.Since(start) >= 1*time.Second)
}
