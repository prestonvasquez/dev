package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
)

func main() {
	rawBSON := []byte{0x03, 0x00, 0x00, 0x00, 0xEA, 0x01, 0x02, 0x03}

	bytesReader := bytes.NewReader(rawBSON)
	reader := bufio.NewReader(bytesReader)

	// Read bytes using wrapper:
	bytes, err := readBytesFull(reader, 4)
	if err != nil {
		log.Fatalf("failed to read bytes from wrapper: %v\n", err)
	}

	log.Printf("readBytes result: %v\n", bytes)

	// Read bytes up to delim
	//var delim byte = 0x00
	var delim byte = 0x02

	bytes, err = reader.ReadBytes(delim)
	if err != nil {
		log.Fatalf("failed to read bytes: %v", err)
	}

	log.Printf("reader.ReadBytes result: %v\n", bytes)

	// Read the size of the buffer
	log.Printf("reader.Size result: %v\n", reader.Size())
}

func readBytesFull(r *bufio.Reader, length int32) ([]byte, error) {
	if length < 0 {
		return nil, fmt.Errorf("invalid length: %d", length)
	}

	buf := make([]byte, length)
	_, err := io.ReadFull(r, buf)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, io.EOF
	} else if err != nil {
		return nil, err
	}

	return buf, nil
}

func readBytesIterative(r *bufio.Reader, length int32) ([]byte, error) {
	if length < 0 {
		return nil, fmt.Errorf("invalid length: %d", length)
	}

	bytes := make([]byte, length)
	for i := 0; i < int(length); i++ {
		var err error

		bytes[i], err = r.ReadByte()
		if err != nil {
			return nil, err
		}
	}

	return bytes, nil
}
