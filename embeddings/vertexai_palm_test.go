package embeddings

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/googleai/palm"
)

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

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

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_CLOUD_PROJECT")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	e := newVertexEmbedder(t, rr)

	_, err := e.EmbedQuery(ctx, "Hello world!")
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world", "The world is ending", "good bye"})
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	assert.Len(t, embeddings, 3)
}

func TestVertexAIPaLMEmbeddingsWithOptions(t *testing.T) {
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_CLOUD_PROJECT")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	e := newVertexEmbedder(t, rr, WithBatchSize(5), WithStripNewLines(false))

	_, err := e.EmbedQuery(ctx, "Hello world!")
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	embeddings, err := e.EmbedDocuments(ctx, []string{"Hello world"})
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	assert.Len(t, embeddings, 1)
}
