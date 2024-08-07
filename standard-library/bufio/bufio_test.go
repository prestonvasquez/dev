package main

import (
	"bufio"
	"bytes"
	"testing"
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
