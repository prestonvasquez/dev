package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkReadBytesFull(b *testing.B) {
	testData := bytes.Repeat([]byte("a"), 1024)
	reader := bufio.NewReader(bytes.NewReader(testData))

	readLength := int32(512)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := readBytesIterative(reader, readLength)
		if err != nil {
			b.Fatal("Failed to read bytes: ", err)
		}

		reader.Reset(bytes.NewReader(testData))
	}
}

func TestPeek_EOF(t *testing.T) {
	testData := bytes.Repeat([]byte("a"), 1024)
	reader := bufio.NewReader(bytes.NewReader(testData))

	// This should cause an EOF error.
	b, err := reader.Peek(10000)
	assert.Error(t, err)
	assert.Equal(t, 1024, len(b), "Expected to read all bytes before EOF")
}

func TestDiscard_EOF(t *testing.T) {
	testData := bytes.Repeat([]byte("a"), 1024)
	reader := bufio.NewReader(bytes.NewReader(testData))

	// First just discard a small amount.
	_, err := reader.Discard(100)
	assert.NoError(t, err, "Expected to discard 100 bytes without error")

	// Then discard a larger amount that exceeds the buffer size.
	n, err := reader.Discard(10000)
	assert.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
	assert.NotEqual(t, err, io.ErrUnexpectedEOF)
	assert.Equal(t, 1024-100, n, "Expected to discard all bytes before EOF")

	// Peek to see if we can still read the remaining bytes.
	b, err := reader.Peek(100)
	assert.Error(t, err)
	assert.Equal(t, 0, len(b), "Expected to read remaining bytes after discard")
}

func TestReadByte_EOF(t *testing.T) {
	testData := bytes.Repeat([]byte("a"), 1024)
	reader := bufio.NewReader(bytes.NewReader(testData))

	reader.Discard(1024)

	b, err := reader.ReadByte()
	assert.Error(t, err, "Expected an error when reading byte after EOF")
	assert.Equal(t, byte(0), b, "Expected byte to be zero after EOF")
}

func TestReadFull_EOF_LT(t *testing.T) {
	reader := bufio.NewReader(bytes.NewReader([]byte{}))

	_, err := io.ReadFull(reader, make([]byte, 10))
	fmt.Println(err)
}

func TestReadFull_EOF(t *testing.T) {
	reader := bufio.NewReader(bytes.NewReader(bytes.Repeat([]byte("a"), 10)))

	_, err := io.ReadFull(reader, make([]byte, 100))
	fmt.Println(err)
}
