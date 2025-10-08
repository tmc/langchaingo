package bedrock

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/llms"
)

func TestToolCallProcessing(t *testing.T) {
	// Test that tool calls are properly processed into bedrock messages
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What's the weather like?"},
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "call_123",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York"}`,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "call_123",
					Name:       "get_weather",
					Content:    "It's sunny and 72°F in New York",
				},
			},
		},
	}

	bedrockMsgs, err := processMessages(messages)
	require.NoError(t, err)
	require.Len(t, bedrockMsgs, 3)

	// First message should be text
	require.Equal(t, llms.ChatMessageTypeHuman, bedrockMsgs[0].Role)
	require.Equal(t, "text", bedrockMsgs[0].Type)
	require.Equal(t, "What's the weather like?", bedrockMsgs[0].Content)

	// Second message should be tool_call
	require.Equal(t, llms.ChatMessageTypeAI, bedrockMsgs[1].Role)
	require.Equal(t, "tool_call", bedrockMsgs[1].Type)
	require.Equal(t, "call_123", bedrockMsgs[1].ToolCallID)
	require.Equal(t, "get_weather", bedrockMsgs[1].ToolName)

	// Third message should be tool_result
	require.Equal(t, llms.ChatMessageTypeTool, bedrockMsgs[2].Role)
	require.Equal(t, "tool_result", bedrockMsgs[2].Type)
	require.Equal(t, "call_123", bedrockMsgs[2].ToolUseID)
	require.Equal(t, "It's sunny and 72°F in New York", bedrockMsgs[2].Content)
}
