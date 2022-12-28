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
	panic("todo")
}

func (c *Client) WriteU32(arg uint32) error {
	panic("todo")
}

func (c *Client) WriteStr(arg string) error {
	panic("todo")
}

func (c *Client) Flush() error {
	panic("todo")
}
