package main

import (
	"bufio"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestReplaceBoguscoin(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"abc", "abc"},
		{"", ""},
		{"7hadhjdfhbbbbasjajajasjsjaj", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7hadhjdfhbbbbasjajajasjsjajsjahadhjdfhbbbbasjajajasjsjajsja", "7hadhjdfhbbbbasjajajasjsjajsjahadhjdfhbbbbasjajajasjsjajsja"},
		{"x 7hadhjdfhbbbbasjajajasjsjaj ", "x 7YWHMfk9JZe0LM0g1ZauHuiSxhI "},
		{"7hadhjdfhbbbbasjajajasjsjaj 7hadhjdfhbbbbasjajajasjsjaj", "7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7F1u3wSD5RbOHQmupo9nx4TnhQ", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"7YWHMfk9JZe0LM0g1ZauHuiSxhI", "7YWHMfk9JZe0LM0g1ZauHuiSxhI"},
		{"This is a product ID, not a Boguscoin: 7YWHMfk9JZe0LM0g1ZauHuiSxhI-GYUyb6MEuGSsxsdUO79xLjNMH8e-1234", "This is a product ID, not a Boguscoin: 7YWHMfk9JZe0LM0g1ZauHuiSxhI-GYUyb6MEuGSsxsdUO79xLjNMH8e-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			updatedMsg := replaceBoguscoins(tt.message)
			assert.Equal(t, tt.expected, updatedMsg)
		})
	}
}

func TestProxy(t *testing.T) {
	testProxy := func(input string, mapper func(string) string) string {
		builder := strings.Builder{}
		proxy(
			bufio.NewReader(strings.NewReader(input)),
			bufio.NewWriter(&builder),
			mapper,
		)
		return builder.String()
	}
	identity := func(s string) string { return s }
	assert.Equal(t, "hello\nworld\n", testProxy("hello\nworld\n", identity))
	assert.Equal(t, "hello\n", testProxy("hello\nworld", identity))
	assert.Equal(t, "\nolleh\ndlrow", testProxy("hello\nworld\n", func(s string) string {
		out := make([]rune, len(s))
		for i, r := range s {
			out[len(s)-i-1] = r
		}
		return string(out)
	}))
}
