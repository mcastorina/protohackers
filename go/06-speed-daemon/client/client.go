package client

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

type Client struct {
	rbuf *bufio.Reader
	wbuf *bufio.Writer
}

func New(conn io.ReadWriteCloser) Client {
	return Client{
		rbuf: bufio.NewReader(conn),
		wbuf: bufio.NewWriter(conn),
	}
}

func (c *Client) ReadU8() (uint8, error) {
	return c.rbuf.ReadByte()
}

func (c *Client) ReadU16() (uint16, error) {
	data, err := io.ReadAll(io.LimitReader(c.rbuf, 2))
	if err != nil {
		return 0, err
	}
	if len(data) != 2 {
		return 0, errors.New("not enough bytes")
	}
	return binary.BigEndian.Uint16(data), nil
}

func (c *Client) ReadU32() (uint32, error) {
	data, err := io.ReadAll(io.LimitReader(c.rbuf, 4))
	if err != nil {
		return 0, err
	}
	if len(data) != 4 {
		return 0, errors.New("not enough bytes")
	}
	return binary.BigEndian.Uint32(data), nil
}

func (c *Client) ReadStr() (string, error) {
	length, err := c.rbuf.ReadByte()
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(io.LimitReader(c.rbuf, int64(length)))
	if err != nil {
		return "", err
	}
	if len(data) != int(length) {
		return "", errors.New("not enough bytes")
	}
	return string(data), nil
}

func (c *Client) WriteU8(arg uint8) error {
	return c.wbuf.WriteByte(arg)
}

func (c *Client) WriteU16(arg uint16) error {
	writeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(writeBytes, arg)
	_, err := c.wbuf.Write(writeBytes)
	return err
}

func (c *Client) WriteU32(arg uint32) error {
	writeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(writeBytes, arg)
	_, err := c.wbuf.Write(writeBytes)
	return err
}

func (c *Client) WriteStr(arg string) error {
	length := len([]byte(arg))
	err := c.wbuf.WriteByte(uint8(length))
	if err != nil {
		return err
	}
	_, err = c.wbuf.WriteString(arg)
	return err
}

func (c *Client) Flush() error {
	return c.wbuf.Flush()
}

func (c *Client) AsCamera() (*CameraClient, error) {
	kind, err := c.ReadU8()
	if err != nil {
		return nil, err
	}
	if kind != 0x80 {
		if err := c.rbuf.UnreadByte(); err != nil {
			return nil, err
		}
		return nil, errors.New("not a camera")
	}

	road, err := c.ReadU16()
	if err != nil {
		return nil, err
	}

	mile, err := c.ReadU16()
	if err != nil {
		return nil, err
	}

	limit, err := c.ReadU16()
	if err != nil {
		return nil, err
	}

	return &CameraClient{
		Client: *c,
		Road:   road,
		Mile:   mile,
		Limit:  limit,
	}, nil
}

func (c *Client) AsDispatcher() (*DispatcherClient, error) {
	kind, err := c.ReadU8()
	if err != nil {
		return nil, err
	}
	if kind != 0x81 {
		if err := c.rbuf.UnreadByte(); err != nil {
			return nil, err
		}
		return nil, errors.New("not a dispatcher")
	}

	numRoads, err := c.ReadU8()
	if err != nil {
		return nil, err
	}

	roadIDs := make([]uint16, numRoads)
	for i := 0; i < int(numRoads); i++ {
		roadID, err := c.ReadU16()
		if err != nil {
			return nil, err
		}
		roadIDs[i] = roadID
	}

	return &DispatcherClient{
		Client: *c,
		Roads:  roadIDs,
	}, nil
}

func (c *Client) SendError(err error) error {
	if err := c.WriteU8(0x10); err != nil {
		return err
	}
	if err := c.WriteStr(err.Error()); err != nil {
		return err
	}
	return c.Flush()
}
