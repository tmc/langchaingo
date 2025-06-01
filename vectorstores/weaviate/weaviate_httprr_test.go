package weaviate

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// MockEmbedder for testing.
type mockEmbedder struct{}

var _ embeddings.EmbedderClient = &mockEmbedder{}

func (m *mockEmbedder) CreateEmbedding(_ context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		// Create a simple embedding based on text length
		embeddings[i] = []float32{float32(len(texts[i])), 0.5, 0.3}
	}
	return embeddings, nil
}

func scrubWeaviateData(req *http.Request) error {
	// Scrub API key from Authorization header if present
	if auth := req.Header.Get("Authorization"); auth != "" && auth != "Bearer test-api-key" {
		req.Header.Set("Authorization", "Bearer test-api-key")
	}
	return nil
}

func TestWeaviateHTTPRR_AddDocuments(t *testing.T) {
	ctx := context.Background()

	// Setup HTTP record/replay
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "WEAVIATE_HOST", "WEAVIATE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	rr.ScrubReq(scrubWeaviateData)

	// Create connection client with httprr transport
	connectionClient := &http.Client{
		Transport: rr,
	}

	host := os.Getenv("WEAVIATE_HOST")
	if host == "" {
		host = "localhost:8080"
	}

	scheme := os.Getenv("WEAVIATE_SCHEME")
	if scheme == "" {
		scheme = "http"
	}

	apiKey := os.Getenv("WEAVIATE_API_KEY")
	if apiKey == "" && rr.Recording() {
		t.Skip("WEAVIATE_API_KEY not set")
	}
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	e, err := embeddings.NewEmbedder(&mockEmbedder{})
	require.NoError(t, err)

	// Create store with test embedder
	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithIndexName("TestDocuments"),
		WithAPIKey(apiKey),
		WithConnectionClient(connectionClient),
	)
	require.NoError(t, err)

	// Add documents
	docs := []schema.Document{
		{
			PageContent: "Tokyo is the capital of Japan",
			Metadata: map[string]any{
				"city":    "Tokyo",
				"country": "Japan",
			},
		},
		{
			PageContent: "Paris is the capital of France",
			Metadata: map[string]any{
				"city":    "Paris",
				"country": "France",
			},
		},
	}

	ids, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)
	require.Len(t, ids, 2)
}

func TestWeaviateHTTPRR_SimilaritySearch(t *testing.T) {
	ctx := context.Background()

	// Setup HTTP record/replay
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "WEAVIATE_HOST", "WEAVIATE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	rr.ScrubReq(scrubWeaviateData)

	// Create connection client with httprr transport
	connectionClient := &http.Client{
		Transport: rr,
	}

	host := os.Getenv("WEAVIATE_HOST")
	if host == "" {
		host = "localhost:8080"
	}

	scheme := os.Getenv("WEAVIATE_SCHEME")
	if scheme == "" {
		scheme = "http"
	}

	apiKey := os.Getenv("WEAVIATE_API_KEY")
	if apiKey == "" && rr.Recording() {
		t.Skip("WEAVIATE_API_KEY not set")
	}
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	e, err := embeddings.NewEmbedder(&mockEmbedder{})
	require.NoError(t, err)

	// Create store
	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithIndexName("TestDocuments"),
		WithAPIKey(apiKey),
		WithConnectionClient(connectionClient),
		WithQueryAttrs([]string{"city", "country"}),
	)
	require.NoError(t, err)

	// Perform similarity search
	results, err := store.SimilaritySearch(ctx, "Japan", 2)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 0)
}

func TestWeaviateHTTPRR_SimilaritySearchWithAuth(t *testing.T) {
	ctx := context.Background()

	// Setup HTTP record/replay
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "WEAVIATE_HOST", "WEAVIATE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	rr.ScrubReq(scrubWeaviateData)

	// Create connection client with httprr transport
	connectionClient := &http.Client{
		Transport: rr,
	}

	host := os.Getenv("WEAVIATE_HOST")
	if host == "" {
		host = "localhost:8080"
	}

	scheme := os.Getenv("WEAVIATE_SCHEME")
	if scheme == "" {
		scheme = "http"
	}

	apiKey := os.Getenv("WEAVIATE_API_KEY")
	if apiKey == "" && rr.Recording() {
		t.Skip("WEAVIATE_API_KEY not set")
	}
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	e, err := embeddings.NewEmbedder(&mockEmbedder{})
	require.NoError(t, err)

	// Create store with API key
	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithIndexName("TestDocuments"),
		WithAPIKey(apiKey),
		WithConnectionClient(connectionClient),
	)
	require.NoError(t, err)

	// Perform similarity search
	results, err := store.SimilaritySearch(ctx, "test query", 5)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 0)
}

func TestWeaviateHTTPRR_WithQueryAttrs(t *testing.T) {
	ctx := context.Background()

	// Setup HTTP record/replay
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "WEAVIATE_HOST", "WEAVIATE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	rr.ScrubReq(scrubWeaviateData)

	// Create connection client with httprr transport
	connectionClient := &http.Client{
		Transport: rr,
	}

	host := os.Getenv("WEAVIATE_HOST")
	if host == "" {
		host = "localhost:8080"
	}

	scheme := os.Getenv("WEAVIATE_SCHEME")
	if scheme == "" {
		scheme = "http"
	}

	e, err := embeddings.NewEmbedder(&mockEmbedder{})
	require.NoError(t, err)

	// Create store with query attributes
	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithIndexName("TestDocuments"),
		WithConnectionClient(connectionClient),
		WithQueryAttrs([]string{"title", "author", "category"}),
		WithAdditionalFields([]string{"id", "certainty"}),
	)
	require.NoError(t, err)

	// Perform similarity search
	results, err := store.SimilaritySearch(ctx, "test", 3,
		vectorstores.WithScoreThreshold(0.7))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 0)
}
