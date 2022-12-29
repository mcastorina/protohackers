package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"06-speed-daemon/client"
)

type Observation struct {
	Mile      uint16
	Timestamp uint32
}

type ObservationsMap struct {
	lock                     sync.Mutex
	plateToRoadToObservation map[string]map[uint16][]Observation
}

func (o *ObservationsMap) addObservation(camera *client.CameraClient, plateInfo *client.PlateInfo) {
	o.lock.Lock()
	defer o.lock.Unlock()
	road := o.plateToRoadToObservation[plateInfo.Plate]
	if road == nil {
		o.plateToRoadToObservation[plateInfo.Plate] = make(map[uint16][]Observation)
	}

	observation := Observation{
		Mile:      camera.Mile,
		Timestamp: plateInfo.Timestamp,
	}

	o.plateToRoadToObservation[plateInfo.Plate][camera.Road] = append(o.plateToRoadToObservation[plateInfo.Plate][camera.Road], observation)
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
						fmt.Printf("%v\n", observationsMap.plateToRoadToObservation)

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
