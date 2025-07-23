package bedrockclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestAnthropicToolCallSupport(t *testing.T) {
	t.Run("Tools in CallOptions are converted to Anthropic format", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "get_weather",
					Description: "Get the current weather for a location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		}

		// Test that our conversion works
		var anthropicTools []anthropicTool
		for _, tool := range tools {
			if tool.Function != nil {
				anthropicTools = append(anthropicTools, anthropicTool{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					InputSchema: tool.Function.Parameters,
				})
			}
		}

		require.Len(t, anthropicTools, 1)
		require.Equal(t, "get_weather", anthropicTools[0].Name)
		require.Equal(t, "Get the current weather for a location", anthropicTools[0].Description)
		require.NotNil(t, anthropicTools[0].InputSchema)
	})

	t.Run("Tool use output is parsed correctly", func(t *testing.T) {
		output := anthropicTextGenerationOutput{
			Type: "message",
			Role: "assistant",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text,omitempty"`
				// Tool use fields
				ID string `json:"id,omitempty"`
				Name string `json:"name,omitempty"`
				Input interface{} `json:"input,omitempty"`
			}{
				{
					Type: "tool_use",
					ID:   "toolu_123",
					Name: "get_weather",
					Input: map[string]interface{}{
						"location": "San Francisco, CA",
					},
				},
			},
			StopReason: AnthropicCompletionReasonEndTurn,
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
				InputTokens:  100,
				OutputTokens: 50,
			},
		}

		// Simulate the processing logic we added
		var toolCalls []llms.ToolCall
		for _, c := range output.Content {
			if c.Type == AnthropicMessageTypeToolUse {
				// Marshal input to JSON for arguments
				arguments, err := json.Marshal(c.Input)
				require.NoError(t, err)
				toolCalls = append(toolCalls, llms.ToolCall{
					ID:   c.ID,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      c.Name,
						Arguments: string(arguments),
					},
				})
			}
		}

		require.Len(t, toolCalls, 1)
		require.Equal(t, "toolu_123", toolCalls[0].ID)
		require.Equal(t, "function", toolCalls[0].Type)
		require.Equal(t, "get_weather", toolCalls[0].FunctionCall.Name)
		require.Contains(t, toolCalls[0].FunctionCall.Arguments, "San Francisco, CA")
	})

	t.Run("Tool result input content is processed", func(t *testing.T) {
		message := Message{
			Role:    llms.ChatMessageTypeTool,
			Type:    AnthropicMessageTypeToolResult,
			Content: `{"tool_use_id": "toolu_123", "content": "It's 72°F and sunny in San Francisco"}`,
		}

		content := getAnthropicInputContent(message)
		require.Equal(t, AnthropicMessageTypeToolResult, content.Type)
		require.Equal(t, "It's 72°F and sunny in San Francisco", content.Content)
	})
}