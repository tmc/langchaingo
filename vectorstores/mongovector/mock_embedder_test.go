package mongovector_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
)

type mockEmbedder struct {
	queryVector []float32
	docs        map[string]schema.Document
	docVectors  map[string][]float32
}

var _ embeddings.Embedder = &mockEmbedder{}

func newMockEmbedder(dim int) *mockEmbedder {
	emb := &mockEmbedder{
		queryVector: newNormalizedVector(dim),
		docs:        make(map[string]schema.Document),
		docVectors:  make(map[string][]float32),
	}

	return emb
}

func (emb *mockEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i := range vectors {
		doc, ok := emb.docs[texts[i]]
		if !ok {
			vectors[i] = make([]float32, len(emb.queryVector))
		}

		existing, ok := emb.docVectors[texts[i]]
		if ok {
			vectors[i] = existing
			continue
		}

		newVectorBasis := newOrthogonalVector(len(emb.queryVector), emb.existingVectors()...)

		newVector := dotProductNormFn(doc.Score, emb.queryVector, newVectorBasis)

		vectors[i] = newVector
		emb.docVectors[texts[i]] = newVector
	}

	return vectors, nil
}

func (emb *mockEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	return emb.queryVector, nil
}

func (emb *mockEmbedder) existingVectors() [][]float32 {
	vectors := make([][]float32, 0, len(emb.docs)+1)
	for _, vec := range emb.docVectors {
		vectors = append(vectors, vec)
	}

	return append(vectors, emb.queryVector)
}

// Helper functions (not exported)

func newNormalizedFloat32() (float32, error) {
	max := big.NewInt(1 << 24)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0.0, fmt.Errorf("failed to normalize float32")
	}

	return 2.0*(float32(n.Int64())/float32(1<<24)) - 1.0, nil
}

func dotProduct(v1, v2 []float32) float32 {
	var sum float32

	for i := range v1 {
		sum += v1[i] * v2[i]
	}

	return sum
}

func linearlyIndependent(v1, v2 []float32) bool {
	var ratio float32

	for i := range v1 {
		if v1[i] != 0 {
			r := v2[i] / v1[i]

			if ratio == 0 {
				ratio = r
				continue
			}

			if r == ratio {
				continue
			}

			return true
		}

		if v2[i] != 0 {
			return true
		}
	}

	return false
}

func newNormalizedVector(dim int) []float32 {
	vector := make([]float32, dim)
	for i := range vector {
		vector[i], _ = newNormalizedFloat32()
	}

	return vector
}

func newOrthogonalVector(dim int, basis ...[]float32) []float32 {
	candidate := newNormalizedVector(dim)

	for _, b := range basis {
		dp := dotProduct(candidate, b)
		basisNorm := dotProduct(b, b)

		for i := range candidate {
			candidate[i] -= (dp / basisNorm) * b[i]
		}
	}

	return candidate
}

func dotProductNormFn(score float32, qvector, basis []float32) []float32 {
	var sum float32

	for i := range qvector[:len(qvector)-1] {
		sum += qvector[i] * basis[i]
	}

	basis[len(basis)-1] = (2*score - 1 - sum) / qvector[len(qvector)-1]

	if !linearlyIndependent(qvector, basis) {
		return dotProductNormFn(score, qvector, basis)
	}

	return basis
}