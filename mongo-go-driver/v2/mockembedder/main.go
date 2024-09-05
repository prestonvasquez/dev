package main

import (
	"context"
	"fmt"
)

func main() {
	const dim = 1536

	helloVector, v9, _ := makeVectors(dim, 0.9, nil)
	_, v17, _ := makeVectors(dim, 0.17, helloVector)
	_, v01, _ := makeVectors(dim, 0.01, helloVector)

	embedder := mockEmbedder{
		searchVector: helloVector,
		queryToVector: map[string][]float32{
			"v9":  v9,
			"v17": v17,
			"v01": v01,
		},
	}

	f32s, _ := embedder.EmbedDocuments(context.Background(), []string{"v9"})

	fmt.Println(f32s)
}
