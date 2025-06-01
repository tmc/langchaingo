package ollamaclient

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestClient_Generate(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OLLAMA_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// No auth scrubbing needed for Ollama as it doesn't use API keys

	baseURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := false
	req := &GenerateRequest{
		Model:  "llama2",
		Prompt: "Hello, how are you?",
		Stream: &stream,
		Options: Options{
			Temperature: 0.0,
			NumPredict:  100,
		},
	}

	var response *GenerateResponse
	err = client.Generate(ctx, req, func(resp GenerateResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response)
	assert.True(t, response.Done)
}

func TestClient_GenerateStream(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OLLAMA_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := true
	req := &GenerateRequest{
		Model:  "llama2",
		Prompt: "Count from 1 to 5",
		Stream: &stream,
		Options: Options{
			Temperature: 0.0,
			NumPredict:  50,
		},
	}

	var responses []GenerateResponse
	err = client.Generate(ctx, req, func(resp GenerateResponse) error {
		responses = append(responses, resp)
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responses)
	assert.True(t, responses[len(responses)-1].Done)
}

func TestClient_GenerateChat(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OLLAMA_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "llama2",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Stream: false,
		Options: Options{
			Temperature: 0.0,
			NumPredict:  50,
		},
	}

	var response *ChatResponse
	err = client.GenerateChat(ctx, req, func(resp ChatResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Message)
	assert.NotEmpty(t, response.Message.Content)
	assert.True(t, response.Done)
}

func TestClient_GenerateChatStream(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OLLAMA_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "llama2",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Count from 1 to 5",
			},
		},
		Stream: true,
		Options: Options{
			Temperature: 0.0,
			NumPredict:  50,
		},
	}

	var responses []ChatResponse
	err = client.GenerateChat(ctx, req, func(resp ChatResponse) error {
		responses = append(responses, resp)
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responses)
	assert.True(t, responses[len(responses)-1].Done)
}

func TestClient_CreateEmbedding(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OLLAMA_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &EmbeddingRequest{
		Model:  "llama2",
		Prompt: "Hello world",
		Options: Options{
			Temperature: 0.0,
		},
	}

	resp, err := client.CreateEmbedding(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Embedding)
}
