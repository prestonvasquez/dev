package main

import (
	"fmt"
	"testing"
)

func main() {
	result := testing.Benchmark(func(b *testing.B) {
		b.Fatalf("meep")
	})

	fmt.Println(result.String())
}
