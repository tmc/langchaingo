package embeddings

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/googleai/palm"
)

func newVertexEmbedder(t *testing.T, rr *httprr.RecordReplay, opts ...Option) *EmbedderImpl {
	t.Helper()

	// Scrub auth headers for security in recordings
	rr.ScrubReq(func(req *http.Request) error {
		if auth := req.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", "Bearer test-token")
		}
		return nil
	})

	// Set test credentials for the PaLM client when replaying
	if rr.Replaying() {
		os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
		os.Setenv("GOOGLE_CLOUD_LOCATION", "test-location")
	}

	llm, err := palm.New(palm.WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	embedder, err := NewEmbedder(llm, opts...)
	require.NoError(t, err)

	return embedder
}

func TestVertexAIPaLMEmbeddings(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_CLOUD_PROJECT")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	e := newVertexEmbedder(t, rr)

	_, err := e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}

func TestVertexAIPaLMEmbeddingsWithOptions(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_CLOUD_PROJECT")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	e := newVertexEmbedder(t, rr, WithBatchSize(5), WithStripNewLines(false))

	_, err := e.EmbedQuery(ctx, "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
}
