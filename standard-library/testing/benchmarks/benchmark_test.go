package main

import (
	"fmt"
	"testing"
)

// Define individual benchmark functions
func BenchmarkExample1(b *testing.B) {
	fmt.Println("meep")
	b.Fatalf("meep")
	for i := 0; i < b.N; i++ {
		// Benchmark code for Example1
	}
}

//	func TestMain(m *testing.M) {
//		result := testing.Benchmark(BenchmarkExample1)
//
//		fmt.Println(result.String())
//	}
func BenchmarkMeep(b *testing.B) {
	result := testing.Benchmark(BenchmarkExample1)

	fmt.Println(result.String())
}
