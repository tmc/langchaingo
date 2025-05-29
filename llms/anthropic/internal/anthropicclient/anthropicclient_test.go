package anthropicclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestClient_CreateCompletion(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "ANTHROPIC_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	client, err := New(apiKey, "claude-2.1", DefaultBaseURL, WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	client.UseLegacyTextCompletionsAPI = true

	req := &CompletionRequest{
		Model:       "claude-2.1",
		Prompt:      "\n\nHuman: Hello, how are you?\n\nAssistant:",
		Temperature: 0.0,
		MaxTokens:   100,
	}

	resp, err := client.CreateCompletion(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Text)
}

func TestClient_CreateMessage(t *testing.T) {
	ctx := context.Background()

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "claude-3-opus-20240229", DefaultBaseURL, WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	req := &MessageRequest{
		Model: "claude-3-opus-20240229",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: []Content{TextContent{Text: "Hello, how are you?"}},
			},
		},
		MaxTokens: 100,
	}

	resp, err := client.CreateMessage(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
}

func TestClient_CreateMessageStream(t *testing.T) {
	ctx := context.Background()

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "claude-3-opus-20240229", DefaultBaseURL, WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	var chunks []string
	req := &MessageRequest{
		Model: "claude-3-opus-20240229",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: []Content{TextContent{Text: "Count from 1 to 5"}},
			},
		},
		MaxTokens: 100,
		Stream:    true,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			chunks = append(chunks, string(chunk))
			return nil
		},
	}

	resp, err := client.CreateMessage(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, chunks)
}

func TestClient_WithAnthropicBetaHeader(t *testing.T) {
	ctx := context.Background()

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	apiKey := "test-api-key"
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client, err := New(apiKey, "claude-3-opus-20240229", DefaultBaseURL,
		WithHTTPClient(rr.Client()),
		WithAnthropicBetaHeader("tools-2024-05-16"),
	)
	require.NoError(t, err)

	req := &MessageRequest{
		Model: "claude-3-opus-20240229",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: []Content{TextContent{Text: "What's the weather like?"}},
			},
		},
		MaxTokens: 100,
		Tools: []Tool{
			{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	resp, err := client.CreateMessage(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

type specialContent struct{}

func (s specialContent) GetType() string {
	return "its-so-special"
}

func TestClient_applyCacheSettings(t *testing.T) {
	basicMessage := MessageRequest{
		System: []SystemTextMessage{{Type: "text", Text: "System message"}},
		Messages: []ChatMessage{{
			Role:    "user",
			Content: []Content{TextContent{Type: "text", Text: "User message"}},
		}},
		Tools: []Tool{{Name: "callMe"}},
	}
	modifiedBasicMessage := func(mod func(m *MessageRequest)) MessageRequest {
		message := basicMessage
		mod(&message)
		return message
	}

	tests := []struct {
		name     string
		client   Client
		input    MessageRequest
		expected MessageRequest
	}{
		{
			name:     "no cache settings - should not modify input",
			client:   Client{},
			input:    basicMessage,
			expected: basicMessage,
		},
		{
			name:     "cache for tools - but no tools",
			client:   Client{cacheTools: true},
			input:    modifiedBasicMessage(func(m *MessageRequest) { m.Tools = nil }),
			expected: modifiedBasicMessage(func(m *MessageRequest) { m.Tools = nil }),
		},
		{
			name:   "cache for tools - only last tool contains cache control",
			client: Client{cacheTools: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Tools = append(m.Tools, Tool{Name: "callMeMore"})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Tools = append(m.Tools, Tool{Name: "callMeMore", CacheControl: &ephemeralCache})
			}),
		},
		{
			name:     "cache for system message - but no system message",
			client:   Client{cacheSystemMessage: true},
			input:    modifiedBasicMessage(func(m *MessageRequest) { m.System = nil }),
			expected: modifiedBasicMessage(func(m *MessageRequest) { m.System = nil }),
		},
		{
			name:   "cache for system message - only last system message contains cache control",
			client: Client{cacheSystemMessage: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.System = append(m.System, SystemTextMessage{Text: "System message #2"})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.System = append(m.System, SystemTextMessage{Text: "System message #2", CacheControl: &ephemeralCache})
			}),
		},
		{
			name:     "cache for chat - but no message",
			client:   Client{cacheChat: true},
			input:    modifiedBasicMessage(func(m *MessageRequest) { m.Messages = nil }),
			expected: modifiedBasicMessage(func(m *MessageRequest) { m.Messages = nil }),
		},
		{
			name:   "cache for chat - only last text message contains cache control",
			client: Client{cacheChat: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{TextContent{Text: "User message #2"}}})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{TextContent{Text: "User message #2", CacheControl: &ephemeralCache}}})
			}),
		},
		{
			name:   "cache for chat - only last non-empty text message contains cache control",
			client: Client{cacheChat: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{TextContent{Text: "User message #2"}}})
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{TextContent{}}})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{TextContent{Text: "User message #2", CacheControl: &ephemeralCache}}})
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{TextContent{}}})
			}),
		},
		{
			name:   "cache for chat - only last image message contains cache control",
			client: Client{cacheChat: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{ImageContent{Source: ImageSource{Data: "beautiful_image.jpg"}}}})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{ImageContent{Source: ImageSource{Data: "beautiful_image.jpg"}, CacheControl: &ephemeralCache}}})
			}),
		},
		{
			name:   "cache for chat - only last tool use contains cache control",
			client: Client{cacheChat: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{ToolUseContent{Name: "callMe"}}})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{ToolUseContent{Name: "callMe", CacheControl: &ephemeralCache}}})
			}),
		},
		{
			name:   "cache for chat - only last tool result contains cache control",
			client: Client{cacheChat: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{ToolResultContent{Content: "is anybody out there?"}}})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{ToolResultContent{Content: "is anybody out there?", CacheControl: &ephemeralCache}}})
			}),
		},
		{
			name:   "cache for chat - no cacheable message",
			client: Client{cacheChat: true},
			input: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{&specialContent{}}})
			}),
			expected: modifiedBasicMessage(func(m *MessageRequest) {
				m.Messages = append(m.Messages, ChatMessage{Content: []Content{&specialContent{}}})
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.client.applyCacheSettings(&tt.input)

			assert.Equal(t, tt.expected, tt.input)
		})
	}

}
