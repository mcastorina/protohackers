package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sync"

	"golang.org/x/sync/errgroup"

	"06-speed-daemon/client"
)

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	coordinator := CoordinatorService{
		readingChan: make(chan *client.Reading),
		ticketChan:  make(chan client.Ticket),
		dispatchers: make(map[uint16]chan client.Ticket),
	}
	go coordinator.Run()

	for conn := range server.Connections() {
		server.Handle(conn, func(conn net.Conn) {
			c := client.New(conn)
			for {
				if camera, err := c.AsCamera(); err == nil {
					log.Println("camera connected", camera)
					coordinator.HandleCamera(camera)
					log.Println("camera finished", camera)
					return
				} else if dispatcher, err := c.AsDispatcher(); err == nil {
					log.Println("dispatcher connected", dispatcher)
					coordinator.HandleDispatcher(dispatcher)
					log.Println("dispatcher finished", dispatcher)
					return
				} else if err := c.HandleCommonMessage(); err == nil {
					continue
				}
				// Unrecognized device.
				_ = c.SendError(errors.New("unrecognized device"))
				return
			}
		})
	}
	server.Wait()
}

type CoordinatorService struct {
	// Funnel of all client readings.
	readingChan chan *client.Reading
	// Funnel of all generated tickets.
	ticketChan chan client.Ticket
	// Table to allow dispatchers to listen for tickets on a specific road.
	dispatchers     map[uint16]chan client.Ticket
	dispatchersLock sync.Mutex
}

func (c *CoordinatorService) Run() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.runTicketGenerator()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.runTicketDispatcher()
	}()
	wg.Wait()
}

func (c *CoordinatorService) runTicketGenerator() {
	// Lookup table for road -> plate -> camera readings
	lut := make(map[uint16]map[string][]*client.Reading)
	for reading := range c.readingChan {
		if _, ok := lut[reading.Road]; !ok {
			lut[reading.Road] = make(map[string][]*client.Reading)
		}
		for _, prevReading := range lut[reading.Road][reading.Plate] {
			ticket := c.calculateTicket(prevReading, reading)
			if ticket == nil {
				continue
			}
			c.ticketChan <- *ticket
		}
		lut[reading.Road][reading.Plate] = append(lut[reading.Road][reading.Plate], reading)
	}
}

func (c *CoordinatorService) runTicketDispatcher() {
	history := make(map[string]struct{})
mainLoop:
	for ticket := range c.ticketChan {
		for day := ticket.Timestamp1 / 86400; day <= ticket.Timestamp2/86400; day++ {
			key := fmt.Sprintf("%s:%d", ticket.Plate, day)
			if _, ok := history[key]; ok {
				// We have already issued a ticket that spans one of the days.
				continue mainLoop
			}
		}
		// No history for this ticket.
		ch := c.getTicketChan(ticket.Road)
		select {
		case ch <- ticket:
			for day := ticket.Timestamp1 / 86400; day <= ticket.Timestamp2/86400; day++ {
				key := fmt.Sprintf("%s:%d", ticket.Plate, day)
				history[key] = struct{}{}
			}
		default:
			log.Println("dropping ticket: buffer full", ticket)
		}
	}
}

func (c *CoordinatorService) calculateTicket(r1 *client.Reading, r2 *client.Reading) *client.Ticket {
	if r1.Road != r2.Road || r1.Limit != r2.Limit || r1.Plate != r2.Plate {
		return nil
	}
	if r1.Timestamp > r2.Timestamp {
		r2, r1 = r1, r2
	}
	speed := c.calculateSpeed(r1, r2)
	if speed <= 100*r1.Limit {
		return nil
	}
	return &client.Ticket{
		Plate:      r1.Plate,
		Road:       r1.Road,
		Mile1:      r1.Mile,
		Timestamp1: r1.Timestamp,
		Mile2:      r2.Mile,
		Timestamp2: r2.Timestamp,
		Speed:      speed,
	}
}

func (c *CoordinatorService) calculateSpeed(r1 *client.Reading, r2 *client.Reading) uint16 {
	dist := math.Abs(float64(r1.Mile) - float64(r2.Mile))
	sec := math.Abs(float64(r1.Timestamp) - float64(r2.Timestamp))

	if dist == 0 || sec == 0 {
		return 0
	}
	// miles per second -> miles per hour (fixed point with 2 decimals)
	return uint16(math.Round(100 * 3600 * dist / sec))
}

func (c *CoordinatorService) HandleCamera(camera *client.Camera) {
	for plate := range camera.Run() {
		c.readingChan <- plate
	}
	camera.Wait()
}

func (c *CoordinatorService) HandleDispatcher(dispatcher *client.Dispatcher) {
	g, ctx := errgroup.WithContext(context.Background())
	for _, road := range dispatcher.Roads {
		ticketChan := c.getTicketChan(road)
		g.Go(func() error {
			for {
				var ticket client.Ticket
				select {
				case ticket = <-ticketChan:
				case <-ctx.Done():
					return ctx.Err()
				}

				err := dispatcher.WriteTicket(ticket)
				if err == nil {
					continue
				}
				log.Println("error sending ticket", err)
				// Put the ticket back on the queue for another client to handle.
				select {
				case ticketChan <- ticket:
				default:
					log.Println("dropping ticket: buffer full", ticket)
				}
				return err
			}
		})
	}
	g.Go(func() error {
		errCh := make(chan error)
		go func() {
			if err := dispatcher.HandleCommonMessage(); err != nil {
				errCh <- err
			}
			errCh <- nil
		}()
		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err := g.Wait(); err != nil {
		// Return an error and disconnect.
		_ = dispatcher.SendError(err)
	}
	dispatcher.Wait()
}

func (c *CoordinatorService) getTicketChan(road uint16) chan client.Ticket {
	c.dispatchersLock.Lock()
	defer c.dispatchersLock.Unlock()
	if _, ok := c.dispatchers[road]; !ok {
		c.dispatchers[road] = make(chan client.Ticket, 250)
	}
	return c.dispatchers[road]
}
