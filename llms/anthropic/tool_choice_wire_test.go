package anthropic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestToolChoiceReachesTheWireInTheMessagesSpelling(t *testing.T) {
	t.Parallel()

	tool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "weather",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}

	for _, tc := range []struct {
		name   string
		choice any
		want   map[string]any
	}{
		{
			"named in the Anthropic spelling",
			llms.ToolChoice{Type: "tool", Function: &llms.FunctionReference{Name: "get_weather"}},
			map[string]any{"type": "tool", "name": "get_weather"},
		},
		{
			"named as a raw map",
			map[string]any{"type": "tool", "name": "get_weather"},
			map[string]any{"type": "tool", "name": "get_weather"},
		},
		{
			"named in the OpenAI spelling",
			llms.ToolChoice{Type: "function", Function: &llms.FunctionReference{Name: "get_weather"}},
			map[string]any{"type": "tool", "name": "get_weather"},
		},
		{
			"any tool",
			llms.ToolChoice{Type: "any"},
			map[string]any{"type": "any"},
		},
		{
			"required, the OpenAI spelling of any tool",
			"required",
			map[string]any{"type": "any"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payload, _ := captureMessagesRequest(t,
				llms.WithTools([]llms.Tool{tool}),
				llms.WithToolChoice(tc.choice),
				llms.WithMaxTokens(64))

			assert.Equal(t, tc.want, payload["tool_choice"])
		})
	}

	t.Run("a choice the model owns is left alone", func(t *testing.T) {
		t.Parallel()

		payload, _ := captureMessagesRequest(t,
			llms.WithTools([]llms.Tool{tool}),
			llms.WithToolChoice(map[string]any{"type": "auto"}),
			llms.WithMaxTokens(64))

		assert.Equal(t, map[string]any{"type": "auto"}, payload["tool_choice"])
	})

	t.Run("no choice sends no field", func(t *testing.T) {
		t.Parallel()

		payload, _ := captureMessagesRequest(t,
			llms.WithTools([]llms.Tool{tool}), llms.WithMaxTokens(64))

		_, present := payload["tool_choice"]
		assert.False(t, present)
	})
}
