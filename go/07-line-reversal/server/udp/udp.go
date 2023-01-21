package udp

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

type Server struct {
	net.PacketConn
	packets chan Packet
	cancel  context.CancelFunc
	workers sync.WaitGroup
}

func NewServer() (*Server, error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	var lc net.ListenConfig

	var addr string
	if _, ok := os.LookupEnv("FLY_APP_NAME"); ok {
		addr = "fly-global-services"
	}

	listener, err := lc.ListenPacket(ctx, "udp", addr+":1337")
	if err != nil {
		return nil, err
	}
	log.Println("listening on udp :1337")

	server := Server{
		PacketConn: listener,
		packets:    make(chan Packet),
		cancel:     cancel,
	}
	server.workers.Add(1)

	go server.closeServerOnCancel(ctx)
	go func() {
		defer server.workers.Done()
		server.readPackets(ctx)
	}()

	return &server, nil
}

func (s *Server) Packets() <-chan Packet {
	return s.packets
}

// Close shuts down the server.
func (s *Server) Close() error {
	defer func() {
		s.workers.Wait()
		close(s.packets)
	}()
	log.Println("shutting down")
	s.cancel()
	return s.PacketConn.Close()
}

// Wait for the context to cancel, then call s.Close().
func (s *Server) closeServerOnCancel(ctx context.Context) {
	<-ctx.Done()
	s.Close()
}

func (s *Server) readPackets(ctx context.Context) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := s.PacketConn.ReadFrom(buf)
		// Check if the context was canceled.
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Println(err)
			continue
		}
		select {
		case s.packets <- Packet{
			Data:   buf[:n],
			Addr:   addr,
			server: s,
		}:
		case <-ctx.Done():
			return
		}
	}
}

type Packet struct {
	Data   []byte
	Addr   net.Addr
	server *Server
}

func (p *Packet) Reply(data []byte) error {
	_, err := p.server.PacketConn.WriteTo(data, p.Addr)
	return err
}
