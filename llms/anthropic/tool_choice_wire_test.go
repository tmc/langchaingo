package anthropic_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
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

	for _, tc := range []struct {
		name   string
		choice any
		want   map[string]any
	}{
		{"auto as a raw map", map[string]any{"type": "auto"}, map[string]any{"type": "auto"}},
		{"auto as a bare string", "auto", map[string]any{"type": "auto"}},
		{"auto as a struct", llms.ToolChoice{Type: "auto"}, map[string]any{"type": "auto"}},
		{"none as a bare string", "none", map[string]any{"type": "none"}},
		{"none as a struct", llms.ToolChoice{Type: "none"}, map[string]any{"type": "none"}},
	} {
		t.Run("a choice the model owns reaches the wire as an object: "+tc.name, func(t *testing.T) {
			t.Parallel()

			payload, _ := captureMessagesRequest(t,
				llms.WithTools([]llms.Tool{tool}),
				llms.WithToolChoice(tc.choice),
				llms.WithMaxTokens(64))

			assert.Equal(t, tc.want, payload["tool_choice"])
		})
	}

	t.Run("no choice sends no field", func(t *testing.T) {
		t.Parallel()

		payload, _ := captureMessagesRequest(t,
			llms.WithTools([]llms.Tool{tool}), llms.WithMaxTokens(64))

		_, present := payload["tool_choice"]
		assert.False(t, present)
	})
}

func TestAFinishedTurnWithNoBlockIsAnAnswer(t *testing.T) {
	t.Parallel()

	tool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "weather",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}

	answer := func(t *testing.T, body string) (*llms.ContentResponse, error) {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.Copy(io.Discard, r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, body)
		}))
		t.Cleanup(srv.Close)

		llm, err := anthropic.New(anthropic.WithToken("test-key"),
			anthropic.WithBaseURL(srv.URL), anthropic.WithModel("claude-haiku-4-5"))
		require.NoError(t, err)

		return llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithTools([]llms.Tool{tool}),
			llms.WithToolChoice("none"),
			llms.WithMaxTokens(512))
	}

	t.Run("no block but a stop reason answers", func(t *testing.T) {
		t.Parallel()

		resp, err := answer(t, `{"id":"msg_1","type":"message","role":"assistant",
			"model":"claude-haiku-4-5","content":[],"stop_reason":"end_turn",
			"usage":{"input_tokens":550,"output_tokens":9}}`)

		require.NoError(t, err)
		require.Len(t, resp.Choices, 1)
		assert.Empty(t, resp.Choices[0].Content)
		assert.Equal(t, "end_turn", resp.Choices[0].StopReason)
		assert.Equal(t, 9, resp.Choices[0].GenerationInfo["CompletionTokens"])
	})

	t.Run("no block and no stop reason is still empty", func(t *testing.T) {
		t.Parallel()

		_, err := answer(t, `{"id":"msg_2","type":"message","role":"assistant",
			"model":"claude-haiku-4-5","content":[],"stop_reason":"",
			"usage":{"input_tokens":1,"output_tokens":0}}`)

		require.ErrorIs(t, err, anthropic.ErrEmptyResponse)
	})
}

func TestFableTakesAForcedToolChoiceWithoutThinking(t *testing.T) {
	t.Parallel()

	const reply = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	var sent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sent = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("test"), anthropic.WithBaseURL(srv.URL),
		anthropic.WithModel("claude-fable-5"))
	require.NoError(t, err)

	tool := llms.Tool{Type: "function", Function: &llms.FunctionDefinition{
		Name: "calc", Description: "multiply",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{}},
	}}

	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "17*23?")},
		llms.WithTools([]llms.Tool{tool}), llms.WithToolChoice("any"))
	require.NoError(t, err, "the vendor answers 200 with a tool call for any, auto and a named "+
		"tool on this model, so a local refusal would take away working behaviour")
	assert.Contains(t, sent, `"tool_choice":{"type":"any"}`)
}
