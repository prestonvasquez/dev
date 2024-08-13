package util

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
)

const OpenAITextEmbedding3SmallSize = 1536

// CreateVector will create a float32 vector with random elements of size n.
func CreateVector(n int) ([]float32, error) {
	f32s := make([]float32, n)

	for i := range f32s {
		var b [4]byte

		_, err := rand.Read(b[:])
		if err != nil {
			return nil, fmt.Errorf("failed to read random bytes: %w", err)
		}

		f32s[i] = float32(binary.LittleEndian.Uint32(b[:])) / math.MaxUint32
	}

	return f32s, nil
}
