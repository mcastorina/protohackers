package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"sync"

	"06-speed-daemon/client"
)

type Observation struct {
	Mile      uint16
	Timestamp uint32
}

func (o Observation) Speed(other Observation) uint16 {
	dist := float64(o.Mile) - float64(other.Mile)
	sec := float64(o.Timestamp) - float64(other.Timestamp)

	if dist == 0 || sec == 0 {
		return 0
	}

	// mile/sec -> mile/hour (100x)
	return uint16(math.Round(360000 * dist / sec))
}

type ObservationsMap struct {
	lock                     sync.RWMutex
	plateToRoadToObservation map[string]map[uint16][]Observation
}

type Ticket struct {
	plate      string
	road       uint16
	mile1      uint16
	timestamp1 uint32
	mile2      uint16
	timestamp2 uint32
	speed      uint16
}

func (o *ObservationsMap) addObservation(camera *client.CameraClient, plateInfo *client.PlateInfo) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.plateToRoadToObservation[plateInfo.Plate] == nil {
		o.plateToRoadToObservation[plateInfo.Plate] = make(map[uint16][]Observation)
	}

	observation := Observation{
		Mile:      camera.Mile,
		Timestamp: plateInfo.Timestamp,
	}

	o.plateToRoadToObservation[plateInfo.Plate][camera.Road] = append(o.plateToRoadToObservation[plateInfo.Plate][camera.Road], observation)
}

func (o *ObservationsMap) calculateTickets(camera *client.CameraClient, plateInfo *client.PlateInfo) []Ticket {
	o.lock.RLock()
	defer o.lock.RUnlock()

	latestObservation := Observation{
		Mile:      camera.Mile,
		Timestamp: plateInfo.Timestamp,
	}

	var tickets []Ticket
	for _, observation := range o.plateToRoadToObservation[plateInfo.Plate][camera.Road] {
		if speed := observation.Speed(latestObservation); speed > 100*camera.Limit {
			ob1, ob2 := observation, latestObservation
			if ob1.Timestamp > ob2.Timestamp {
				ob2, ob1 = ob1, ob2
			}
			tickets = append(tickets, Ticket{
				plate:      plateInfo.Plate,
				road:       camera.Road,
				mile1:      ob1.Mile,
				timestamp1: ob1.Timestamp,
				mile2:      ob2.Mile,
				timestamp2: ob2.Timestamp,
				speed:      speed,
			})
		}
	}
	return tickets
}

type Ticketmaster struct {
	lock                sync.Mutex
	dayToPlatesObserved map[string]map[string]bool
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	observationsMap := ObservationsMap{
		plateToRoadToObservation: make(map[string]map[uint16][]Observation),
	}

	for conn := range server.Connections() {
		server.Handle(conn, func(conn net.Conn) {
			c := client.New(conn)
			if camera, err := c.AsCamera(); err == nil {
				fmt.Println("camera", camera)
				for {
					if plateInfo, err := camera.ReadPlate(); err == nil {
						// TODO play with plateinfo logic
						fmt.Println(plateInfo)
						observationsMap.addObservation(camera, plateInfo)
						if tickets := observationsMap.calculateTickets(camera, plateInfo); len(tickets) > 0 {
							fmt.Println(tickets)
						}
						fmt.Println(observationsMap.plateToRoadToObservation)
					} else {
						// Return an error and disconnect.
						_ = camera.SendError(err)
						return
					}
				}
			} else if dispatcher, err := c.AsDispatcher(); err == nil {
				fmt.Println("dispatcher", dispatcher)
			} else {
				// Unrecognized device.
				return
			}
			fmt.Println(c)
		})
	}
	server.Wait()
}
