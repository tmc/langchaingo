package mongovector

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/averikitsch/langchaingo/schema"
	"github.com/averikitsch/langchaingo/vectorstores"
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

// mockDocuments will add the given documents to the embedder, assigning each
// a vector such that similarity score = 0.5 * ( 1 + vector * queryVector).
func (emb *mockEmbedder) mockDocuments(doc ...schema.Document) {
	for _, d := range doc {
		emb.docs[d.PageContent] = d
	}
}

// existingVectors returns all the vectors that have been added to the embedder.
// The query vector is included in the list to maintain orthogonality.
func (emb *mockEmbedder) existingVectors() [][]float32 {
	vectors := make([][]float32, 0, len(emb.docs)+1)
	for _, vec := range emb.docVectors {
		vectors = append(vectors, vec)
	}

	return append(vectors, emb.queryVector)
}

// EmbedDocuments will return the embedded vectors for the given texts. If the
// text does not exist in the document set, a zero vector will be returned.
func (emb *mockEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i := range vectors {
		// If the text does not exist in the document set, return a zero vector.
		doc, ok := emb.docs[texts[i]]
		if !ok {
			vectors[i] = make([]float32, len(emb.queryVector))
		}

		// If the vector exists, use it.
		existing, ok := emb.docVectors[texts[i]]
		if ok {
			vectors[i] = existing

			continue
		}

		// If it does not exist, make a linearly independent vector.
		newVectorBasis := newOrthogonalVector(len(emb.queryVector), emb.existingVectors()...)

		// Update the newVector to be scaled by the score.
		newVector := dotProductNormFn(doc.Score, emb.queryVector, newVectorBasis)

		vectors[i] = newVector
		emb.docVectors[texts[i]] = newVector
	}

	return vectors, nil
}

// EmbedQuery returns the query vector.
func (emb *mockEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	return emb.queryVector, nil
}

// Insert all of the mock documents collected by the embedder.
func flushMockDocuments(ctx context.Context, store Store, emb *mockEmbedder) error {
	docs := make([]schema.Document, 0, len(emb.docs))
	for _, doc := range emb.docs {
		docs = append(docs, doc)
	}

	_, err := store.AddDocuments(ctx, docs, vectorstores.WithEmbedder(emb))
	if err != nil {
		return err
	}

	// Consistency on indexes is not synchronous.
	// nolint:mnd
	time.Sleep(10 * time.Second)

	return nil
}

func newNormalizedFloat32() (float32, error) {
	maxInt := big.NewInt(1 << 24)
	n, err := rand.Int(rand.Reader, maxInt)
	if err != nil {
		return 0.0, fmt.Errorf("failed to normalize float32")
	}
	return 2.0*(float32(n.Int64())/float32(maxInt.Int64())) - 1.0, nil
}

// dotProduct will return the dot product between two slices of f32.
func dotProduct(v1, v2 []float32) float32 {
	var sum float32

	for i := range v1 {
		sum += v1[i] * v2[i]
	}

	return sum
}

// linearlyIndependent true if the vectors are linearly independent.
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

// Create a vector of values between [-1, 1] of the specified size.
func newNormalizedVector(dim int) []float32 {
	vector := make([]float32, dim)
	for i := range vector {
		vector[i], _ = newNormalizedFloat32()
	}

	return vector
}

// Use Gram Schmidt to return a vector orthogonal to the basis, so long as
// the vectors in the basis are linearly independent.
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

// return a new vector such that v1 * v2 = 2S - 1.
func dotProductNormFn(score float32, qvector, basis []float32) []float32 {
	var sum float32

	// Populate v2 upto dim-1.
	for i := range qvector[:len(qvector)-1] {
		sum += qvector[i] * basis[i]
	}

	// Calculate v_{2, dim} such that v1 * v2 = 2S - 1:
	basis[len(basis)-1] = (2*score - 1 - sum) / qvector[len(qvector)-1]

	// If the vectors are linearly independent, regenerate the dim-1 elements
	// of v2.
	if !linearlyIndependent(qvector, basis) {
		return dotProductNormFn(score, qvector, basis)
	}

	return basis
}
