package client

import (
	"bytes"
	"io"
	"testing"

	"gotest.tools/v3/assert"
)

type rwc struct {
	io.Reader
	io.Writer
	io.Closer
}

func readTester(reader io.Reader) rwc {
	return rwc{Reader: reader}
}

func TestReadU8(t *testing.T) {
	input := []byte{10, 55}
	client := New(readTester(bytes.NewReader(input)))

	for _, b := range input {
		gotByte, _ := client.ReadU8()
		assert.Equal(t, b, gotByte)
	}
}

func TestReadU16(t *testing.T) {
	input := []byte{0x10, 0x55, 0x13, 0x15, 0x20}
	client := New(readTester(bytes.NewReader(input)))

	gotNum, _ := client.ReadU16()
	assert.Equal(t, uint16(0x1055), gotNum)

	gotNum, _ = client.ReadU16()
	assert.Equal(t, uint16(0x1315), gotNum)

	_, err := client.ReadU16()
	assert.Error(t, err, "not enough bytes")
}

func TestReadU32(t *testing.T) {
	input := []byte{0x10, 0x55, 0x13, 0x15, 0x20}
	client := New(readTester(bytes.NewReader(input)))

	gotNum, _ := client.ReadU32()
	assert.Equal(t, uint32(0x10551315), gotNum)

	_, err := client.ReadU32()
	assert.Error(t, err, "not enough bytes")
}

func TestReadStr(t *testing.T) {
	input := []byte{0x3, 0x62, 0x61, 0x64, 0x55, 0x13, 0x15, 0x20}
	client := New(readTester(bytes.NewReader(input)))

	gotStr, _ := client.ReadStr()
	assert.Equal(t, "bad", gotStr)

	_, err := client.ReadStr()
	assert.Error(t, err, "not enough bytes")
}

func writeTester(writer io.Writer) rwc {
	return rwc{Writer: writer}
}

func TestWriteU8(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	client := New(writeTester(buff))

	client.WriteU8(0x10)
	client.Flush()

	actualBytes := buff.Bytes()
	assert.Equal(t, 1, len(actualBytes))
	assert.Equal(t, byte(0x10), actualBytes[0])
}

func TestWriteU16(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	client := New(writeTester(buff))

	client.WriteU16(0x1025)
	client.Flush()

	actualBytes := buff.Bytes()
	assert.Equal(t, 2, len(actualBytes))
	assert.Equal(t, byte(0x10), actualBytes[0])
	assert.Equal(t, byte(0x25), actualBytes[1])
}

func TestWriteU32(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	client := New(writeTester(buff))

	client.WriteU32(0x10203040)
	client.Flush()

	actualBytes := buff.Bytes()
	assert.Equal(t, 4, len(actualBytes))
	assert.Equal(t, byte(0x10), actualBytes[0])
	assert.Equal(t, byte(0x20), actualBytes[1])
	assert.Equal(t, byte(0x30), actualBytes[2])
	assert.Equal(t, byte(0x40), actualBytes[3])
}

func TestWriteStr(t *testing.T) {
	input := "hello"
	buff := bytes.NewBuffer(nil)
	client := New(writeTester(buff))

	client.WriteStr(input)
	client.Flush()

	expectedBytes := append([]byte{byte(len(input))}, []byte(input)...)
	actualBytes := buff.Bytes()

	for i, actualByte := range actualBytes {
		assert.Equal(t, expectedBytes[i], actualByte)
	}
}
