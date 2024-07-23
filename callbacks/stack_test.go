package callbacks

import (
	"context"
	"fmt"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type MyCustomHandler struct {
	name string
	ch   chan string
}

var _ Handler = (*MyCustomHandler)(nil)

func NewMyCustomHandler(name string, ch chan string) *MyCustomHandler {
	return &MyCustomHandler{
		name: name,
		ch:   ch,
	}
}

func (m *MyCustomHandler) HandleText(context.Context, string) {
	m.ch <- fmt.Sprintf("[HandleText] %s", m.name)
}

func (m *MyCustomHandler) HandleLLMStart(context.Context, []string) {
	m.ch <- fmt.Sprintf("[HandleLLMStart] %s", m.name)
}

func (m *MyCustomHandler) HandleLLMGenerateContentStart(context.Context, []llms.MessageContent) {
	m.ch <- fmt.Sprintf("[HandleLLMGenerateContentStart] %s", m.name)
}

func (m *MyCustomHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	m.ch <- fmt.Sprintf("[HandleLLMGenerateContentEnd] %s", m.name)
}

func (m *MyCustomHandler) HandleLLMError(context.Context, error) {
	m.ch <- fmt.Sprintf("[HandleLLMError] %s", m.name)
}

func (m *MyCustomHandler) HandleChainStart(context.Context, map[string]any) {
	m.ch <- fmt.Sprintf("[HandleChainStart] %s", m.name)
}

func (m *MyCustomHandler) HandleChainEnd(context.Context, map[string]any) {
	m.ch <- fmt.Sprintf("[HandleChainEnd] %s", m.name)
}

func (m *MyCustomHandler) HandleChainError(context.Context, error) {
	m.ch <- fmt.Sprintf("[HandleChainError] %s", m.name)
}

func (m *MyCustomHandler) HandleToolStart(context.Context, string) {
	m.ch <- fmt.Sprintf("[HandleToolStart] %s", m.name)
}

func (m *MyCustomHandler) HandleToolEnd(context.Context, string) {
	m.ch <- fmt.Sprintf("[HandleToolEnd] %s", m.name)
}

func (m *MyCustomHandler) HandleToolError(context.Context, error) {
	m.ch <- fmt.Sprintf("[HandleToolError] %s", m.name)
}

func (m *MyCustomHandler) HandleAgentAction(context.Context, schema.AgentAction) {
	m.ch <- fmt.Sprintf("[HandleAgentAction] %s", m.name)
}

func (m *MyCustomHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	m.ch <- fmt.Sprintf("[HandleAgentFinish] %s", m.name)
}

func (m *MyCustomHandler) HandleRetrieverStart(ctx context.Context, query string) {
	m.ch <- fmt.Sprintf("[HandleRetrieverStart] %s", m.name)
}

func (m *MyCustomHandler) HandleRetrieverEnd(context.Context, string, []schema.Document) {
	m.ch <- fmt.Sprintf("[HandleRetrieverEnd] %s", m.name)
}

func (m *MyCustomHandler) HandleStreamingFunc(context.Context, []byte) {
	m.ch <- fmt.Sprintf("[HandleStreamingFunc] %s", m.name)
}

func TestStackHandler(t *testing.T) {
	ch := make(chan string, 2)
	defer close(ch)

	h := NewStackHandler(
		&SimpleHandler{},
		NewMyCustomHandler("my-custom-handler-1", ch),
		NewMyCustomHandler("my-custom-handler-2", ch),
	)

	tests := []struct {
		name string
		fn   func()
	}{
		{"HandleText", func() { h.HandleText(context.Background(), "text") }},
		{"HandleLLMStart", func() { h.HandleLLMStart(context.Background(), []string{"prompt"}) }},
		{"HandleLLMGenerateContentStart", func() { h.HandleLLMGenerateContentStart(context.Background(), nil) }},
		{"HandleLLMGenerateContentEnd", func() { h.HandleLLMGenerateContentEnd(context.Background(), nil) }},
		{"HandleLLMError", func() { h.HandleLLMError(context.Background(), fmt.Errorf("error")) }},
		{"HandleChainStart", func() { h.HandleChainStart(context.Background(), map[string]any{"input": "value"}) }},
		{"HandleChainEnd", func() { h.HandleChainEnd(context.Background(), map[string]any{"output": "value"}) }},
		{"HandleChainError", func() { h.HandleChainError(context.Background(), fmt.Errorf("error")) }},
		{"HandleToolStart", func() { h.HandleToolStart(context.Background(), "input") }},
		{"HandleToolEnd", func() { h.HandleToolEnd(context.Background(), "output") }},
		{"HandleToolError", func() { h.HandleToolError(context.Background(), fmt.Errorf("error")) }},
		{"HandleAgentAction", func() { h.HandleAgentAction(context.Background(), schema.AgentAction{}) }},
		{"HandleAgentFinish", func() { h.HandleAgentFinish(context.Background(), schema.AgentFinish{}) }},
		{"HandleRetrieverStart", func() { h.HandleRetrieverStart(context.Background(), "query") }},
		{"HandleRetrieverEnd", func() { h.HandleRetrieverEnd(context.Background(), "query", nil) }},
		{"HandleStreamingFunc", func() { h.HandleStreamingFunc(context.Background(), nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn()

			for _, name := range []string{
				"my-custom-handler-1",
				"my-custom-handler-2",
			} {
				if got := <-ch; got != fmt.Sprintf("[%s] %s", tt.name, name) {
					t.Errorf("unexpected value: got %q", got)
				}
			}
		})
	}
}
