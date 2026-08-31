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

func TestReasoningEffortOmittedOnDoorsThatRejectIt(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	capture := func(t *testing.T, model string, opt llms.CallOption) (string, error) {
		t.Helper()
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
		_, err = llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opt)
		return body, err
	}

	for _, model := range []string{"qwen3-next-80b-a3b-thinking"} {
		t.Run(model+" drops a requested effort", func(t *testing.T) {
			body, err := capture(t, model, llms.WithReasoning(llms.ReasoningMedium, 0))
			if err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}
			if strings.Contains(body, "reasoning_effort") {
				t.Fatalf("%s must not receive reasoning_effort, got body: %s", model, body)
			}
		})
	}

	for _, model := range []string{"kimi-k2.7-code-highspeed", "kimi-k2-thinking", "qwen3-next-80b-a3b-thinking"} {
		t.Run(model+" reports off as unsupported", func(t *testing.T) {
			_, err := capture(t, model, llms.WithReasoningDisabled())
			if err == nil {
				t.Fatalf("%s cannot disable thinking and must return a typed error", model)
			}
			if !strings.Contains(err.Error(), "reasoning cannot be disabled") {
				t.Fatalf("expected the typed disable error, got: %v", err)
			}
		})
	}

	for _, model := range []string{"kimi-k2.6", "kimi-k2.7-code-highspeed"} {
		t.Run(model+" carries the effort to the wire", func(t *testing.T) {
			body, err := capture(t, model, llms.WithReasoning(llms.ReasoningMedium, 0))
			if err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}
			if !strings.Contains(body, `"reasoning_effort":"medium"`) {
				t.Fatalf("%s accepts reasoning_effort and must receive it, got body: %s", model, body)
			}
		})
	}

	t.Run("kimi-k3 still receives the effort", func(t *testing.T) {
		body, err := capture(t, "kimi-k3", llms.WithReasoning(llms.ReasoningMedium, 0))
		if err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		if !strings.Contains(body, `"reasoning_effort":"medium"`) {
			t.Fatalf("kimi-k3 accepts reasoning_effort and must receive it, got body: %s", body)
		}
	})

	t.Run("kimi-k3 still disables via none", func(t *testing.T) {
		body, err := capture(t, "kimi-k3", llms.WithReasoningDisabled())
		if err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		if !strings.Contains(body, `"reasoning_effort":"none"`) {
			t.Fatalf("kimi-k3 disables via the none token, got body: %s", body)
		}
	})
}
