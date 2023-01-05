package client

type Camera struct {
	*Client
	CameraDetails
}

type CameraDetails struct {
	Road  uint16
	Mile  uint16
	Limit uint16
}

type PlateInfo struct {
	Plate     string
	Timestamp uint32
}

func (c *Camera) ReadPlate() (*PlateInfo, error) {
	if err := c.Expect(0x20); err != nil {
		return nil, err
	}

	plate, err := c.ReadStr()
	if err != nil {
		return nil, err
	}

	timestamp, err := c.ReadU32()
	if err != nil {
		return nil, err
	}

	return &PlateInfo{
		Plate:     plate,
		Timestamp: timestamp,
	}, nil
}

type Reading struct {
	CameraDetails
	PlateInfo
}

func (c *Camera) Read() (*Reading, error) {
	plate, err := c.ReadPlate()
	if err != nil {
		return nil, err
	}
	return &Reading{
		CameraDetails: c.CameraDetails,
		PlateInfo:     *plate,
	}, nil
}

func (c *Camera) Run() <-chan *Reading {
	ch := make(chan *Reading)
	c.workersWG.Add(1)
	go func() {
		defer c.workersWG.Done()
		defer close(ch)
		for {
			if reading, err := c.Read(); err == nil {
				ch <- reading
				continue
			}
			if err := c.HandleCommonMessage(); err != nil {
				// Return an error and disconnect.
				_ = c.SendError(err)
				return
			}
		}
	}()
	return ch
}
