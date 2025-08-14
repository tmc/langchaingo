package qdrant

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/vectorstores"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbedder is a mock embedder for testing.
type MockEmbedder struct{}

func (m MockEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		// Create a simple embedding based on text length
		embeddings[i] = []float32{float32(len(texts[i])), 0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

func (m MockEmbedder) EmbedQuery(_ context.Context, text string) ([]float32, error) {
	// Create a simple embedding based on text length
	return []float32{float32(len(text)), 0.1, 0.2, 0.3}, nil
}

func TestStore_AddDocuments(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "QDRANT_URL")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	endpoint := "http://localhost:6333"
	apiKey := ""
	collectionName := "test-collection"

	if envEndpoint := os.Getenv("QDRANT_URL"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("QDRANT_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	endpointURL, err := url.Parse(endpoint)
	require.NoError(t, err)

	// Replace httputil.DefaultClient with our recording client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

	store, err := New(
		WithURL(*endpointURL),
		WithAPIKey(apiKey),
		WithCollectionName(collectionName),
		WithEmbedder(&MockEmbedder{}),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{
			PageContent: "The quick brown fox jumps over the lazy dog",
			Metadata: map[string]any{
				"source": "test1",
				"page":   1,
			},
		},
		{
			PageContent: "Machine learning is a subset of artificial intelligence",
			Metadata: map[string]any{
				"source": "test2",
				"page":   2,
			},
		},
	}

	ids, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.NotEmpty(t, ids[0])
	assert.NotEmpty(t, ids[1])
}

func TestStore_SimilaritySearch(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "QDRANT_URL")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	endpoint := "http://localhost:6333"
	apiKey := ""
	collectionName := "test-collection"

	if envEndpoint := os.Getenv("QDRANT_URL"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("QDRANT_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	endpointURL, err := url.Parse(endpoint)
	require.NoError(t, err)

	// Replace httputil.DefaultClient with our recording client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

	store, err := New(
		WithURL(*endpointURL),
		WithAPIKey(apiKey),
		WithCollectionName(collectionName),
		WithEmbedder(&MockEmbedder{}),
	)
	require.NoError(t, err)

	query := "What is machine learning?"
	numDocuments := 2

	docs, err := store.SimilaritySearch(ctx, query, numDocuments)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(docs), numDocuments)
}

func TestStore_SimilaritySearchWithScore(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "QDRANT_URL")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	endpoint := "http://localhost:6333"
	apiKey := ""
	collectionName := "test-collection"

	if envEndpoint := os.Getenv("QDRANT_URL"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("QDRANT_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	endpointURL, err := url.Parse(endpoint)
	require.NoError(t, err)

	// Replace httputil.DefaultClient with our recording client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

	store, err := New(
		WithURL(*endpointURL),
		WithAPIKey(apiKey),
		WithCollectionName(collectionName),
		WithEmbedder(&MockEmbedder{}),
	)
	require.NoError(t, err)

	query := "What is machine learning?"
	numDocuments := 2
	scoreThreshold := float32(0.5)

	docs, err := store.SimilaritySearch(ctx, query, numDocuments,
		vectorstores.WithScoreThreshold(scoreThreshold))
	require.NoError(t, err)
	assert.LessOrEqual(t, len(docs), numDocuments)
}

func TestDoRequest(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "QDRANT_URL")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	endpoint := "http://localhost:6333"
	apiKey := ""

	if envEndpoint := os.Getenv("QDRANT_URL"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("QDRANT_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	// Replace httputil.DefaultClient with our recording client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

	testURL, err := url.Parse(endpoint + "/collections")
	require.NoError(t, err)

	// Test GET request
	body, status, err := DoRequest(ctx, *testURL, apiKey, http.MethodGet, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	defer body.Close()

	// Read response to ensure it's valid
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}
