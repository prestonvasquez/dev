package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Start a TCP listener on an available port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server listening on", listener.Addr())

	// Start a goroutine to accept a single connection and then close it with a RST.
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Server accept error:", err)
			return
		}
		fmt.Println("Server: connection accepted, now closing immediately with RST.")
		// If it's a TCP connection, set linger to 0 so that the close sends a RST.
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetLinger(0)
		}
		conn.Close()
	}()

	// Connect to the server as a client.
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Give the server time to accept and force close the connection.
	time.Sleep(200 * time.Millisecond)

	// Attempt to write to the now-closed connection.
	_, err = conn.Write([]byte("Hello, server!"))
	if err != nil {
		// Expect an error like "write: broken pipe" on Unix-like systems.
		fmt.Println("Client write error:", err)
	} else {
		fmt.Println("Client write succeeded unexpectedly")
	}
}
