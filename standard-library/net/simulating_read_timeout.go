package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	// Start a TCP server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	fmt.Println("Listening on:", addr)

	// Start server
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			return
		}
		defer conn.Close()

		fmt.Println("Server accepted connection but will not send data")
		time.Sleep(5 * time.Second) // Simulate a delay (client times out before this)
	}()

	// Connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Set a 1-second read timeout
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	// Try to read 4 bytes (but server isn't sending anything)
	var sizeBuf [4]byte
	_, err = io.ReadFull(conn, sizeBuf[:])
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("Read error: operation timed out")
		} else {
			fmt.Println("Read error:", err)
		}
	} else {
		fmt.Println("Read succeeded (unexpected)")
	}
}
