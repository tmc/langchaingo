package bedrock_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
)

func TestLegacyDoorCarriesTheToolChoice(t *testing.T) {
	t.Parallel()

	const answer = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	tool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "weather",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}

	choiceOnWire := func(t *testing.T, choice any) any {
		t.Helper()

		llm, sent := legacyLLMCapturing(t, answer,
			bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

		opts := []llms.CallOption{llms.WithTools([]llms.Tool{tool})}
		if choice != nil {
			opts = append(opts, llms.WithToolChoice(choice))
		}
		_, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...)
		require.NoError(t, err)

		var body map[string]any
		require.NoError(t, json.Unmarshal([]byte(*sent), &body))
		return body["tool_choice"]
	}

	for _, tc := range []struct {
		name   string
		choice any
		want   any
	}{
		{
			"named in the Anthropic spelling",
			llms.ToolChoice{Type: "tool", Function: &llms.FunctionReference{Name: "get_weather"}},
			map[string]any{"type": "tool", "name": "get_weather"},
		},
		{
			"named in the OpenAI spelling",
			llms.ToolChoice{Type: "function", Function: &llms.FunctionReference{Name: "get_weather"}},
			map[string]any{"type": "tool", "name": "get_weather"},
		},
		{"any tool", llms.ToolChoice{Type: "any"}, map[string]any{"type": "any"}},
		{"required", "required", map[string]any{"type": "any"}},
		{"auto", "auto", map[string]any{"type": "auto"}},
		{"none", "none", map[string]any{"type": "none"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, choiceOnWire(t, tc.choice))
		})
	}

	t.Run("no choice sends no field", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, choiceOnWire(t, nil))
	})
}
