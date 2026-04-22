package anthropic

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
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

func TestHandleAIMessage_MultipleParts(t *testing.T) {
	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "I'll check the weather."},
			llms.ToolCall{
				ID:   "toolu_1",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location":"Paris"}`,
				},
			},
			llms.ToolCall{
				ID:   "toolu_2",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location":"Berlin"}`,
				},
			},
		},
	}

	got, err := handleAIMessage(msg)
	if err != nil {
		t.Fatalf("handleAIMessage() error = %v", err)
	}
	if got.Role != RoleAssistant {
		t.Errorf("role = %q, want %q", got.Role, RoleAssistant)
	}
	contents, ok := got.Content.([]anthropicclient.Content)
	if !ok {
		t.Fatalf("content type = %T, want []anthropicclient.Content", got.Content)
	}
	if len(contents) != 3 {
		t.Fatalf("content blocks = %d, want 3", len(contents))
	}

	text, ok := contents[0].(*anthropicclient.TextContent)
	if !ok {
		t.Fatalf("content[0] type = %T, want *TextContent", contents[0])
	}
	if text.Text != "I'll check the weather." {
		t.Errorf("text = %q, want %q", text.Text, "I'll check the weather.")
	}

	for i, want := range []struct{ id, loc string }{{"toolu_1", "Paris"}, {"toolu_2", "Berlin"}} {
		tu, ok := contents[i+1].(anthropicclient.ToolUseContent)
		if !ok {
			t.Fatalf("content[%d] type = %T, want ToolUseContent", i+1, contents[i+1])
		}
		if tu.ID != want.id {
			t.Errorf("tool[%d] id = %q, want %q", i, tu.ID, want.id)
		}
		if tu.Input["location"] != want.loc {
			t.Errorf("tool[%d] location = %v, want %q", i, tu.Input["location"], want.loc)
		}
	}
}

func TestHandleAIMessage_Errors(t *testing.T) {
	t.Run("empty parts", func(t *testing.T) {
		_, err := handleAIMessage(llms.MessageContent{Role: llms.ChatMessageTypeAI})
		if err == nil {
			t.Fatal("expected error for empty parts")
		}
	})

	t.Run("invalid tool call arguments", func(t *testing.T) {
		msg := llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:           "toolu_x",
					FunctionCall: &llms.FunctionCall{Name: "x", Arguments: "not-json"},
				},
			},
		}
		if _, err := handleAIMessage(msg); err == nil {
			t.Fatal("expected error for invalid tool call arguments")
		}
	})
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

func TestGenerateMessagesContent_EmptyContent(t *testing.T) {
	// This test demonstrates the need for checking len(result.Content) == 0
	// Without the fix, accessing result.Content[0] would panic when Anthropic
	// returns a response with nil or empty content (addresses issue #993)
	t.Skip("Requires mock client - would demonstrate panic without len(result.Content) == 0 check")
}
