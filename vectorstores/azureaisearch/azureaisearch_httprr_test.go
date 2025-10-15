package azureaisearch

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/schema"
)

// MockEmbedder is a mock embedder for testing.
type mockEmbedder struct{}

func (m mockEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		// Create a simple embedding based on text length
		embeddings[i] = []float32{float32(len(texts[i])), 0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

func (m mockEmbedder) EmbedQuery(_ context.Context, text string) ([]float32, error) {
	// Create a simple embedding based on text length
	return []float32{float32(len(text)), 0.1, 0.2, 0.3}, nil
}

func TestStoreHTTPRR_CreateIndex(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AZURE_AI_SEARCH_ENDPOINT", "AZURE_AI_SEARCH_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	endpoint := "https://test.search.windows.net"
	apiKey := "test-api-key"
	if envEndpoint := os.Getenv("AZURE_AI_SEARCH_ENDPOINT"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("AZURE_AI_SEARCH_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	store, err := New(
		WithAPIKey(apiKey),
		WithEmbedder(&mockEmbedder{}),
		WithHTTPClient(rr.Client()),
		WithEndpoint(endpoint),
	)
	require.NoError(t, err)

	indexName := "test-index"

	// Create index with default options
	err = store.CreateIndex(ctx, indexName)
	require.NoError(t, err)
}

func TestStoreHTTPRR_AddDocuments(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AZURE_AI_SEARCH_ENDPOINT", "AZURE_AI_SEARCH_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	endpoint := "https://test.search.windows.net"
	apiKey := "test-api-key"
	if envEndpoint := os.Getenv("AZURE_AI_SEARCH_ENDPOINT"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("AZURE_AI_SEARCH_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	store, err := New(
		WithAPIKey(apiKey),
		WithEmbedder(&mockEmbedder{}),
		WithHTTPClient(rr.Client()),
		WithEndpoint(endpoint),
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

func TestStoreHTTPRR_SimilaritySearch(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AZURE_AI_SEARCH_ENDPOINT", "AZURE_AI_SEARCH_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	endpoint := "https://test.search.windows.net"
	apiKey := "test-api-key"
	if envEndpoint := os.Getenv("AZURE_AI_SEARCH_ENDPOINT"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("AZURE_AI_SEARCH_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	store, err := New(
		WithAPIKey(apiKey),
		WithEmbedder(&mockEmbedder{}),
		WithHTTPClient(rr.Client()),
		WithEndpoint(endpoint),
	)
	require.NoError(t, err)

	query := "What is machine learning?"
	numDocuments := 2

	docs, err := store.SimilaritySearch(ctx, query, numDocuments)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(docs), numDocuments)
}

func TestStoreHTTPRR_DeleteIndex(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AZURE_AI_SEARCH_ENDPOINT", "AZURE_AI_SEARCH_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	endpoint := "https://test.search.windows.net"
	apiKey := "test-api-key"
	if envEndpoint := os.Getenv("AZURE_AI_SEARCH_ENDPOINT"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("AZURE_AI_SEARCH_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	store, err := New(
		WithAPIKey(apiKey),
		WithEmbedder(&mockEmbedder{}),
		WithHTTPClient(rr.Client()),
		WithEndpoint(endpoint),
	)
	require.NoError(t, err)

	indexName := "test-index-to-delete"

	err = store.DeleteIndex(ctx, indexName)
	require.NoError(t, err)
}

func TestStoreHTTPRR_ListIndexes(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AZURE_AI_SEARCH_ENDPOINT", "AZURE_AI_SEARCH_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	endpoint := "https://test.search.windows.net"
	apiKey := "test-api-key"
	if envEndpoint := os.Getenv("AZURE_AI_SEARCH_ENDPOINT"); envEndpoint != "" && rr.Recording() {
		endpoint = envEndpoint
	}
	if envKey := os.Getenv("AZURE_AI_SEARCH_API_KEY"); envKey != "" && rr.Recording() {
		apiKey = envKey
	}

	store, err := New(
		WithAPIKey(apiKey),
		WithEmbedder(&mockEmbedder{}),
		WithHTTPClient(rr.Client()),
		WithEndpoint(endpoint),
	)
	require.NoError(t, err)

	var indexes map[string]interface{}
	err = store.ListIndexes(ctx, &indexes)
	require.NoError(t, err)
	assert.NotNil(t, indexes)
}
