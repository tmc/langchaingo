package cohereclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
)

// setupTestClient creates a test client with httprr recording/replay support.
func setupTestClient(t *testing.T, baseURL, model string) (*Client, *httprr.RecordReplay) {
	t.Helper()

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	apiKey := "test-api-key"
	if key := os.Getenv("COHERE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, baseURL, model, WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	return client, rr
}

func TestClient_CreateGeneration(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	client, rr := setupTestClient(t, "", "command")
	defer rr.Close()

	req := &GenerationRequest{
		Prompt: "Once upon a time in a magical forest, there lived",
	}

	resp, err := client.CreateGeneration(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_CreateGenerationWithCustomModel(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	client, rr := setupTestClient(t, "https://api.cohere.ai", "command-light")
	defer rr.Close()

	req := &GenerationRequest{
		Prompt: "What is the capital of France?",
	}

	resp, err := client.CreateGeneration(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}
