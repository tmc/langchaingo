package huggingfaceclient

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

const testURL = "https://api-inference.huggingface.co"

func TestClient_RunInference(t *testing.T) {
	t.Parallel()

	httprr.SkipIfNoCredentialsOrRecording(t, "HUGGINGFACEHUB_API_TOKEN")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	client, err := New(apiKey, "gpt2", testURL)
	require.NoError(t, err)

	req := &InferenceRequest{
		Model:       "gpt2",
		Prompt:      "Hello, my name is",
		Task:        InferenceTaskTextGeneration,
		Temperature: 0.5,
		MaxLength:   20,
	}

	resp, err := client.RunInference(t.Context(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_RunInferenceText2Text(t *testing.T) {
	t.Parallel()

	httprr.SkipIfNoCredentialsOrRecording(t, "HUGGINGFACEHUB_API_TOKEN")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	client, err := New(apiKey, "google/flan-t5-base", testURL)
	require.NoError(t, err)

	req := &InferenceRequest{
		Model:       "google/flan-t5-base",
		Prompt:      "Translate to French: Hello, how are you?",
		Task:        InferenceTaskText2TextGeneration,
		Temperature: 0.5,
		MaxLength:   50,
	}

	resp, err := client.RunInference(t.Context(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_CreateEmbedding(t *testing.T) {
	t.Parallel()

	httprr.SkipIfNoCredentialsOrRecording(t, "HUGGINGFACEHUB_API_TOKEN")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" && rr.Recording() {
		apiKey = key
	}

	// Replace http.DefaultClient with our recording client
	oldClient := http.DefaultClient
	http.DefaultClient = rr.Client()
	defer func() { http.DefaultClient = oldClient }()

	client, err := New(apiKey, "", testURL)
	require.NoError(t, err)

	req := &EmbeddingRequest{
		Inputs: []string{"Hello world", "How are you?"},
		Options: map[string]any{
			"wait_for_model": true,
		},
	}

	embeddings, err := client.CreateEmbedding(t.Context(), "sentence-transformers/all-MiniLM-L6-v2", "feature-extraction", req)
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
