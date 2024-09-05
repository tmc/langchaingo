package mongovector

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type mockEmbedder struct {
	dim           int
	query         string // query that will be used in the search
	queryVector   []float32
	docSet        map[string]float32 // pageContent to expected score
	flushedDocSet map[string][]float32
}

var _ embeddings.Embedder = &mockEmbedder{}

func newMockEmbedder(dim int, query string) *mockEmbedder {
	emb := &mockEmbedder{
		dim:    dim,
		query:  query,
		docSet: make(map[string]float32),
	}

	return emb
}

// Store a document that will be returned by a similarity search on the provided
// query with the specific score.
func (emb *mockEmbedder) addDocument(pageContent string, score float32) error {
	if emb.flushedDocSet != nil {
		return fmt.Errorf("cannot make new queries after flushing")
	}

	emb.docSet[pageContent] = score

	return nil
}

func (emb *mockEmbedder) flush(ctx context.Context, store Store) error {
	// Create a vector for each document such that all vectors are linearly
	// independent. Leave one space at teh end for scaling.
	vectors := makeLinearlyIndependentVectors(len(emb.docSet), emb.dim)

	// Create a linearly independent query vector.
	emb.queryVector = makeOrthogonalVector(emb.dim, vectors...)

	// For each pageContent + score combo, update the corresponding vector with
	// a final element so that it's dot product with queryVector is 2S - 1
	// where S is the desired simlarity score.
	emb.flushedDocSet = make(map[string][]float32)
	docs := []schema.Document{}

	count := 0
	for pageContent, score := range emb.docSet {
		emb.flushedDocSet[pageContent] = makeScoreVector(score, emb.queryVector, vectors[count])
		docs = append(docs, schema.Document{PageContent: pageContent})

		count++
	}

	_, err := store.AddDocuments(ctx, docs, vectorstores.WithEmbedder(emb))
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	// The read consistency for vector search isn't automatic.
	time.Sleep(1 * time.Second)

	return nil
}

func (emb *mockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i := range vectors {
		var ok bool

		vectors[i], ok = emb.flushedDocSet[texts[i]]
		if !ok {
			vectors[i] = makeVector(emb.dim)
		}
	}

	return vectors, nil
}

func (emb *mockEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return emb.queryVector, nil
}

// newNormalizedFloat32 will generate a random float32 in [-1, 1].
func newNormalizedFloat32() (float32, error) {
	max := big.NewInt(1 << 24)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0.0, fmt.Errorf("failed to normalize float32")
	}

	return 2.0*(float32(n.Int64())/float32(1<<24)) - 1.0, nil
}

// dotProduct will return the dot product between two slices of f32.
func dotProduct(v1, v2 []float32) (sum float32) {
	for i := range v1 {
		sum += v1[i] * v2[i]
	}

	return
}

// linearlyIndependent true if the vectors are linearly independent
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

// Update the basis vector such that qvector * basis = 2S - 1.
func makeScoreVector(S float32, qvector []float32, basis []float32) []float32 {
	var sum float32

	// Populate v2 upto dim-1.
	for i := 0; i < len(qvector)-1; i++ {
		sum += qvector[i] * basis[i]
	}

	// Calculate v_{2, dim} such that v1 * v2 = 2S - 1:
	basis[len(basis)-1] = (2*S - 1 - sum) / qvector[len(qvector)-1]

	// If the vectors are linearly independent, regenerate the dim-1 elements
	// of v2.
	if !linearlyIndependent(qvector, basis) {
		return makeScoreVector(S, qvector, basis)
	}

	return basis
}

// makeVector will create a vector of values beween [-1, 1] of the specified
// size.
func makeVector(dim int) []float32 {
	vector := make([]float32, dim)
	for i := range vector {
		vector[i], _ = newNormalizedFloat32()
	}

	return vector
}

// Use Gram Schmidt to return a vector orthogonal to the basis, so long as
// the vectors in the basis are linearly independent.
func makeOrthogonalVector(dim int, basis ...[]float32) []float32 {
	candidate := makeVector(dim)

	for _, b := range basis {
		dp := dotProduct(candidate, b)
		basisNorm := dotProduct(b, b)

		for i := range candidate {
			candidate[i] -= (dp / basisNorm) * b[i]
		}
	}

	return candidate
}

// Make n linearly independent vectors of size dim.
func makeLinearlyIndependentVectors(n int, dim int) [][]float32 {
	vectors := [][]float32{}

	for i := 0; i < n; i++ {
		v := makeOrthogonalVector(dim, vectors...)

		vectors = append(vectors, v)
	}

	return vectors
}
