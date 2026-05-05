package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// ChatRoom owns all shared chat state including connected clients and history.
type ChatRoom struct {
	mu      sync.Mutex
	clients map[*Client]bool
	history []string
	limit   int
}

// NewChatRoom initializes a new chat room with a specific capacity.
func NewChatRoom(limit int) *ChatRoom {
	return &ChatRoom{
		clients: make(map[*Client]bool),
		history: make([]string, 0),
		limit:   limit,
	}
}

// Join adds a client to the room if capacity allows and returns the current message history.
func (r *ChatRoom) Join(c *Client) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.clients) >= r.limit {
		return nil, errors.New("Chat room is full")
	}
	r.clients[c] = true
	log.Printf("Client joined: %s (%s)", c.name, c.conn.RemoteAddr())

	// Return a copy of history to avoid race conditions
	res := make([]string, len(r.history))
	copy(res, r.history)
	return res, nil
}

// Leave removes a client from the room.
func (r *ChatRoom) Leave(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.Printf("Client left: %s (%s)", c.name, c.conn.RemoteAddr())
	delete(r.clients, c)
}

// BroadcastMessage adds a message to history and sends it to all connected clients.
func (r *ChatRoom) BroadcastMessage(sender *Client, text string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf("[%s][%s]:%s", timestamp, sender.name, text)

	log.Println("Broadcast:", formatted)
	r.mu.Lock()
	r.history = append(r.history, formatted)
	targets := r.copyClients()
	r.mu.Unlock()

	for _, c := range targets {
		if c != sender {
			c.Write(formatted + "\n")
		}
	}
}

// BroadcastSystem sends a system event message to clients.
func (r *ChatRoom) BroadcastSystem(text string, except *Client, saveToHistory bool) {
	if saveToHistory {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		text = fmt.Sprintf("[%s] %s", timestamp, text)
		r.mu.Lock()
		r.history = append(r.history, text)
		r.mu.Unlock()
	}

	log.Println("System Message:", text)
	r.mu.Lock()
	targets := r.copyClients()
	r.mu.Unlock()

	for _, c := range targets {
		if c != except {
			c.Write(text + "\n")
		}
	}
}

// GetClientNames returns a list of names of all currently connected clients.
func (r *ChatRoom) GetClientNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]string, 0, len(r.clients))
	for c := range r.clients {
		names = append(names, c.name)
	}
	return names
}

func (r *ChatRoom) copyClients() []*Client {
	targets := make([]*Client, 0, len(r.clients))
	for c := range r.clients {
		targets = append(targets, c)
	}
	return targets
}