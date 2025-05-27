package qdrant

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// MockEmbedder is a mock embedder for testing
type MockEmbedder struct{}

func (m MockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		// Create a simple embedding based on text length
		embeddings[i] = []float32{float32(len(texts[i])), 0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

func (m MockEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	// Create a simple embedding based on text length
	return []float32{float32(len(text)), 0.1, 0.2, 0.3}, nil
}

func TestStore_AddDocuments(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		if req.Header.Get("api-key") != "" {
			req.Header.Set("api-key", "test-api-key")
		}
		return nil
	})

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

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

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

	ids, err := store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.NotEmpty(t, ids[0])
	assert.NotEmpty(t, ids[1])
}

func TestStore_SimilaritySearch(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		if req.Header.Get("api-key") != "" {
			req.Header.Set("api-key", "test-api-key")
		}
		return nil
	})

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

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	store, err := New(
		WithURL(*endpointURL),
		WithAPIKey(apiKey),
		WithCollectionName(collectionName),
		WithEmbedder(&MockEmbedder{}),
	)
	require.NoError(t, err)

	query := "What is machine learning?"
	numDocuments := 2

	docs, err := store.SimilaritySearch(context.Background(), query, numDocuments)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(docs), numDocuments)
}

func TestStore_SimilaritySearchWithScore(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		if req.Header.Get("api-key") != "" {
			req.Header.Set("api-key", "test-api-key")
		}
		return nil
	})

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

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

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

	docs, err := store.SimilaritySearch(context.Background(), query, numDocuments, 
		vectorstores.WithScoreThreshold(scoreThreshold))
	require.NoError(t, err)
	assert.LessOrEqual(t, len(docs), numDocuments)
}

func TestDoRequest(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		if req.Header.Get("api-key") != "" {
			req.Header.Set("api-key", "test-api-key")
		}
		return nil
	})

	endpoint := "http://localhost:6333"
	apiKey := ""
	
	if envEndpoint := os.Getenv("QDRANT_URL"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("QDRANT_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	testURL, err := url.Parse(endpoint + "/collections")
	require.NoError(t, err)

	// Test GET request
	body, status, err := DoRequest(context.Background(), *testURL, apiKey, http.MethodGet, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	defer body.Close()

	// Read response to ensure it's valid
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}