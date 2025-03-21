package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Start the TCP server
	go startServer()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Client connects
	clientConn, err := net.Dial("tcp", "127.0.0.1:9001")
	if err != nil {
		fmt.Println("Client failed to connect:", err)
		return
	}

	// Read 1 byte to check if the connection is already closed
	var buf [1]byte
	_, err = clientConn.Read(buf[:])
	if err != nil {
		fmt.Println("Read failed (connection likely closed):", err)
	}

	// Now try setting the deadline (should fail)
	err = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		fmt.Println("Failed to set read deadline:", err)
	} else {
		fmt.Println("Successfully set read deadline (unexpected)")
	}

	clientConn.Close()
}

func startServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:9001")
	if err != nil {
		fmt.Println("Server failed to start:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Server accept error:", err)
			return
		}

		fmt.Println("Server: Accepted connection, now closing it immediately")
		conn.Close()
		break // Exit after first connection
	}
}
