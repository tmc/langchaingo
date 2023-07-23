package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombineVectors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		vectors  [][]float64
		weights  []int
		expected []float64
	}{
		{
			vectors:  [][]float64{{10, 5, 10, -3}, {14, 12, 2, 3}},
			weights:  []int{4, 6},
			expected: []float64{0.7605787665953052, 0.5643003752158716, 0.3189523859915796, 0.0368021983836438},
		},
	}

	for _, tc := range cases {
		combined, err := CombineVectors(tc.vectors, tc.weights)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, combined)
	}
}

func TestGetAverage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		vectors  [][]float64
		weights  []int
		expected []float64
	}{
		{
			vectors:  [][]float64{{10, 5, 10}, {-10, -10, 20}},
			weights:  []int{1, 1},
			expected: []float64{0, -2.5, 15},
		},
		{
			vectors:  [][]float64{{10, 5, 10}, {-10, -10, 20}},
			weights:  []int{9, 1},
			expected: []float64{8, 3.5, 11},
		},
		{
			vectors:  [][]float64{{79, 26, -2}, {4, 78, 23}},
			weights:  []int{3, 7},
			expected: []float64{26.5, 62.4, 15.5},
		},
	}

	for _, tc := range cases {
		average, err := getAverage(tc.vectors, tc.weights)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, average)
	}
}

func TestGetNorm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		vector   []float64
		expected float64
	}{
		{
			vector:   []float64{5, 5, 5, 5},
			expected: 10.0,
		},
		{
			vector:   []float64{3, 4},
			expected: 5.0,
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, getNorm(tc.vector))
	}
}
