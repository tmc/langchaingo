package jina

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestJina_EmbedDocuments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Check if we have API key or httprr recording
	if os.Getenv("JINA_API_KEY") == "" {
		testName := httprr.CleanFileName(t.Name())
		httprrFile := filepath.Join("testdata", testName+".httprr")
		httprrGzFile := httprrFile + ".gz"
		if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
			if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
				t.Skip("JINA_API_KEY not set and no httprr recording available")
			}
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("JINA_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	embedder, err := NewJina(
		WithAPIKey(apiKey),
		WithModel("jina-embeddings-v2-base-en"),
	)
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
	t.Parallel()

	// Check if we have API key or httprr recording
	if os.Getenv("JINA_API_KEY") == "" {
		testName := httprr.CleanFileName(t.Name())
		httprrFile := filepath.Join("testdata", testName+".httprr")
		httprrGzFile := httprrFile + ".gz"
		if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
			if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
				t.Skip("JINA_API_KEY not set and no httprr recording available")
			}
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("JINA_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	embedder, err := NewJina(
		WithAPIKey(apiKey),
		WithModel("jina-embeddings-v2-base-en"),
	)
	require.NoError(t, err)

	query := "What is machine learning?"

	embedding, err := embedder.EmbedQuery(ctx, query)
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
}

func TestJina_WithBatchSize(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Check if we have API key or httprr recording
	if os.Getenv("JINA_API_KEY") == "" {
		testName := httprr.CleanFileName(t.Name())
		httprrFile := filepath.Join("testdata", testName+".httprr")
		httprrGzFile := httprrFile + ".gz"
		if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
			if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
				t.Skip("JINA_API_KEY not set and no httprr recording available")
			}
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("JINA_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	embedder, err := NewJina(
		WithAPIKey(apiKey),
		WithModel("jina-embeddings-v2-base-en"),
		WithBatchSize(2),
	)
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
	t.Parallel()

	// Check if we have API key or httprr recording
	if os.Getenv("JINA_API_KEY") == "" {
		testName := httprr.CleanFileName(t.Name())
		httprrFile := filepath.Join("testdata", testName+".httprr")
		httprrGzFile := httprrFile + ".gz"
		if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
			if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
				t.Skip("JINA_API_KEY not set and no httprr recording available")
			}
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("JINA_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	embedder, err := NewJina(
		WithAPIKey(apiKey),
		WithModel("jina-embeddings-v2-base-en"),
		WithStripNewLines(true),
	)
	require.NoError(t, err)

	query := "Text with\nnew lines\nshould be processed"

	embedding, err := embedder.EmbedQuery(ctx, query)
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
}
