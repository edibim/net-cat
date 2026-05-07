package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

const (
	defaultPort  = "8989"
	usageMessage = "[USAGE]: ./TCPChat $port"
)

var (
	rooms   = make(map[string]*ChatRoom)
	roomsMu sync.Mutex
)

func GetOrCreateRoom(name string) *ChatRoom {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	if r, ok := rooms[name]; ok {
		return r
	}
	r := NewChatRoom(name, 10)
	rooms[name] = r
	return r
}

func main() {
	port, err := parsePort(os.Args[1:])
	if err != nil {
		fmt.Println(usageMessage)
		return
	}

	// Bonus: Server Logging to File
	f, err := os.OpenFile("chat.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		defer f.Close()
		// Log to both Stdout and the file
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	log.Printf("Listening on the port :%s\n", port)

	// Pre-initialize 10 specific rooms requested
	roomNames := []string{"main room", "lobby", "advertise room", "gaming", "news", "music", "tech", "random", "help", "staff"}
	for _, name := range roomNames {
		GetOrCreateRoom(name)
	}

	// Default entry room (matches the first in our standard list)
	room := GetOrCreateRoom("main room")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		client := NewClient(conn, room)
		go client.Run()
	}
}

func parsePort(args []string) (string, error) {
	if len(args) == 0 {
		return defaultPort, nil
	}
	if len(args) == 1 && args[0] != "" {
		return args[0], nil
	}
	return "", fmt.Errorf("invalid usage")
}
