package cloudflare

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/cloudflare/internal/cloudflareclient"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "basic initialization",
			opts: []Option{
				WithToken("test-token"),
				WithAccountID("test-account-id"),
				WithModel("test-model"),
			},
			wantErr: false,
		},
		{
			name: "with custom http client",
			opts: []Option{
				WithToken("test-token"),
				WithAccountID("test-account-id"),
				WithModel("test-model"),
				WithHTTPClient(&http.Client{}),
			},
			wantErr: false,
		},
		{
			name: "with system message",
			opts: []Option{
				WithToken("test-token"),
				WithAccountID("test-account-id"),
				WithModel("test-model"),
				WithSystemPrompt("You are a helpful assistant"),
			},
			wantErr: false,
		},
		{
			name: "with embedding model",
			opts: []Option{
				WithToken("test-token"),
				WithAccountID("test-account-id"),
				WithModel("test-model"),
				WithEmbeddingModel("test-embedding-model"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestNewWithServerURL(t *testing.T) {
	// Test with valid URL
	validURL, _ := url.Parse("https://custom.cloudflare.com")
	llm, err := New(
		WithToken("test-token"),
		WithAccountID("test-account-id"),
		WithModel("test-model"),
		WithCloudflareServerURL(validURL),
	)
	if err != nil {
		t.Errorf("New() with valid URL error = %v", err)
	}
	if llm == nil {
		t.Error("New() returned nil LLM")
	}
}

func TestTypeToRole(t *testing.T) {
	tests := []struct {
		name     string
		typ      llms.ChatMessageType
		expected cloudflareclient.Role
	}{
		{
			name:     "system message",
			typ:      llms.ChatMessageTypeSystem,
			expected: cloudflareclient.RoleSystem,
		},
		{
			name:     "AI message",
			typ:      llms.ChatMessageTypeAI,
			expected: cloudflareclient.RoleAssistant,
		},
		{
			name:     "human message",
			typ:      llms.ChatMessageTypeHuman,
			expected: cloudflareclient.RoleTypeUser,
		},
		{
			name:     "generic message",
			typ:      llms.ChatMessageTypeGeneric,
			expected: cloudflareclient.RoleTypeUser,
		},
		{
			name:     "function message",
			typ:      llms.ChatMessageTypeFunction,
			expected: "function",
		},
		{
			name:     "tool message",
			typ:      llms.ChatMessageTypeTool,
			expected: "tool",
		},
		{
			name:     "unknown type",
			typ:      llms.ChatMessageType("unknown"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typeToRole(tt.typ)
			if result != tt.expected {
				t.Errorf("typeToRole(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

func TestGenerateContentErrors(t *testing.T) {
	llm, _ := New(
		WithToken("test-token"),
		WithAccountID("test-account-id"),
		WithModel("test-model"),
	)

	ctx := context.Background()

	tests := []struct {
		name     string
		messages []llms.MessageContent
		wantErr  string
	}{
		{
			name: "multiple text parts",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "First text"},
						llms.TextContent{Text: "Second text"},
					},
				},
			},
			wantErr: "expecting a single Text content",
		},
		{
			name: "binary content not supported",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.BinaryContent{Data: []byte("data"), MIMEType: "image/png"},
					},
				},
			},
			wantErr: "only supports Text right now",
		},
		{
			name: "unsupported content type",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.ToolCall{ID: "test", Type: "function"},
					},
				},
			},
			wantErr: "only supports Text right now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := llm.GenerateContent(ctx, tt.messages)
			if err == nil {
				t.Error("GenerateContent() expected error but got nil")
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("GenerateContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCall(t *testing.T) {
	// Create a test server that returns a successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"result": {
				"response": "Hello! How can I help you today?"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	llm, _ := New(
		WithToken("test-token"),
		WithAccountID("test-account-id"),
		WithModel("test-model"),
		WithCloudflareServerURL(serverURL),
	)

	result, err := llm.Call(context.Background(), "Hello")
	if err != nil {
		t.Errorf("Call() error = %v", err)
	}
	if result != "Hello! How can I help you today?" {
		t.Errorf("Call() result = %v, want %v", result, "Hello! How can I help you today?")
	}
}

func TestCreateEmbedding(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		inputTexts     []string
		wantErr        error
		checkEmbedding bool
	}{
		{
			name: "successful embedding",
			serverResponse: `{
				"result": {
					"data": [[0.1, 0.2, 0.3], [0.4, 0.5, 0.6]]
				},
				"success": true
			}`,
			inputTexts:     []string{"text1", "text2"},
			wantErr:        nil,
			checkEmbedding: true,
		},
		{
			name: "empty response",
			serverResponse: `{
				"result": {
					"data": []
				},
				"success": true
			}`,
			inputTexts: []string{"text1"},
			wantErr:    ErrEmptyResponse,
		},
		{
			name: "incomplete embedding",
			serverResponse: `{
				"result": {
					"data": [[0.1, 0.2, 0.3]]
				},
				"success": true
			}`,
			inputTexts: []string{"text1", "text2"},
			wantErr:    ErrIncompleteEmbedding,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(tt.serverResponse))
				assert.NoError(t, err)
			}))
			defer server.Close()

			serverURL, _ := url.Parse(server.URL)
			llm, _ := New(
				WithToken("test-token"),
				WithAccountID("test-account-id"),
				WithModel("test-model"),
				WithEmbeddingModel("test-embedding-model"),
				WithCloudflareServerURL(serverURL),
			)

			embeddings, err := llm.CreateEmbedding(context.Background(), tt.inputTexts)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateEmbedding() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkEmbedding && len(embeddings) != len(tt.inputTexts) {
				t.Errorf("CreateEmbedding() returned %d embeddings, want %d", len(embeddings), len(tt.inputTexts))
			}
		})
	}
}

func TestGenerateContentWithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"result": {
				"response": "I am a helpful assistant. How can I help you?"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	llm, _ := New(
		WithToken("test-token"),
		WithAccountID("test-account-id"),
		WithModel("test-model"),
		WithCloudflareServerURL(serverURL),
		WithSystemPrompt("You are a helpful assistant"),
	)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello"},
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), messages)
	if err != nil {
		t.Errorf("GenerateContent() error = %v", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		t.Error("GenerateContent() returned empty response")
	}
}

func TestGenerateContentWithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"result": {
				"response": ""
			},
			"success": false,
			"errors": [
				{"message": "Invalid API key"},
				{"message": "Rate limit exceeded"}
			],
			"messages": []
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	llm, _ := New(
		WithToken("test-token"),
		WithAccountID("test-account-id"),
		WithModel("test-model"),
		WithCloudflareServerURL(serverURL),
	)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello"},
			},
		},
	}

	_, err := llm.GenerateContent(context.Background(), messages)
	if err == nil {
		t.Error("GenerateContent() expected error but got nil")
	}
	// The error should contain the first error message
	if err.Error() != "Invalid API key" {
		t.Errorf("GenerateContent() error = %v, want error containing 'Invalid API key'", err)
	}
}
