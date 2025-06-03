package huggingfaceclient

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
)

const testURL = "https://api-inference.huggingface.co"

func TestClient_RunInference(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Check both HF_TOKEN and HUGGINGFACEHUB_API_TOKEN
	if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
	}

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if rr.Recording() {
		// Try HF_TOKEN first, then fall back to HUGGINGFACEHUB_API_TOKEN
		if key := os.Getenv("HF_TOKEN"); key != "" {
			apiKey = key
		} else if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" {
			apiKey = key
		}
	}

	// Create client with recording HTTP client
	client, err := New(apiKey, "gpt2", testURL)
	require.NoError(t, err)
	// Note: The client already uses httputil.DefaultClient internally,
	// which is wrapped by httprr through httputil.DefaultTransport

	req := &InferenceRequest{
		Model:       "gpt2",
		Prompt:      "Hello, my name is",
		Task:        InferenceTaskTextGeneration,
		Temperature: 0.5,
		MaxLength:   20,
	}

	resp, err := client.RunInference(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_RunInferenceText2Text(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Check both HF_TOKEN and HUGGINGFACEHUB_API_TOKEN
	if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
	}

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if rr.Recording() {
		// Try HF_TOKEN first, then fall back to HUGGINGFACEHUB_API_TOKEN
		if key := os.Getenv("HF_TOKEN"); key != "" {
			apiKey = key
		} else if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" {
			apiKey = key
		}
	}

	// Create client with recording HTTP client
	client, err := New(apiKey, "google/flan-t5-base", testURL)
	require.NoError(t, err)
	// Note: The client already uses httputil.DefaultClient internally,
	// which is wrapped by httprr through httputil.DefaultTransport

	req := &InferenceRequest{
		Model:       "google/flan-t5-base",
		Prompt:      "Translate to French: Hello, how are you?",
		Task:        InferenceTaskText2TextGeneration,
		Temperature: 0.5,
		MaxLength:   50,
	}

	resp, err := client.RunInference(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_CreateEmbedding(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Check both HF_TOKEN and HUGGINGFACEHUB_API_TOKEN
	if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
	}

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if rr.Recording() {
		// Try HF_TOKEN first, then fall back to HUGGINGFACEHUB_API_TOKEN
		if key := os.Getenv("HF_TOKEN"); key != "" {
			apiKey = key
		} else if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" {
			apiKey = key
		}
	}

	// Create client with recording HTTP client
	client, err := New(apiKey, "", testURL)
	require.NoError(t, err)
	// Note: The client already uses httputil.DefaultClient internally,
	// which is wrapped by httprr through httputil.DefaultTransport

	req := &EmbeddingRequest{
		Inputs: []string{"Hello world", "How are you?"},
		Options: map[string]any{
			"wait_for_model": true,
		},
	}

	embeddings, err := client.CreateEmbedding(ctx, "BAAI/bge-small-en-v1.5", "feature-extraction", req)
	require.NoError(t, err)
	assert.NotNil(t, embeddings)
	assert.Len(t, embeddings, 2)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
}

func TestClient_InvalidToken(t *testing.T) {
	t.Parallel()

	_, err := New("", "model", testURL)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
