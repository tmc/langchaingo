package memory_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestConversationBufferRetainsToolCallContext(t *testing.T) {
	t.Parallel()
	call := llms.AIChatMessage{ToolCalls: []llms.ToolCall{
		{ID: "call-1", Type: "function", FunctionCall: &llms.FunctionCall{
			Name: "lookup", Arguments: `{"city":"Paris"}`,
		}},
	}}
	history := memory.NewChatMessageHistory(memory.WithPreviousMessages([]llms.ChatMessage{
		llms.HumanChatMessage{Content: "Find the weather"},
		call,
		llms.ToolChatMessage{ID: "call-1", Content: "20 degrees"},
	}))
	buffer := memory.NewConversationBuffer(memory.WithChatHistory(history))
	result, err := buffer.LoadMemoryVariables(context.Background(), nil)
	require.NoError(t, err)
	text, ok := result["history"].(string)
	require.True(t, ok)
	require.Contains(t, text, `"id":"call-1"`)
	require.Contains(t, text, `"name":"lookup"`)
	require.Contains(t, text, `\"city\":\"Paris\"`)
	require.Contains(t, text, "tool: 20 degrees")
}

func TestTokenBufferCountsToolArguments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	call := llms.AIChatMessage{ToolCalls: []llms.ToolCall{
		{ID: "call-1", Type: "function", FunctionCall: &llms.FunctionCall{
			Name: "lookup", Arguments: `{"query":"` + strings.Repeat("weather ", 2000) + `"}`,
		}},
	}}
	history := memory.NewChatMessageHistory(memory.WithPreviousMessages([]llms.ChatMessage{
		llms.HumanChatMessage{Content: "Find the weather"},
		call,
		llms.ToolChatMessage{ID: "call-1", Content: "20 degrees"},
	}))
	model := &llmtest.MockLLM{}
	buffer := memory.NewConversationTokenBuffer(model, 100, memory.WithChatHistory(history))
	require.NoError(t, buffer.SaveContext(ctx, map[string]any{"input": "next"}, map[string]any{"output": "ok"}))
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	require.NotContains(t, messages, call, "tool arguments exceed the token budget")
	require.Equal(t, 0, model.CallCount)
	require.Equal(t, 0, model.GenerateCount)
}
