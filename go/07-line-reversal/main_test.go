package main

import (
	"testing"

	"gotest.tools/assert"
)

// parseLRCP parses a LRCP packet into its parts
func TestParseLRCP(t *testing.T) {
	got, err := parseLRCP("/foo/bar/baz/")
	assert.NilError(t, err)
	assert.Equal(t, 3, len(got))
	for i, expected := range []string{"foo", "bar", "baz"} {
		assert.Equal(t, expected, got[i])
	}

	got, err = parseLRCP("/data/1234567/0/hello/")
	assert.NilError(t, err)
	assert.Equal(t, 4, len(got))
	for i, expected := range []string{"data", "1234567", "0", "hello"} {
		assert.Equal(t, expected, got[i])
	}
}
