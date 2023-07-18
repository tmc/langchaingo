package embeddings

import (
	"errors"
	"math"
)

var (
	// ErrVectorsNotSameSize is returned if the vectors returned from the
	// embeddings api have different sizes.
	ErrVectorsNotSameSize = errors.New("vectors gotten not the same size")
	// ErrAllTextsLenZero is returned if all texts to be embedded has the combined
	// length of zero.
	ErrAllTextsLenZero = errors.New("all texts have length 0")
)

func CombineVectors(vectors [][]float64, weights []int) ([]float64, error) {
	average, err := getAverage(vectors, weights)
	if err != nil {
		return nil, err
	}

	averageNorm := getNorm(average)
	for i := 0; i < len(average); i++ {
		average[i] /= averageNorm
	}

	return average, nil
}

// getAverage does the following calculation:
//
//	avg = sum(vectors * weights) / sum(weights).
func getAverage(vectors [][]float64, weights []int) ([]float64, error) {
	// Check that all vectors are the same size and get that size.
	vectorLen := -1
	for _, vector := range vectors {
		if vectorLen == -1 {
			vectorLen = len(vector)
			continue
		}

		if len(vector) != vectorLen {
			return nil, ErrVectorsNotSameSize
		}
	}

	if vectorLen == -1 {
		return []float64{}, nil
	}

	// Get the sum of the weights.
	weightSum := 0
	for _, weight := range weights {
		weightSum += weight
	}

	if weightSum == 0 {
		return nil, ErrAllTextsLenZero
	}

	average := make([]float64, vectorLen)
	for i := 0; i < vectorLen; i++ {
		for j := 0; j < len(vectors); j++ {
			average[i] += vectors[j][i] * float64(weights[j])
		}
	}

	for i := 0; i < len(average); i++ {
		average[i] /= float64(weightSum)
	}

	return average, nil
}

func getNorm(v []float64) float64 {
	var sum float64
	for i := 0; i < len(v); i++ {
		sum += v[i] * v[i]
	}

	return math.Sqrt(sum)
}
