package ollamaclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

// getOllamaTestURL returns the Ollama server URL to use for testing.
// It uses OLLAMA_HOST if set, otherwise defaults to localhost:11434.
func getOllamaTestURL(t *testing.T, rr *httprr.RecordReplay) string {
	t.Helper()

	// Default to localhost
	baseURL := "http://localhost:11434"

	// Use environment variable if set and we're recording
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	return baseURL
}

// checkOllamaEndpoint performs a lightweight health check on the Ollama endpoint.
// Returns true if the endpoint is available, false otherwise.
func checkOllamaEndpoint(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func TestClient_Generate(t *testing.T) {
	ctx := context.Background()

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// No auth scrubbing needed for Ollama as it doesn't use API keys

	baseURL := getOllamaTestURL(t, rr)

	// If recording and endpoint is not available, skip the test
	if rr.Recording() && !checkOllamaEndpoint(baseURL) {
		t.Skipf("Ollama endpoint not available at %s", baseURL)
	}

	// Skip if no recording exists and we're not recording
	if rr.Replaying() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := false
	req := &GenerateRequest{
		Model:  "gemma3:1b",
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

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := getOllamaTestURL(t, rr)

	// If recording and endpoint is not available, skip the test
	if rr.Recording() && !checkOllamaEndpoint(baseURL) {
		t.Skipf("Ollama endpoint not available at %s", baseURL)
	}

	// Skip if no recording exists and we're not recording
	if rr.Replaying() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := true
	req := &GenerateRequest{
		Model:  "gemma3:1b",
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

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := getOllamaTestURL(t, rr)

	// If recording and endpoint is not available, skip the test
	if rr.Recording() && !checkOllamaEndpoint(baseURL) {
		t.Skipf("Ollama endpoint not available at %s", baseURL)
	}

	// Skip if no recording exists and we're not recording
	if rr.Replaying() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "gemma3:1b",
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

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := getOllamaTestURL(t, rr)

	// If recording and endpoint is not available, skip the test
	if rr.Recording() && !checkOllamaEndpoint(baseURL) {
		t.Skipf("Ollama endpoint not available at %s", baseURL)
	}

	// Skip if no recording exists and we're not recording
	if rr.Replaying() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "gemma3:1b",
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

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := getOllamaTestURL(t, rr)

	// If recording and endpoint is not available, skip the test
	if rr.Recording() && !checkOllamaEndpoint(baseURL) {
		t.Skipf("Ollama endpoint not available at %s", baseURL)
	}

	// Skip if no recording exists and we're not recording
	if rr.Replaying() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &EmbeddingRequest{
		Model: "nomic-embed-text",
		Input: "Hello world",
		Options: Options{
			Temperature: 0.0,
		},
	}

	resp, err := client.CreateEmbedding(ctx, req)
	if err != nil && strings.Contains(err.Error(), "does not support embeddings") {
		t.Skipf("Model %s does not support embeddings", req.Model)
	}
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Embeddings)
}

func TestClient_GenerateChatWithThink(t *testing.T) {
	ctx := context.Background()

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := getOllamaTestURL(t, rr)

	// If recording and endpoint is not available, skip the test
	if rr.Recording() && !checkOllamaEndpoint(baseURL) {
		t.Skipf("Ollama endpoint not available at %s", baseURL)
	}

	// Skip if no recording exists and we're not recording
	if rr.Replaying() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "gemma3:1b",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "What is 2+2? Show your reasoning.",
			},
		},
		Stream: false,
		Think:  boolPtr(true), // Enable reasoning mode (top-level per Ollama API)
		Options: Options{
			Temperature: 0.0,
			NumPredict:  100,
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

	// The think parameter should be included in the request
	// This test verifies that the parameter is properly serialized
}

func TestChatRequestJSONMarshalWithThink(t *testing.T) {
	// The Ollama API defines "think" as a top-level field of the chat request,
	// not as part of "options". See https://github.com/tmc/langchaingo/issues/1514.
	t.Run("think true is top-level", func(t *testing.T) {
		req := ChatRequest{
			Model: "qwen3",
			Think: boolPtr(true),
			Options: Options{
				Temperature: 0.5,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal(data, &result))

		think, exists := result["think"]
		assert.True(t, exists, "think field should exist at the top level")
		assert.Equal(t, true, think, "think field should be true")

		// think must not leak into the nested options object.
		opts, ok := result["options"].(map[string]any)
		require.True(t, ok, "options object should be present")
		_, thinkInOptions := opts["think"]
		assert.False(t, thinkInOptions, "think must not be nested inside options")
	})

	t.Run("think false is preserved", func(t *testing.T) {
		// A plain bool with omitempty would drop false; the pointer keeps it so
		// callers can explicitly disable thinking on models that think by default.
		req := ChatRequest{Model: "qwen3", Think: boolPtr(false)}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal(data, &result))

		think, exists := result["think"]
		assert.True(t, exists, "explicit false think field should be preserved")
		assert.Equal(t, false, think, "think field should be false")
	})

	t.Run("think unset is omitted", func(t *testing.T) {
		req := ChatRequest{Model: "qwen3"}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal(data, &result))

		_, exists := result["think"]
		assert.False(t, exists, "unset think field should be omitted")
	})
}

func boolPtr(b bool) *bool { return &b }
