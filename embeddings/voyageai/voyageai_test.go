package voyageai

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestVoyageAIEmbeddings(t *testing.T) {
	t.Parallel()

	if voyageaiKey := os.Getenv("VOYAGEAI_API_KEY"); voyageaiKey == "" {
		t.Skip("VOYAGEAI_API_KEY not set")
	}
	e, err := NewVoyageAI()
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

func TestVoyageAIEmbeddingsWithHTTPrr(t *testing.T) {
	t.Parallel()

	if os.Getenv("VOYAGEAI_API_KEY") == "" && os.Getenv("TEST_RECORD") != "true" {
		t.Skip("VOYAGEAI_API_KEY not set")
	}

	// Enable recording mode with TEST_RECORD=true
	mode := httprr.ModeReplay
	if os.Getenv("TEST_RECORD") == "true" {
		mode = httprr.ModeRecord
	}

	httpClient := httprr.Client("testdata/voyageai_embeddings.json", mode)
	
	e, err := NewVoyageAI(WithHTTPClient(httpClient))
	require.NoError(t, err)

	// Test single query embedding
	queryEmbedding, err := e.EmbedQuery(context.Background(), "What is machine learning?")
	require.NoError(t, err)
	assert.Greater(t, len(queryEmbedding), 0)

	// Test batch document embeddings
	docs := []string{
		"Machine learning is a subset of artificial intelligence.",
		"Neural networks are inspired by biological neurons.",
		"Deep learning uses multiple layers of neural networks.",
	}
	
	embeddings, err := e.EmbedDocuments(context.Background(), docs)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
	
	for i, embedding := range embeddings {
		assert.Greater(t, len(embedding), 0, "Embedding %d should not be empty", i)
	}
}
