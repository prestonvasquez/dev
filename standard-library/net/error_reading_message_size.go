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

	// Try to read a 4-byte message length
	var msgSize uint32
	err = binary.Read(conn, binary.BigEndian, &msgSize)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Read error: connection closed before full message size was received (EOF)")
		} else {
			fmt.Println("Read error:", err)
		}
	} else {
		fmt.Println("Message size received:", msgSize)
	}
}
