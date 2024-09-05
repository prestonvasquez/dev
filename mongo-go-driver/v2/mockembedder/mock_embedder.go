package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

type mockEmbedder struct {
	searchVector  []float32
	queryToVector map[string][]float32
}

func (embdr mockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i := range vectors {
		var ok bool

		vectors[i], _ = embdr.queryToVector[texts[i]]
		if !ok {
			return nil, fmt.Errorf("failed to find a mock vector")
		}
	}

	return vectors, nil
}

func (embdr mockEmbedder) EmbedQuery(ctx context.Context, _ string) ([]float32, error) {
	return embdr.searchVector, nil
}

func dotProduct(v1, v2 []float32) (sum float32) {
	for i := range v1 {
		sum += v1[i] * v2[i]
	}

	return
}

// makeVectors will return two vectors with the provided similarity score.
func makeVectors(dim int, s float32, primary []float32) ([]float32, []float32, error) {
	if s < 0.0 || s > 1.0 {
		return nil, nil, fmt.Errorf("the score must be in [0,1]")
	}

	v1 := primary
	if primary == nil {
		// If primary is not given, seed v1 randomly
		v1 = make([]float32, dim)

		// Seed the first vector
		rand.Seed(uint64(time.Now().UnixNano()))
		for i := range v1 {
			v1[i] = rand.Float32()*2 - 1 // random values in [-1, 1]
		}
	}

	v2 := make([]float32, dim)

	scalingFactor := (2*s - 1) / dotProduct(v1, v1)

	for i := range v2 {
		v2[i] = v1[i] * scalingFactor
	}

	return v1, v2, nil
}
