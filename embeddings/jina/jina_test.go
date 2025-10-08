package jina

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
)

func TestJina_EmbedDocuments(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "JINA_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	var opts []Option
	opts = append(opts, WithModel("jina-embeddings-v2-base-en"))
	opts = append(opts, WithClient(rr.Client()))

	if rr.Replaying() {
		opts = append(opts, WithAPIKey("test-api-key"))
	}

	embedder, err := NewJina(opts...)
	require.NoError(t, err)

	texts := []string{
		"The quick brown fox jumps over the lazy dog",
		"Machine learning is a subset of artificial intelligence",
		"Natural language processing enables computers to understand human language",
	}

	embeddings, err := embedder.EmbedDocuments(ctx, texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
	assert.NotEmpty(t, embeddings[2])
}

func TestJina_EmbedQuery(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "JINA_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	var opts []Option
	opts = append(opts, WithModel("jina-embeddings-v2-base-en"))
	opts = append(opts, WithClient(rr.Client()))

	if rr.Replaying() {
		opts = append(opts, WithAPIKey("test-api-key"))
	}

	embedder, err := NewJina(opts...)
	require.NoError(t, err)

	query := "What is machine learning?"

	embedding, err := embedder.EmbedQuery(ctx, query)
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
}

func TestJina_WithBatchSize(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "JINA_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	var opts []Option
	opts = append(opts, WithModel("jina-embeddings-v2-base-en"))
	opts = append(opts, WithBatchSize(2))
	opts = append(opts, WithClient(rr.Client()))

	if rr.Replaying() {
		opts = append(opts, WithAPIKey("test-api-key"))
	}

	embedder, err := NewJina(opts...)
	require.NoError(t, err)

	// Create 5 texts to test batching with batch size 2
	texts := []string{
		"Text 1",
		"Text 2",
		"Text 3",
		"Text 4",
		"Text 5",
	}

	embeddings, err := embedder.EmbedDocuments(ctx, texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 5)
	for i, emb := range embeddings {
		assert.NotEmpty(t, emb, "embedding %d should not be empty", i)
	}
}

func TestJina_StripNewLines(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "JINA_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	var opts []Option
	opts = append(opts, WithModel("jina-embeddings-v2-base-en"))
	opts = append(opts, WithStripNewLines(true))
	opts = append(opts, WithClient(rr.Client()))

	if rr.Replaying() {
		opts = append(opts, WithAPIKey("test-api-key"))
	}

	embedder, err := NewJina(opts...)
	require.NoError(t, err)

	query := "Text with\nnew lines\nshould be processed"

	embedding, err := embedder.EmbedQuery(ctx, query)
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
}
