package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func weatherTool() llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:       "get_weather",
			Parameters: map[string]any{"type": "object"},
		},
	}
}

func bodyForCall(t *testing.T, model string, opts ...llms.CallOption) string {
	t.Helper()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, completion)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func TestEffortAndToolsOnTheWire(t *testing.T) {
	t.Parallel()

	tools := llms.WithTools([]llms.Tool{weatherTool()})
	high := llms.WithReasoning(llms.ReasoningHigh, 0)

	for _, tc := range []struct {
		name   string
		model  string
		opts   []llms.CallOption
		want   string
		absent bool
	}{
		{"5.6 с инструментами и уровнем деградирует до none", "gpt-5.6-sol",
			[]llms.CallOption{tools, high}, `"reasoning_effort":"none"`, false},
		{"5.6 с инструментами без просьбы всё равно шлёт none", "gpt-5.6-sol",
			[]llms.CallOption{tools}, `"reasoning_effort":"none"`, false},
		{"5.6 без инструментов сохраняет уровень", "gpt-5.6-sol",
			[]llms.CallOption{high}, `"reasoning_effort":"high"`, false},
		{"5.5 с инструментами не шлёт поле", "gpt-5.5",
			[]llms.CallOption{tools, high}, `"reasoning_effort"`, true},
		{"5.4-nano с инструментами не шлёт поле", "gpt-5.4-nano",
			[]llms.CallOption{tools, high}, `"reasoning_effort"`, true},
		{"5.5 без инструментов сохраняет уровень", "gpt-5.5",
			[]llms.CallOption{high}, `"reasoning_effort":"high"`, false},
		{"5.2 с инструментами сохраняет уровень", "gpt-5.2",
			[]llms.CallOption{tools, high}, `"reasoning_effort":"high"`, false},
		{"gpt-4o не получает поле вовсе", "gpt-4o",
			[]llms.CallOption{high}, `"reasoning_effort"`, true},
		{"gpt-4o с инструментами тоже не получает", "gpt-4o",
			[]llms.CallOption{tools, high}, `"reasoning_effort"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := bodyForCall(t, tc.model, tc.opts...)
			if got := strings.Contains(body, tc.want); got == tc.absent {
				verb := "не содержит"
				if tc.absent {
					verb = "содержит"
				}
				t.Errorf("тело запроса %s %s\nтело: %s", verb, tc.want, body)
			}
		})
	}
}
