package main

import (
	"bufio"
	"context"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestExtractBoguscoin(t *testing.T) {
	var tests = []struct {
		message  string
		expected string
	}{
		{"abc", ""},
		{"", ""},
		{"7hadhjdfhbbbbasjajajasjsjaj", "7hadhjdfhbbbbasjajajasjsjaj"},
		{"7hadhjdfhbbbbasjajajasjsjajsjahadhjdfhbbbbasjajajasjsjajsja", ""},
		{"x 7hadhjdfhbbbbasjajajasjsjaj ", "7hadhjdfhbbbbasjajajasjsjaj"},
		{"x 7hadhjdfhbbbbasjajajasjsjaj ", "7hadhjdfhbbbbasjajajasjsjaj"},
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ", "7F1u3wSD5RbOHQmupo9nx4TnhQ"},
		{"7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX", "7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			extractedBogusCoin, _ := extractBoguscoin(tt.message)
			assert.Equal(t, tt.expected, extractedBogusCoin)
		})

	}

}

func TestReadLine(t *testing.T) {
	buffer := bufio.NewReader(strings.NewReader("hello\nworld\n"))
	got, ok := readLine(context.Background(), buffer)
	assert.Equal(t, true, ok)
	assert.Equal(t, "hello\n", got)

	got, ok = readLine(context.Background(), buffer)
	assert.Equal(t, true, ok)
	assert.Equal(t, "world\n", got)

	got, ok = readLine(context.Background(), buffer)
	assert.Equal(t, false, ok)
	assert.Equal(t, "", got)
}

func TestReadLineCancel(t *testing.T) {
	buffer := bufio.NewReader(strings.NewReader("hello\nworld\n"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, ok := readLine(ctx, buffer)
	assert.Equal(t, false, ok)
	assert.Equal(t, "", got)
}
