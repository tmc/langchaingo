package embeddings

import (
	"math"
	"reflect"
	"testing"
)

func TestCombineVectors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		vectors  [][]float32
		weights  []int
		expected []float32
	}{
		{
			name:     "basic combination",
			vectors:  [][]float32{{10, 5, 10, -3}, {14, 12, 2, 3}},
			weights:  []int{4, 6},
			expected: []float32{0.7605787665953052, 0.5643003752158716, 0.3189523859915796, 0.036802202},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			combined, err := CombineVectors(tc.vectors, tc.weights)
			if err != nil {
				t.Fatalf("CombineVectors() error = %v", err)
			}
			if !reflect.DeepEqual(tc.expected, combined) {
				t.Errorf("CombineVectors() = %v, want %v", combined, tc.expected)
			}
		})
	}
}

func TestGetAverage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		vectors  [][]float32
		weights  []int
		expected []float32
	}{
		{
			name:     "equal weights",
			vectors:  [][]float32{{10, 5, 10}, {-10, -10, 20}},
			weights:  []int{1, 1},
			expected: []float32{0, -2.5, 15},
		},
		{
			name:     "unequal weights",
			vectors:  [][]float32{{10, 5, 10}, {-10, -10, 20}},
			weights:  []int{9, 1},
			expected: []float32{8, 3.5, 11},
		},
		{
			name:     "different values",
			vectors:  [][]float32{{79, 26, -2}, {4, 78, 23}},
			weights:  []int{3, 7},
			expected: []float32{26.5, 62.4, 15.5},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			average, err := getAverage(tc.vectors, tc.weights)
			if err != nil {
				t.Fatalf("getAverage() error = %v", err)
			}
			if !reflect.DeepEqual(tc.expected, average) {
				t.Errorf("getAverage() = %v, want %v", average, tc.expected)
			}
		})
	}
}

func TestGetNorm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		vector   []float32
		expected float32
	}{
		{
			name:     "equal components",
			vector:   []float32{5, 5, 5, 5},
			expected: 10.0,
		},
		{
			name:     "3-4-5 triangle",
			vector:   []float32{3, 4},
			expected: 5.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := getNorm(tc.vector)
			delta := math.Abs(float64(tc.expected - got))
			if delta > 0.0001 {
				t.Errorf("getNorm(%v) = %v, want %v (diff: %v)", tc.vector, got, tc.expected, delta)
			}
		})
	}
}
