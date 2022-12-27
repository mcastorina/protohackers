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
		{"abc", "abc"},
		{"", ""},
		{"7hadhjdfhbbbbasjajajasjsjaj", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7hadhjdfhbbbbasjajajasjsjajsjahadhjdfhbbbbasjajajasjsjajsja", "7hadhjdfhbbbbasjajajasjsjajsjahadhjdfhbbbbasjajajasjsjajsja"},
		{"x 7hadhjdfhbbbbasjajajasjsjaj ", "x 7YWHMfk9JZe0LM0g1ZauHuiSxhI "},
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7YWHMfk9JZe0LM0g1ZauHuiSxhI", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"This is a product ID, not a Boguscoin: 7YWHMfk9JZe0LM0g1ZauHuiSxhI-GYUyb6MEuGSsxsdUO79xLjNMH8e-1234", "This is a product ID, not a Boguscoin: 7YWHMfk9JZe0LM0g1ZauHuiSxhI-GYUyb6MEuGSsxsdUO79xLjNMH8e-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			updatedMsg := updateMessage(tt.message)
			assert.Equal(t, tt.expected, updatedMsg)
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
