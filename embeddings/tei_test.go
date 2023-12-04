package embeddings

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/tei"
)

func newTEIEmbedder(t *testing.T, opts ...Option) *EmbedderImpl {
	t.Helper()
	teiURL := os.Getenv("TEI_API_URL")
	if teiURL == "" {
		t.Skip("TEI_API_URL not set")
		return nil
	}

	llm, err := tei.New(
		tei.WithAPIBaseURL(teiURL),
		tei.WithPoolSize(4),
	)
	require.NoError(t, err)
	embedder, err := NewEmbedder(llm, opts...)
	require.NoError(t, err)
	return embedder
}

func TestTEIEmbeddings(t *testing.T) {
	t.Parallel()
	e := newTEIEmbedder(t)
	texts := []string{"Hello world"}
	emb, err := e.EmbedDocuments(context.Background(), texts)
	require.NoError(t, err)
	assert.Len(t, emb, 1)
}

func TestTEIEmbeddingsQueryVsDocuments(t *testing.T) {
	t.Parallel()

	e := newTEIEmbedder(t)
	text := "hi there"
	eq, err := e.EmbedQuery(context.Background(), text)
	require.NoError(t, err)

	eb, err := e.EmbedDocuments(context.Background(), []string{text})
	require.NoError(t, err)

	// Using strict equality should be OK here because we expect the same values
	// for the same string, deterministically.
	assert.Equal(t, eq, eb[0])
}

func TestTEIEmbeddingsWithOptions(t *testing.T) {
	t.Parallel()

	e := newTEIEmbedder(t, WithBatchSize(1), WithStripNewLines(false))

	_, err := e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}
