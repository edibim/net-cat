package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

func TestPromptName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid name",
			input:    "Yenlik\n",
			expected: "Yenlik",
		},
		{
			name:     "Name with spaces",
			input:    "  Lee  \n",
			expected: "Lee",
		},
		{
			name:     "Empty then valid",
			input:    "\n  \nJohn\n",
			expected: "John",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// net.Pipe provides a synchronous in-memory pipe
			serverConn, clientConn := net.Pipe()
			defer serverConn.Close()
			defer clientConn.Close()

			room := NewChatRoom(10)
			client := NewClient(serverConn, room)
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := client.promptName(reader)
			if err != nil || got != tt.expected {
				t.Errorf("promptName() = %q, %v; want %q, nil", got, err, tt.expected)
			}
		})
	}
}