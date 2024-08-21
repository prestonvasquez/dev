package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

func main() {
	buf := bytes.NewReader([]byte("hello world"))

	log.Println("len before read: ", buf.Len())

	var first4 [4]byte

	n, err := io.ReadFull(buf, first4[:])
	if err != nil {
		log.Fatalf("failed to read full: %v", err)
	}

	log.Println("n: ", n)
	log.Println("first4: ", first4)
	log.Println("buf length: ", buf.Len())
}

func readSize(r io.Reader) (int32, error) {
	const wireMessageSizePrefix = 4

	var wmSizeBytes [wireMessageSizePrefix]byte
	if _, err := io.ReadFull(r, wmSizeBytes[:]); err != nil {
		return 0, fmt.Errorf("error reading the message size: %w", err)
	}

	size := (int32(wmSizeBytes[0])) |
		(int32(wmSizeBytes[1]) << 8) |
		(int32(wmSizeBytes[2]) << 16) |
		(int32(wmSizeBytes[3]) << 24)

	if size < 4 {
		return 0, fmt.Errorf("malformed message length: %d", size)
	}

	return size, nil
}
