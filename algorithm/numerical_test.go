package algorithm

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
		{
			name: "3 and 5",
			low:  3,
			high: 5,
			want: 1,
		},
		{
			name: "6 and 9",
			low:  6,
			high: 9,
			want: 3,
		},
		{
			name: "15 and 10",
			low:  15,
			high: 10,
			want: 5,
		},
		{
			name: "56 and 15",
			low:  56,
			high: 15,
			want: 1,
		},
		{
			name: "48 and 18",
			low:  48,
			high: 18,
			want: 6,
		},
		{
			name: "100 and 25",
			low:  100,
			high: 25,
			want: 25,
		},
		{
			name: "270 and 192",
			low:  270,
			high: 192,
			want: 6,
		},
		{
			name: "462 and 1071",
			low:  462,
			high: 1071,
			want: 21,
		},
		{
			name: "0 and 5",
			low:  0,
			high: 5,
			want: 5,
		},
		{
			name: "0 and 0",
			low:  0,
			high: 0,
			want: 0, // Depending on how you want to handle the case where both are zero
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
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
