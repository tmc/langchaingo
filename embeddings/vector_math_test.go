package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombineVectors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		vectors  [][]float32
		weights  []int
		expected []float32
	}{
		{
			vectors:  [][]float32{{10, 5, 10, -3}, {14, 12, 2, 3}},
			weights:  []int{4, 6},
			expected: []float32{0.7605787665953052, 0.5643003752158716, 0.3189523859915796, 0.036802202},
		},
	}

	for _, tc := range cases {
		combined, err := CombineVectors(tc.vectors, tc.weights)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, combined)
	}
}

func TestGetAverage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		vectors  [][]float32
		weights  []int
		expected []float32
	}{
		{
			vectors:  [][]float32{{10, 5, 10}, {-10, -10, 20}},
			weights:  []int{1, 1},
			expected: []float32{0, -2.5, 15},
		},
		{
			vectors:  [][]float32{{10, 5, 10}, {-10, -10, 20}},
			weights:  []int{9, 1},
			expected: []float32{8, 3.5, 11},
		},
		{
			vectors:  [][]float32{{79, 26, -2}, {4, 78, 23}},
			weights:  []int{3, 7},
			expected: []float32{26.5, 62.4, 15.5},
		},
	}

	for _, tc := range cases {
		average, err := getAverage(tc.vectors, tc.weights)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, average)
	}
}

func TestGetNorm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		vector   []float32
		expected float32
	}{
		{
			vector:   []float32{5, 5, 5, 5},
			expected: 10.0,
		},
		{
			vector:   []float32{3, 4},
			expected: 5.0,
		},
	}

	for _, tc := range cases {
		assert.InEpsilon(t, tc.expected, getNorm(tc.vector), 0.0001)
	}
}
