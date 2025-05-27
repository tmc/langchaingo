package cohereclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestClient_CreateGeneration(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer test-api-key")
		return nil
	})

	apiKey := "test-api-key"
	if key := os.Getenv("COHERE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "", "command", WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	req := &GenerationRequest{
		Prompt: "Once upon a time in a magical forest, there lived",
	}

	resp, err := client.CreateGeneration(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_CreateGenerationWithCustomModel(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer test-api-key")
		return nil
	})

	apiKey := "test-api-key"
	if key := os.Getenv("COHERE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "https://api.cohere.ai", "command-light", WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	req := &GenerationRequest{
		Prompt: "What is the capital of France?",
	}

	resp, err := client.CreateGeneration(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}