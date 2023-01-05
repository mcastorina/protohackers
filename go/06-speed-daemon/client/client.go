package client

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type Client struct {
	rbuf      *bufio.Reader
	wbuf      *bufio.Writer
	heartbeat *time.Duration
	writeLock sync.Mutex
	workersWG sync.WaitGroup
}

func New(conn io.ReadWriter) Client {
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

func (c *Client) writeAllBytes(data []byte) error {
	sent := 0
	for i := 0; sent < len(data) && i < 5; i++ {
		n, _ := c.wbuf.Write(data[sent:])
		sent += n
	}
	if sent == len(data) {
		return nil
	}
	return errors.New("could not send all data")
}

func (c *Client) WriteU8(arg uint8) error {
	return c.wbuf.WriteByte(arg)
}

func (c *Client) WriteU16(arg uint16) error {
	writeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(writeBytes, arg)
	return c.writeAllBytes(writeBytes)
}

func (c *Client) WriteU32(arg uint32) error {
	writeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(writeBytes, arg)
	return c.writeAllBytes(writeBytes)
}

func (c *Client) WriteStr(arg string) error {
	length := len([]byte(arg))
	err := c.wbuf.WriteByte(uint8(length))
	if err != nil {
		return err
	}
	return c.writeAllBytes([]byte(arg))
}

func (c *Client) Flush() error {
	if err := c.wbuf.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}
	return nil
}

func (c *Client) AsCamera() (*Camera, error) {
	if err := c.Expect(0x80); err != nil {
		return nil, err
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

	return &Camera{
		Client: c,
		CameraDetails: CameraDetails{
			Road:  road,
			Mile:  mile,
			Limit: limit,
		},
	}, nil
}

func (c *Client) AsDispatcher() (*Dispatcher, error) {
	if err := c.Expect(0x81); err != nil {
		return nil, err
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

	return &Dispatcher{
		Client: c,
		Roads:  roadIDs,
	}, nil
}

func (c *Client) SendError(err error) error {
	return c.WriteAll(0x10, err.Error())
}

func (c *Client) HandleCommonMessage() error {
	hb, err := c.WantHeartbeat()
	if err != nil {
		return err
	}
	if c.heartbeat != nil {
		return errors.New("heartbeat already exists")
	}
	c.heartbeat = &hb
	if hb == 0 {
		return nil
	}
	c.workersWG.Add(1)
	go func() {
		defer c.workersWG.Done()
		ticker := time.NewTicker(hb)
		defer ticker.Stop()
		for range ticker.C {
			if err := c.WriteHeartbeat(); err != nil {
				return
			}
		}
	}()
	return nil
}

func (c *Client) WantHeartbeat() (time.Duration, error) {
	if err := c.Expect(0x40); err != nil {
		return 0, err
	}
	heartbeat, err := c.ReadU32()
	if err != nil {
		return 0, err
	}
	return time.Duration(heartbeat) * 100 * time.Millisecond, nil
}

func (c *Client) Expect(want uint8) error {
	kind, err := c.ReadU8()
	if err != nil {
		return err
	}
	if kind != want {
		if err := c.rbuf.UnreadByte(); err != nil {
			return err
		}
		return errors.New("unexpected id")
	}
	return nil
}

func (c *Client) WriteHeartbeat() error {
	return c.WriteAll(0x41)
}

func (c *Client) WriteAll(fields ...any) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	for _, field := range fields {
		var err error
		switch field := field.(type) {
		case int:
			err = c.WriteU8(uint8(field))
		case uint8:
			err = c.WriteU8(field)
		case uint16:
			err = c.WriteU16(field)
		case uint32:
			err = c.WriteU32(field)
		case string:
			err = c.WriteStr(field)
		default:
			panic(fmt.Sprintf("write called with unexpected type: %T", field))
		}
		if err != nil {
			return err
		}
	}
	return c.Flush()
}

func (c *Client) Wait() {
	c.workersWG.Wait()
}
