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

	think := true
	req := &ChatRequest{
		Model: "gemma3:1b",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "What is 2+2? Show your reasoning.",
			},
		},
		Stream: false,
		Think:  &think, // Enable reasoning mode (top-level per Ollama API)
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

	// The think parameter should be included at the top level of the request
	// This test verifies that the parameter is properly serialized
}

func TestChatRequestJSONMarshalWithThink(t *testing.T) {
	// Test that the think parameter is properly marshaled as a top-level field
	// (not nested inside "options") per the Ollama API spec.

	t.Run("think=true", func(t *testing.T) {
		think := true
		req := ChatRequest{
			Model:  "test-model",
			Think:  &think,
			Stream: false,
			Options: Options{
				Temperature: 0.5,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		// think should be a top-level field
		thinkVal, exists := result["think"]
		assert.True(t, exists, "think should be a top-level field in JSON")
		assert.Equal(t, true, thinkVal, "think should be true")

		// think should NOT be inside options
		options, hasOptions := result["options"].(map[string]interface{})
		assert.True(t, hasOptions, "options field should exist")
		_, thinkInOptions := options["think"]
		assert.False(t, thinkInOptions, "think should NOT be nested inside options")
	})

	t.Run("think=false", func(t *testing.T) {
		think := false
		req := ChatRequest{
			Model:  "test-model",
			Think:  &think,
			Stream: false,
			Options: Options{
				Temperature: 0.5,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		// think=false should still be present (pointer to false, not zero-value bool)
		thinkVal, exists := result["think"]
		assert.True(t, exists, "think=false should be present in JSON (pointer distinguishes unset from false)")
		assert.Equal(t, false, thinkVal, "think should be false")
	})

	t.Run("think=nil (unset)", func(t *testing.T) {
		req := ChatRequest{
			Model:  "test-model",
			Think:  nil,
			Stream: false,
			Options: Options{
				Temperature: 0.5,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		// think should be omitted entirely when nil
		_, exists := result["think"]
		assert.False(t, exists, "think should be omitted when nil (unset)")
	})
}

func TestGenerateRequestJSONMarshalWithThink(t *testing.T) {
	// Test that the think parameter is properly marshaled as a top-level field
	// in GenerateRequest as well.

	t.Run("think=true", func(t *testing.T) {
		think := true
		stream := false
		req := GenerateRequest{
			Model:  "test-model",
			Prompt: "hello",
			Think:  &think,
			Stream: &stream,
			Options: Options{
				Temperature: 0.5,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		thinkVal, exists := result["think"]
		assert.True(t, exists, "think should be a top-level field")
		assert.Equal(t, true, thinkVal)

		// Should NOT be in options
		options := result["options"].(map[string]interface{})
		_, inOpts := options["think"]
		assert.False(t, inOpts, "think should NOT be inside options")
	})

	t.Run("think=nil (unset)", func(t *testing.T) {
		stream := false
		req := GenerateRequest{
			Model:  "test-model",
			Prompt: "hello",
			Think:  nil,
			Stream: &stream,
			Options: Options{
				Temperature: 0.5,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		_, exists := result["think"]
		assert.False(t, exists, "think should be omitted when nil")
	})
}
