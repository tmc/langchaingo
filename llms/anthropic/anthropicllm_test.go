package anthropic

import (
	"os"
	"testing"

	"github.com/0xDezzy/langchaingo/llms"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		envToken string
		opts     []Option
		wantErr  bool
	}{
		{
			name:     "with token from env",
			envToken: "test-token",
			opts:     []Option{},
			wantErr:  false,
		},
		{
			name:     "with token option",
			envToken: "",
			opts:     []Option{WithToken("test-token")},
			wantErr:  false,
		},
		{
			name:     "missing token",
			envToken: "",
			opts:     []Option{},
			wantErr:  true,
		},
		{
			name:     "with all options",
			envToken: "test-token",
			opts: []Option{
				WithModel("claude-3-opus-20240229"),
				WithBaseURL("https://api.example.com"),
				WithAnthropicBetaHeader("max-tokens-3-5-sonnet-2024-07-15"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ANTHROPIC_API_KEY", tt.envToken)

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

func TestProcessMessages(t *testing.T) {
	tests := []struct {
		name       string
		messages   []llms.MessageContent
		wantLen    int
		wantSystem string
		wantErr    bool
	}{
		{
			name: "basic text message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
			},
			wantLen:    1,
			wantSystem: "",
			wantErr:    false,
		},
		{
			name: "system message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "You are helpful"},
					},
				},
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hi"},
					},
				},
			},
			wantLen:    1,
			wantSystem: "You are helpful",
			wantErr:    false,
		},
		{
			name: "ai and human messages",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hi there!"},
					},
				},
			},
			wantLen:    2,
			wantSystem: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, systemPrompt, err := processMessages(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != tt.wantLen {
					t.Errorf("processMessages() returned %d messages, want %d", len(result), tt.wantLen)
				}
				if systemPrompt != tt.wantSystem {
					t.Errorf("processMessages() system prompt = %q, want %q", systemPrompt, tt.wantSystem)
				}
			}
		})
	}
}

func TestToolsToTools(t *testing.T) {
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
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

	result := toolsToTools(tools)

	if len(result) != 1 {
		t.Fatalf("toolsToTools() returned %d tools, want 1", len(result))
	}
	if result[0].Name != "get_weather" {
		t.Errorf("toolsToTools() tool name = %q, want %q", result[0].Name, "get_weather")
	}
	if result[0].Description != "Get the weather for a location" {
		t.Errorf("toolsToTools() tool description = %q, want %q", result[0].Description, "Get the weather for a location")
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		opts := &options{}
		WithModel("claude-3-opus")(opts)
		if opts.model != "claude-3-opus" {
			t.Errorf("WithModel() got %s, want claude-3-opus", opts.model)
		}
	})

	t.Run("WithToken", func(t *testing.T) {
		opts := &options{}
		WithToken("test-token")(opts)
		if opts.token != "test-token" {
			t.Errorf("WithToken() got %s, want test-token", opts.token)
		}
	})

	t.Run("WithBaseURL", func(t *testing.T) {
		opts := &options{}
		WithBaseURL("https://test.com")(opts)
		if opts.baseURL != "https://test.com" {
			t.Errorf("WithBaseURL() got %s, want https://test.com", opts.baseURL)
		}
	})

	t.Run("WithAnthropicBetaHeader", func(t *testing.T) {
		opts := &options{}
		WithAnthropicBetaHeader("test-beta")(opts)
		if opts.anthropicBetaHeader != "test-beta" {
			t.Errorf("WithAnthropicBetaHeader() got %s, want test-beta", opts.anthropicBetaHeader)
		}
	})

	t.Run("WithLegacyTextCompletionsAPI", func(t *testing.T) {
		opts := &options{}
		WithLegacyTextCompletionsAPI()(opts)
		if !opts.useLegacyTextCompletionsAPI {
			t.Error("WithLegacyTextCompletionsAPI() did not set flag")
		}
	})
}

func TestCall(t *testing.T) {
	// Test that Call delegates to GenerateContent
	t.Skip("Call() requires integration testing with mock client")
}
