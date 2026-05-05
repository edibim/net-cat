package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const (
	defaultPort  = "8989"
	usageMessage = "[USAGE]: ./TCPChat $port"
)

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

	room := NewChatRoom(10)

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
