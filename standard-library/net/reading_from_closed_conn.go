package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Create a TCP listener on a random available port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	// Get the listener's address
	addr := ln.Addr().String()
	fmt.Println("Listening on:", addr)

	// Start a goroutine to accept a connection
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			return
		}
		fmt.Println("Client connected")

		// Close the connection after a short delay
		time.Sleep(1 * time.Second)
		conn.Close()
		fmt.Println("Server closed connection")
	}()

	// Dial the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	conn.Close()

	// Wait for the server to close the connection
	time.Sleep(2 * time.Second)

	// Try to read from the closed connection
	buf := make([]byte, 1024)
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		panic(err)
	}
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err) // Expect "use of closed network connection"
	} else {
		fmt.Println("Read", n, "bytes:", string(buf[:n]))
	}
}
