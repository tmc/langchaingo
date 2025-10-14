package openai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/llms"
)

// TestOpenRouterStreamingPrefix tests that the client correctly handles OpenRouter's ":openrouter" prefix
func TestOpenRouterStreamingPrefix(t *testing.T) {
	t.Parallel()

	// Create a mock server that simulates OpenRouter's response with prefix
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Simulate OpenRouter's response with the problematic prefix
		responses := []string{
			":openrouter", // This is the prefix that used to cause errors
			"",
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"test\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\"},\"finish_reason\":null}]}",
			"",
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"test\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Test\"},\"finish_reason\":null}]}",
			"",
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"test\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\" response\"},\"finish_reason\":null}]}",
			"",
			"data: {\"id\":\"chatcmpl-123\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"test\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}",
			"",
			"data: [DONE]",
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for _, response := range responses {
			_, _ = w.Write([]byte(response + "\n"))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Create client pointing to mock server
	llm, err := New(
		WithToken("test-key"),
		WithBaseURL(server.URL),
		WithModel("test-model"),
	)
	require.NoError(t, err)

	ctx := context.Background()
	var chunks []string

	_, err = llm.Call(ctx, "Test message",
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			chunks = append(chunks, string(chunk))
			return nil
		}),
	)

	// The fix ensures this doesn't fail despite the ":openrouter" prefix
	require.NoError(t, err, "should handle OpenRouter prefix without error")
	assert.NotEmpty(t, chunks, "should receive streamed chunks")

	fullResponse := strings.Join(chunks, "")
	assert.Equal(t, "Test response", fullResponse, "should receive complete response")
}

// TestOpenRouterRateLimitHandling tests graceful handling of 429 rate limit errors
func TestOpenRouterRateLimitHandling(t *testing.T) {
	t.Parallel()

	requestCount := 0

	// Create a mock server that returns 429 on first request, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount == 1 {
			// First request returns rate limit error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{
				"error": {
					"message": "Rate limit exceeded: limit_rpm/meta-llama/llama-3.2-3b-instruct. Limited to 1 request per minute.",
					"type": "rate_limit_error",
					"code": "rate_limit_exceeded"
				}
			}`))
			return
		}

		// Second request succeeds
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "test",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Success after retry"
				},
				"finish_reason": "stop"
			}]
		}`))
	}))
	defer server.Close()

	// Create client pointing to mock server
	llm, err := New(
		WithToken("test-key"),
		WithBaseURL(server.URL),
		WithModel("test-model"),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// First call should fail with rate limit
	_, err = llm.Call(ctx, "Test message")
	assert.Error(t, err, "should return rate limit error")
	assert.Contains(t, err.Error(), "429", "error should indicate rate limit")

	// Second call should succeed (simulating retry after rate limit)
	response, err := llm.Call(ctx, "Test message")
	require.NoError(t, err, "should succeed after rate limit clears")
	assert.Equal(t, "Success after retry", response)
}
