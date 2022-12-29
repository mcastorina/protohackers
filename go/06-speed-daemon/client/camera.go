package client

import "errors"

type CameraClient struct {
	Client
	Road  uint16
	Mile  uint16
	Limit uint16
}

type PlateInfo struct {
	Plate     string
	Timestamp uint32
}

func (c *CameraClient) ReadPlate() (*PlateInfo, error) {
	kind, err := c.ReadU8()
	if err != nil {
		return nil, err
	}
	if kind != 0x20 {
		if err := c.rbuf.UnreadByte(); err != nil {
			return nil, err
		}
		return nil, errors.New("not a plate")
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
