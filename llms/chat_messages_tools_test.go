package llms_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestGetBufferStringPreservesToolCalls(t *testing.T) {
	t.Parallel()
	ai := llms.AIChatMessage{
		Content: "Checking",
		ToolCalls: []llms.ToolCall{
			{ID: "call-weather", Type: "function", FunctionCall: &llms.FunctionCall{
				Name: "weather", Arguments: `{"city":"Paris"}`,
			}},
			{ID: "call-time", Type: "function", FunctionCall: &llms.FunctionCall{
				Name: "time", Arguments: `{"zone":"Europe/Paris"}`,
			}},
		},
	}
	for name, message := range map[string]llms.ChatMessage{"value": ai, "pointer": &ai} {
		t.Run(name, func(t *testing.T) {
			buffer, err := llms.GetBufferString([]llms.ChatMessage{message}, "Human", "Assistant")
			require.NoError(t, err)
			const prefix = "Assistant: Checking "
			require.True(t, strings.HasPrefix(buffer, prefix))
			var restored []llms.ToolCall
			require.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(buffer, prefix)), &restored))
			require.Equal(t, ai.ToolCalls, restored)
		})
	}
}

func TestGetBufferStringLegacyFunctionCall(t *testing.T) {
	t.Parallel()
	buffer, err := llms.GetBufferString([]llms.ChatMessage{llms.AIChatMessage{
		Content: "Checking", FunctionCall: &llms.FunctionCall{Name: "lookup", Arguments: `{"city":"Paris"}`},
	}}, "Human", "Assistant")
	require.NoError(t, err)
	require.Equal(t, `Assistant: Checking {"name":"lookup","arguments":"{\"city\":\"Paris\"}"}`, buffer)
}
