package main

import (
	"encoding/binary"
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

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			return
		}
		defer conn.Close()

		// Simulate an error: Send only 2 bytes of the expected 4-byte message size
		partialHeader := []byte{0x00, 0x00}
		_, err = conn.Write(partialHeader)
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}

		// Close the connection before sending the full message size
		time.Sleep(1 * time.Second)
		fmt.Println("Server closed connection early")
	}()

	// Connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Use ReadFull to ensure we get all 4 bytes or fail
	buf := make([]byte, 4)
	n, err := io.ReadFull(conn, buf)
	if err != nil {
		fmt.Printf("Read error: only %d bytes read: %v\n", n, err)
	} else {
		msgSize := binary.BigEndian.Uint32(buf)
		fmt.Println("Message size received:", msgSize)
	}
}
