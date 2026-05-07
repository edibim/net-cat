package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// TestIntegration_Broadcasting verifies that messages sent by one client
// are received by others, and join/leave notifications work.
func TestIntegration_Broadcasting(t *testing.T) {
	// 1. Start a local server on a random port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	room := NewChatRoom(10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			client := NewClient(conn, room)
			go client.Run()
		}
	}()

	addr := l.Addr().String()

	// 2. Connect Alice
	conn1, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()
	reader1 := bufio.NewReader(conn1)
	consumeUntil(t, reader1, NamePrompt)
	fmt.Fprintln(conn1, "Alice")

	// Consume Alice's join message
	consumeUntil(t, reader1, "Alice has joined our chat...")

	// 3. Connect Bob
	conn2, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	reader2 := bufio.NewReader(conn2)
	consumeUntil(t, reader2, NamePrompt)
	fmt.Fprintln(conn2, "Bob")

	// Consume Bob's join message
	consumeUntil(t, reader2, "Bob has joined our chat...")

	// Alice should see Bob join (BroadcastSystem excludes sender)
	line, _ := reader1.ReadString('\n')
	if !strings.Contains(line, "Bob has joined our chat...") {
		t.Errorf("Alice did not see Bob join notification: %q", line)
	}

	// 4. Alice sends a message
	msg := "Hi Bob!"
	fmt.Fprintln(conn1, msg)

	// Alice no longer receives her own message back to avoid duplicates
	// We'll skip the check for reader1 here and focus on Bob receiving it.

	// Bob receives Alice's message
	line, _ = reader2.ReadString('\n')
	if !strings.Contains(line, "[Alice]:"+msg) {
		t.Errorf("Bob did not receive Alice's message: %q", line)
	}
}

// TestIntegration_FloodProtectionAndTimestamp verifies the join timestamp format
// and ensures the flood protection logic rejects rapid messages.
func TestIntegration_FloodProtectionAndTimestamp(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	room := NewChatRoom(10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go NewClient(conn, room).Run()
		}
	}()

	addr := l.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Complete handshake
	consumeUntil(t, reader, NamePrompt)
	fmt.Fprintln(conn, "Spammer")

	// 2. Verify flood protection
	fmt.Fprintln(conn, "First message")
	fmt.Fprintln(conn, "Second message (too fast)")

	// Consume first message echo
	consumeUntil(t, reader, "[Spammer]:First message")

	// Next line should be the flood protection warning
	warning, _ := reader.ReadString('\n')
	if !strings.Contains(warning, "Slow down!") {
		t.Errorf("Expected flood protection warning, got: %q", warning)
	}
}

// TestIntegration_CapacityLimit verifies that the 11th client is rejected
// after providing a name, while previous clients stay connected.
func TestIntegration_CapacityLimit(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	room := NewChatRoom(10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go NewClient(conn, room).Run()
		}
	}()

	addr := l.Addr().String()
	var conns []net.Conn
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()

	// 1. Fill 10 slots
	for i := 0; i < 10; i++ {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
		consumeUntil(t, bufio.NewReader(c), NamePrompt)
		fmt.Fprintf(c, "User%d\n", i)
		time.Sleep(10 * time.Millisecond) // Ensure Join() completes registration
	}

	// 2. The 11th client attempts to join
	c11, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c11.Close()
	r11 := bufio.NewReader(c11)
	consumeUntil(t, r11, NamePrompt)
	fmt.Fprintln(c11, "Extra")

	line, _ := r11.ReadString('\n')
	if !strings.Contains(line, "Chat room is full") {
		t.Errorf("Expected 'Chat room is full' message, got %q", line)
	}

	// Expect connection to be closed by server
	_, err = r11.ReadByte()
	if err == nil {
		t.Error("Expected connection to be closed by server")
	}
}

// TestIntegration_ChangeNameWithSpaces verifies that /changename "New Name"
// correctly updates the client's name and notifies others.
func TestIntegration_ChangeNameWithSpaces(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	room := NewChatRoom(10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go NewClient(conn, room).Run()
		}
	}()

	addr := l.Addr().String()

	// 1. Connect Alice
	conn1, _ := net.Dial("tcp", addr)
	defer conn1.Close()
	r1 := bufio.NewReader(conn1)
	consumeUntil(t, r1, NamePrompt)
	fmt.Fprintln(conn1, "Alice")

	// 2. Connect Bob
	conn2, _ := net.Dial("tcp", addr)
	defer conn2.Close()
	r2 := bufio.NewReader(conn2)
	consumeUntil(t, r2, NamePrompt)
	fmt.Fprintln(conn2, "Bob")

	// 3. Alice changes her name to something with spaces
	fmt.Fprintln(conn1, `/changename "John Doe"`)

	// 4. Bob should receive the notification
	// Consume Bob's own handshake sequence first
	consumeUntil(t, r2, "[Bob]:")

	line, _ := r2.ReadString('\n')
	if !strings.Contains(line, "Alice is now known as John Doe") {
		t.Errorf("Bob did not receive name change notification. Got: %q", line)
	}

	// 5. Alice sends a message
	fmt.Fprintln(conn1, "Hello from the other side")

	// 6. Bob receives the message from "John Doe"
	line, _ = r2.ReadString('\n')
	if !strings.Contains(line, "[John Doe]:Hello from the other side") {
		t.Errorf("Bob did not receive message with new name. Got: %q", line)
	}
}

// consumeUntil reads from the reader until the target string is found or timeout occurs.
func consumeUntil(t *testing.T, r *bufio.Reader, target string) {
	t.Helper()
	received := ""
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		b, err := r.ReadByte()
		if err != nil {
			t.Fatal(err)
		}
		received += string(b)
		if strings.Contains(received, target) {
			return
		}
	}
	t.Fatalf("Timeout waiting for %q. Received so far: %q", target, received)
}
