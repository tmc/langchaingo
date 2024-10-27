package embeddings

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/starmvp/langchaingo/llms/googleai/palm"
)

func newVertexEmbedder(t *testing.T, opts ...Option) *EmbedderImpl {
	t.Helper()
	if gcpProjectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); gcpProjectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT not set")
		return nil
	}

	llm, err := palm.New()
	require.NoError(t, err)

	embedder, err := NewEmbedder(llm, opts...)
	require.NoError(t, err)

	return embedder
}

func TestVertexAIPaLMEmbeddings(t *testing.T) {
	t.Parallel()
	e := newVertexEmbedder(t)

	_, err := e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

func TestVertexAIPaLMEmbeddingsWithOptions(t *testing.T) {
	t.Parallel()
	e := newVertexEmbedder(t, WithBatchSize(5), WithStripNewLines(false))

	_, err := e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}
