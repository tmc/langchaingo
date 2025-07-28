package openaiclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/internal/httprr"
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
