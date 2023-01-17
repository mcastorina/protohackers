package lrcp

import (
	"testing"

	"gotest.tools/assert"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		input    string
		expected lrcpMsg
	}{
		{"", nil},
		{"/", nil},
		{"//", nil},
		{"/foo/", nil},
		{"/foo/bar/", nil},
		{"/foo/123/", nil},
		{"/connect/bar/", nil},
		{"/close/bar/", nil},
		{"/ack/bar/", nil},
		{"/ack/bar/baz/", nil},
		{"/ack/123/baz/", nil},
		{"/data/foo/", nil},
		{"/data/foo/bar/", nil},
		{"/data/foo/bar/baz/", nil},
		{"/data/123/bar/baz/", nil},
		{"/data/foo/456/baz/", nil},
		{"/connect/123/", connectMsg{123}},
		{"/close/123/", closeMsg{123}},
		{"/ack/123/456/", ackMsg{123, 456}},
		{"/data/123/456/baz/", dataMsg{123, 456, "baz"}},
		{"/data/123/456/\\/\\\\baz\\/\\\\/", dataMsg{123, 456, "/\\baz/\\"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseMsg([]byte(tt.input))
			assert.Equal(t, tt.expected == nil, err != nil)
			assert.Equal(t, tt.expected, got)
		})
	}
}
