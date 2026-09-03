package anthropic

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestHandleAIMessagePreservesAllParts(t *testing.T) {
	t.Parallel()

	message, err := handleAIMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "I'll check both locations."},
			llms.ToolCall{
				ID: "tool-1",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location":"Paris"}`,
				},
			},
			llms.ToolCall{
				ID: "tool-2",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location":"London"}`,
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, RoleAssistant, message.Role)

	contents, ok := message.Content.([]anthropicclient.Content)
	require.True(t, ok)
	require.Len(t, contents, 3)

	text, ok := contents[0].(*anthropicclient.TextContent)
	require.True(t, ok)
	require.Equal(t, "I'll check both locations.", text.Text)

	firstTool, ok := contents[1].(anthropicclient.ToolUseContent)
	require.True(t, ok)
	require.Equal(t, "tool-1", firstTool.ID)
	require.Equal(t, "get_weather", firstTool.Name)
	require.Equal(t, map[string]interface{}{"location": "Paris"}, firstTool.Input)

	secondTool, ok := contents[2].(anthropicclient.ToolUseContent)
	require.True(t, ok)
	require.Equal(t, "tool-2", secondTool.ID)
	require.Equal(t, map[string]interface{}{"location": "London"}, secondTool.Input)
}

func TestHandleToolMessagePreservesAllResults(t *testing.T) {
	t.Parallel()

	message, err := handleToolMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{ToolCallID: "tool-1", Content: "sunny"},
			llms.ToolCallResponse{ToolCallID: "tool-2", Content: "rainy"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, RoleUser, message.Role)

	contents, ok := message.Content.([]anthropicclient.Content)
	require.True(t, ok)
	require.Len(t, contents, 2)

	firstResult, ok := contents[0].(anthropicclient.ToolResultContent)
	require.True(t, ok)
	require.Equal(t, "tool-1", firstResult.ToolUseID)
	require.Equal(t, "sunny", firstResult.Content)

	secondResult, ok := contents[1].(anthropicclient.ToolResultContent)
	require.True(t, ok)
	require.Equal(t, "tool-2", secondResult.ToolUseID)
	require.Equal(t, "rainy", secondResult.Content)
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
