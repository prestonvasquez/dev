//go:build dotProduct

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func dotProduct(v1, v2 []float64) (sum float64) {
	for i := range v1 {
		sum += v1[i] * v2[i]
	}

	return
}

func makeVectors(dim int, s float64) ([]float64, []float64, error) {
	if s < 0.0 || s > 1.0 {
		return nil, nil, fmt.Errorf("the score must be in [0,1]")
	}

	v1 := make([]float64, dim)
	v2 := make([]float64, dim)

	// Seed the first vector
	rand.Seed(time.Now().UnixNano())
	for i := range v1 {
		v1[i] = rand.Float64()*2 - 1 // random values in [-1, 1]
	}

	scalingFactor := (2*s - 1) / dotProduct(v1, v1)

	for i := range v2 {
		v2[i] = v1[i] * scalingFactor
	}

	return v1, v2, nil
}

func main() {
	v1, v2, _ := makeVectors(1536, 0.8)

	fmt.Println(dotProduct(v1, v2))
}
