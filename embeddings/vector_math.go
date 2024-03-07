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

func CombineVectors(vectors [][]float32, weights []int) ([]float32, error) {
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
func getAverage(vectors [][]float32, weights []int) ([]float32, error) {
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
		return []float32{}, nil
	}

	// Get the sum of the weights.
	weightSum := 0
	for _, weight := range weights {
		weightSum += weight
	}

	if weightSum == 0 {
		return nil, ErrAllTextsLenZero
	}

	average := make([]float32, vectorLen)
	for i := 0; i < vectorLen; i++ {
		for j := 0; j < len(vectors); j++ {
			average[i] += vectors[j][i] * float32(weights[j])
		}
	}

	for i := 0; i < len(average); i++ {
		average[i] /= float32(weightSum)
	}

	return average, nil
}

func getNorm(v []float32) float32 {
	var sum float32
	for i := 0; i < len(v); i++ {
		sum += v[i] * v[i]
	}

	return float32(math.Sqrt(float64(sum)))
}
