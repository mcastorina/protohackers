package client

type Dispatcher struct {
	*Client
	Roads []uint16
}

type Ticket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}

func (d *Dispatcher) WriteTicket(t Ticket) error {
	return d.WriteAll(0x21,
		t.Plate, t.Road,
		t.Mile1, t.Timestamp1,
		t.Mile2, t.Timestamp2,
		t.Speed,
	)
}
