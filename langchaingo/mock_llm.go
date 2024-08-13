package langchaingo

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/tmc/langchaingo/embeddings"
)

const textEmbedding3SmallSize = 1536

// MockOpenAILLM will create consistent text embeddings mocking the OpenAI
// text-embedding-3-small algorithm.
type MockOpenAILLM struct {
	seen map[string][]float32
}

var _ embeddings.EmbedderClient = &MockOpenAILLM{}

// createVector will create a float32 vector with random elements of size n.
func createVector(n int) ([]float32, error) {
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

// CreateEmbedding will return vector embeddings for the mock OpenAILLM,
// maintaining consitency.
func (emb *MockOpenAILLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	emb.seen = map[string][]float32{}
	if emb.seen == nil {
	}

	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		if f32s := emb.seen[text]; len(f32s) > 0 {
			vectors[i] = f32s

			continue
		}

		f32s, err := createVector(textEmbedding3SmallSize)
		if err != nil {
			return nil, fmt.Errorf("failed to create vectors for mock openAI LLM: %w", err)
		}

		vectors[i] = f32s
		emb.seen[text] = f32s // ensure consistency
	}

	return vectors, nil
}
