package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func captureChatRequest(t *testing.T, callOpts ...llms.CallOption) map[string]any {
	t.Helper()

	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
	}))
	t.Cleanup(srv.Close)

	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))
	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, callOpts...); err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	return payload
}

func TestToolChoiceReachesTheWireInTheChatCompletionsSpelling(t *testing.T) {
	t.Parallel()

	tool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "weather",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}
	named := map[string]any{"type": "function", "function": map[string]any{"name": "get_weather"}}

	for _, tc := range []struct {
		name   string
		choice any
		want   any
	}{
		{
			"named in the OpenAI spelling",
			llms.ToolChoice{Type: "function", Function: &llms.FunctionReference{Name: "get_weather"}},
			named,
		},
		{
			"named in the Anthropic spelling",
			llms.ToolChoice{Type: "tool", Function: &llms.FunctionReference{Name: "get_weather"}},
			named,
		},
		{
			"named as a raw Anthropic map",
			map[string]any{"type": "tool", "name": "get_weather"},
			named,
		},
		{
			"any tool, the Anthropic spelling",
			llms.ToolChoice{Type: "any"},
			"required",
		},
		{
			"required",
			"required",
			"required",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payload := captureChatRequest(t,
				llms.WithTools([]llms.Tool{tool}), llms.WithToolChoice(tc.choice))

			if got := payload["tool_choice"]; !jsonEqual(got, tc.want) {
				t.Fatalf("tool_choice = %#v, want %#v", got, tc.want)
			}
		})
	}

	for _, tc := range []struct {
		name   string
		choice any
		want   string
	}{
		{"auto as a bare string", "auto", "auto"},
		{"auto as a struct", llms.ToolChoice{Type: "auto"}, "auto"},
		{"auto as a raw map", map[string]any{"type": "auto"}, "auto"},
		{"none as a bare string", "none", "none"},
		{"none as a struct", llms.ToolChoice{Type: "none"}, "none"},
	} {
		t.Run("a choice the model owns reaches the wire as a string: "+tc.name, func(t *testing.T) {
			t.Parallel()

			payload := captureChatRequest(t,
				llms.WithTools([]llms.Tool{tool}), llms.WithToolChoice(tc.choice))

			if got := payload["tool_choice"]; got != tc.want {
				t.Fatalf("tool_choice = %#v, want %q", got, tc.want)
			}
		})
	}

	t.Run("no choice sends no field", func(t *testing.T) {
		t.Parallel()

		payload := captureChatRequest(t, llms.WithTools([]llms.Tool{tool}))

		if _, present := payload["tool_choice"]; present {
			t.Fatalf("tool_choice must stay absent, got %#v", payload["tool_choice"])
		}
	})
}

func jsonEqual(a, b any) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(left) == string(right)
}
