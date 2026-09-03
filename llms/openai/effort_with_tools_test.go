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

func TestQwenThinkingBudgetOnTheWire(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		model  string
		opt    llms.CallOption
		want   string
		absent bool
	}{
		{"бюджет уходит вместо эффорта", "dashscope/qwen3.8-max",
			llms.WithReasoning(llms.ReasoningLow, 4096), `"thinking_budget":4096`, false},
		{"эффорт при бюджете не уходит", "dashscope/qwen3.8-max",
			llms.WithReasoning(llms.ReasoningLow, 4096), `"reasoning_effort"`, true},
		{"без бюджета уходит эффорт", "dashscope/qwen3.8-max",
			llms.WithReasoning(llms.ReasoningLow, 0), `"reasoning_effort":"low"`, false},
		{"без бюджета поля бюджета нет", "dashscope/qwen3.8-max",
			llms.WithReasoning(llms.ReasoningLow, 0), `"thinking_budget"`, true},
		{"3.7 тоже берёт бюджет", "dashscope/qwen3.7-plus",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen-plus берёт бюджет", "dashscope/qwen-plus",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen-flash берёт бюджет", "dashscope/qwen-flash",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen3-max берёт бюджет", "dashscope/qwen3-max",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen3-vl-plus берёт бюджет", "dashscope/qwen3-vl-plus",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen3-vl-flash берёт бюджет", "dashscope/qwen3-vl-flash",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen3-next берёт бюджет", "dashscope/qwen3-next-80b-a3b-thinking",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"qwen-turbo берёт бюджет", "dashscope/qwen-turbo",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"открытые веса VL берут бюджет", "dashscope/qwen3-vl-30b-a3b",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"крупные открытые веса VL тоже", "dashscope/qwen3-vl-235b-a22b",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget":2048`, false},
		{"недумающий next бюджета не получает", "dashscope/qwen3-next-80b-a3b-instruct",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget"`, true},
		{"флаг остаётся вместе с бюджетом", "dashscope/qwen-plus",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"enable_thinking":true`, false},
		{"чужой хост бюджета не получает", "openrouter/qwen/qwen3.7-plus",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget"`, true},
		{"не qwen бюджета не получает", "gpt-5.2",
			llms.WithReasoning(llms.ReasoningLow, 2048), `"thinking_budget"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := bodyForCall(t, tc.model, llms.WithMaxTokens(8000), tc.opt)
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
