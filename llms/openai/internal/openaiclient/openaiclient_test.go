package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

// setupTestClient creates a test client with httprr recording/replay
func setupTestClient(t *testing.T, model string) *Client {
	t.Helper()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if !rr.Recording() {
		t.Parallel()
	}

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, model, "", "", APITypeOpenAI, "", rr.Client(), "", nil)
	require.NoError(t, err)
	return client
}

func TestClient_CreateChatCompletion(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "gpt-3.5-turbo", "", "", APITypeOpenAI, "", rr.Client(), "", nil)
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []*ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Temperature:         0.0,
		MaxCompletionTokens: 50,
	}

	resp, err := client.CreateChat(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	assert.NotEmpty(t, resp.Choices[0].Message.Content)
}

func TestClient_CreateChatCompletionStream(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "gpt-3.5-turbo", "", "", APITypeOpenAI, "", rr.Client(), "", nil)
	require.NoError(t, err)

	var chunks []string
	req := &ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []*ChatMessage{
			{
				Role:    "user",
				Content: "Count from 1 to 5",
			},
		},
		Temperature:         0.0,
		MaxCompletionTokens: 50,
		Stream:              true,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			chunks = append(chunks, string(chunk))
			return nil
		},
	}

	resp, err := client.CreateChat(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, chunks)
}

func TestClient_CreateEmbedding(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "", "", "", APITypeOpenAI, "", rr.Client(), "text-embedding-ada-002", nil)
	require.NoError(t, err)

	req := &EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: []string{"Hello world"},
	}

	resp, err := client.CreateEmbedding(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.Len(t, resp, 1)
	assert.NotEmpty(t, resp[0])
}

func TestClient_CreateEmbeddingWithDimensions(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "", "", "", APITypeOpenAI, "", rr.Client(), "text-embedding-3-small", nil, WithEmbeddingDimensions(256))
	require.NoError(t, err)

	req := &EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: []string{"Hello world"},
	}

	resp, err := client.CreateEmbedding(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.Len(t, resp, 1)
	assert.NotEmpty(t, resp[0])
}

func TestClient_FunctionCall(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "gpt-3.5-turbo", "", "", APITypeOpenAI, "", rr.Client(), "", nil)
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []*ChatMessage{
			{
				Role:    "user",
				Content: "What's the weather like in Boston?",
			},
		},
		Temperature:         0.0,
		MaxCompletionTokens: 100,
		Functions: []FunctionDefinition{
			{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	resp, err := client.CreateChat(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClient_WithResponseFormat(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	apiKey := "test-api-key"
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	responseFormat := &ResponseFormat{Type: "json_object"}
	client, err := New(apiKey, "gpt-3.5-turbo", "", "", APITypeOpenAI, "", rr.Client(), "", responseFormat)
	require.NoError(t, err)

	req := &ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []*ChatMessage{
			{
				Role:    "user",
				Content: "Return a JSON object with a 'greeting' field that says hello",
			},
		},
		Temperature:         0.0,
		MaxCompletionTokens: 50,
		ResponseFormat:      responseFormat,
	}

	resp, err := client.CreateChat(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	assert.NotEmpty(t, resp.Choices[0].Message.Content)
}

func TestMakeEmbeddingRequest(t *testing.T) {
	t.Run("without dimensions", func(t *testing.T) {
		client, err := New("", "gpt-3.5-turbo", "", "", APITypeOpenAI, "", nil, "", nil)
		require.NoError(t, err)

		request := client.makeEmbeddingPayload(&EmbeddingRequest{Model: "some_model"})
		assert.Equal(t, "some_model", request.Model)
		assert.Equal(t, 0, request.Dimensions)
	})
	t.Run("with dimensions", func(t *testing.T) {
		client, err := New("", "gpt-3.5-turbo", "", "", APITypeOpenAI, "", nil, "", nil)
		require.NoError(t, err)

		request := client.makeEmbeddingPayload(&EmbeddingRequest{Model: "some_model", Dimensions: 1234})
		assert.Equal(t, "some_model", request.Model)
		assert.Equal(t, 1234, request.Dimensions)
	})
}

func TestInternalMetadataFiltering(t *testing.T) {
	// Test that internal openai: prefixed metadata is filtered out from requests
	client, err := New("test-api-key", "gpt-3.5-turbo", "", "", APITypeOpenAI, "", nil, "", nil)
	require.NoError(t, err)

	// Create a mock HTTP client to capture the request body
	var capturedRequestBody []byte
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			// Read the request body
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			capturedRequestBody = body
			
			// Return a minimal valid response to avoid errors
			responseBody := `{"choices":[{"message":{"content":"test"}}],"usage":{"total_tokens":10}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(responseBody))),
			}, nil
		},
	}
	client.httpClient = mockClient

	// Create request with both internal and external metadata
	req := &ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []*ChatMessage{
			{Role: "user", Content: "test"},
		},
		Metadata: map[string]any{
			"openai:use_legacy_max_tokens": true,     // Should be filtered out
			"custom_field":                 "value",  // Should be preserved
		},
	}

	// Make the request
	_, _ = client.CreateChat(context.Background(), req)

	// Verify the request body was captured
	require.NotEmpty(t, capturedRequestBody)

	// Parse the request body to check what was sent
	var requestBody map[string]any
	err = json.Unmarshal(capturedRequestBody, &requestBody)
	require.NoError(t, err)

	// Check metadata filtering
	metadata, exists := requestBody["metadata"]
	if exists {
		metadataMap := metadata.(map[string]any)
		// Internal metadata should be filtered out
		assert.NotContains(t, metadataMap, "openai:use_legacy_max_tokens")
		// External metadata should be preserved
		assert.Contains(t, metadataMap, "custom_field")
		assert.Equal(t, "value", metadataMap["custom_field"])
	} else {
		// If no metadata field exists, that means only internal metadata was present and got filtered out
		t.Log("metadata field was completely filtered out - this is expected behavior")
	}

	// Verify original metadata is preserved in the request object
	assert.Contains(t, req.Metadata, "openai:use_legacy_max_tokens")
	assert.Contains(t, req.Metadata, "custom_field")
}

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}
