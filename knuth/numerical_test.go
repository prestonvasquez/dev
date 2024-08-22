package knuth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCD(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		low, high, want int
	}{
		{
			name: "1 and 1",
			low:  1,
			high: 1,
			want: 1,
		},
		{
			name: "2 and 1",
			low:  2,
			high: 1,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GCD(tt.low, tt.high)

			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkKnuth(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Vary the inputs with each iteration
		gcdAlgorithm1(i+1, b.N-i, 0, 1)
	}
}

func BenchmarkChatGPT(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Vary the inputs with each iteration
		gcdAlgorithm2(i+1, b.N-i)
	}
}

func BenchmarkStein(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Vary the inputs with each iteration
		gcdAlgorithm3(i+1, b.N-i)
	}
}
