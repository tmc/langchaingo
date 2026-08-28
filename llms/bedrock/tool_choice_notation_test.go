package bedrock_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
)

func converseChoiceOnWire(t *testing.T, choice any) string {
	t.Helper()

	const answer = `{"output":{"message":{"role":"assistant","content":[{"text":"ok"}]}},` +
		`"stopReason":"end_turn","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`

	tool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "weather",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}

	var body string
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, answer)
		}))
		t.Cleanup(srv.Close)

		llm := bedrockLLMAgainst(t, srv,
			bedrock.WithModel("amazon.nova-lite-v1:0"), bedrock.WithConverseAPI())

		opts := []llms.CallOption{llms.WithTools([]llms.Tool{tool})}
		if choice != nil {
			opts = append(opts, llms.WithToolChoice(choice))
		}
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}
}

func TestConverseKeepsAForcedChoiceInEveryNotation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		choice any
		want   string
	}{
		{
			"named in the Anthropic spelling",
			llms.ToolChoice{Type: "tool", Function: &llms.FunctionReference{Name: "get_weather"}},
			`"toolChoice":{"tool":{"name":"get_weather"}}`,
		},
		{
			"named in the OpenAI spelling",
			llms.ToolChoice{Type: "function", Function: &llms.FunctionReference{Name: "get_weather"}},
			`"toolChoice":{"tool":{"name":"get_weather"}}`,
		},
		{
			"named as a raw OpenAI map",
			map[string]any{"type": "function", "function": map[string]any{"name": "get_weather"}},
			`"toolChoice":{"tool":{"name":"get_weather"}}`,
		},
		{
			"any tool",
			llms.ToolChoice{Type: "any"},
			`"toolChoice":{"any":{}}`,
		},
		{
			"required, the OpenAI spelling of any tool",
			"required",
			`"toolChoice":{"any":{}}`,
		},
		{
			"an unset choice stays with the model",
			nil,
			`"toolChoice":{"auto":{}}`,
		},
		{
			"an explicit auto stays with the model",
			llms.ToolChoice{Type: "auto"},
			`"toolChoice":{"auto":{}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body := converseChoiceOnWire(t, tc.choice)
			if !strings.Contains(body, tc.want) {
				t.Fatalf("wire must carry %s, got body: %s", tc.want, body)
			}
		})
	}
}

func TestConverseOffersNoToolWhenTheCallerForbidsThem(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		choice any
	}{
		{"none as a bare string", "none"},
		{"none as a struct", llms.ToolChoice{Type: "none"}},
		{"none as a raw map", map[string]any{"type": "none"}},
	} {
		t.Run("forbidding tools offers the model none: "+tc.name, func(t *testing.T) {
			t.Parallel()

			body := converseChoiceOnWire(t, tc.choice)
			if strings.Contains(body, `"toolConfig"`) {
				t.Fatalf("a forbidden tool must not be offered at all, got body: %s", body)
			}
		})
	}
}
