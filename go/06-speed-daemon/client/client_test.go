package client

import (
	"bytes"
	"io"
	"testing"

	"gotest.tools/v3/assert"
)

func TestReadU8(t *testing.T) {
	input := []byte{10, 55}

	rwc := struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		bytes.NewReader(input),
		nil,
		nil,
	}

	client := New(rwc)

	for _, b := range input {
		gotByte, _ := client.ReadU8()
		assert.Equal(t, b, gotByte)
	}
}
