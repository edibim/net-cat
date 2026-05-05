package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Client represents a single connected user.
type Client struct {
	name            string
	conn            net.Conn
	room            *ChatRoom
	lastMessageTime time.Time
	floodCount      int
	muteUntil       time.Time
}

// NewClient initializes a new client session.
func NewClient(conn net.Conn, room *ChatRoom) *Client {
	return &Client{
		conn: conn,
		room: room,
	}
}

// Run handles the client session lifecycle.
func (c *Client) Run() {
	defer c.conn.Close()

	// 1. Send the welcome banner
	c.Write(WelcomeBanner)

	// 2. Prompt for name
	reader := bufio.NewReader(c.conn)
	name, err := c.promptName(reader)
	if err != nil {
		return // Connection closed or error reading name
	}
	c.name = name

	// Show join timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(c.conn, "[%s]\n", timestamp)

	// 3. Room capacity check and registration
	history, err := c.room.Join(c)
	if err != nil {
		c.Write(err.Error() + ". Try again later.\n")
		return
	}
	defer c.room.Leave(c)

	// 4. Send history and join notification
	for _, msg := range history {
		c.Write(msg + "\n")
	}
	c.room.BroadcastSystem(fmt.Sprintf("%s has joined our chat...", c.name), c, false)

	// 5. Message loop
	for {
		// Display prompt with current timestamp and client name
		now := time.Now()
		fmt.Fprintf(c.conn, "[%s][%s]:", now.Format("2006-01-02 15:04:05"), c.name)

		line, err := reader.ReadString('\n')
		if err != nil {
			break // Connection lost
		}

		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}

		now = time.Now()

		// Check if client is currently muted
		if now.Before(c.muteUntil) {
			remaining := int(time.Until(c.muteUntil).Seconds())
			fmt.Fprintf(c.conn, "You are muted for spamming. Please wait %ds.\n", remaining)
			continue
		}

		// Flood protection: limit to 1 message per 500ms.
		// We update lastMessageTime on every attempt to penalize rapid retry.
		delta := now.Sub(c.lastMessageTime)
		c.lastMessageTime = now

		if delta < 500*time.Millisecond {
			c.floodCount++
			if c.floodCount >= 5 {
				c.muteUntil = now.Add(time.Minute)
				c.floodCount = 0
				c.Write("\nExceeded spam limit. You are muted for 1 minute.\n")
			} else {
				c.Write("Slow down! Messages sent too fast are blocked.\n")
			}
			continue
		}
		c.floodCount = 0 // Reset count on a successful message

		// Bonus: Command Parsing
		if strings.HasPrefix(text, "/") {
			c.handleCommand(text)
			continue
		}

		c.room.BroadcastMessage(c, text)
	}

	// 6. Leave notification (handled by defer Leave which follows)
	c.room.BroadcastSystem(fmt.Sprintf("%s has left our chat...", c.name), c, false)
}

// handleCommand processes slash commands like /nick or /help.
func (c *Client) handleCommand(text string) {
	cmd := strings.SplitN(text, " ", 2)[0]

	switch cmd {
	case "/changename":
		firstQuote := strings.Index(text, "\"")
		lastQuote := strings.LastIndex(text, "\"")
		if firstQuote == -1 || lastQuote == -1 || firstQuote == lastQuote {
			c.Write("Usage: /changename \"new_name\"\n")
			return
		}

		newName := strings.TrimSpace(text[firstQuote+1 : lastQuote])
		if newName == "" {
			c.Write("Name cannot be empty.\n")
			return
		}

		oldName := c.name
		c.name = newName
		log.Printf("Name change: %s -> %s", oldName, newName)
		c.room.BroadcastSystem(fmt.Sprintf("%s is now known as %s", oldName, newName), c, true)
	case "/list":
		names := c.room.GetClientNames()
		c.Write("Connected users: " + strings.Join(names, ", ") + "\n")
	case "/help":
		c.Write("Available commands: /changename \"name\", /list, /help\n")
	default:
		c.Write(fmt.Sprintf("Unknown command: %s. Type /help for a list of commands.\n", cmd))
	}
}

// promptName repeatedly asks the user for a name until a valid one is provided.
func (c *Client) promptName(reader *bufio.Reader) (string, error) {
	for {
		c.Write(NamePrompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		name := strings.TrimSpace(line)
		if name != "" {
			return name, nil
		}

		c.Write(EmptyNameMsg)
	}
}

// Write sends a string message to the client's connection.
func (c *Client) Write(message string) {
	fmt.Fprint(c.conn, message)
}
