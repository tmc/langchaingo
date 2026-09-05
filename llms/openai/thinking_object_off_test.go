package openai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestDisablingThinkingWhereTheEffortTokenDoesNotReachTheVendor(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"glm-5.1", "glm-5.2", "glm-4.6", "minimax-m3", "kimi-k2.6",
		"deepseek-v4-flash", "deepseek-v4-pro",
	} {
		body := sendForWire(t, model, llms.WithReasoningDisabled())
		if !strings.Contains(body, `"thinking":{"type":"disabled"}`) {
			t.Errorf("%s: body carries no disabled thinking object: %s", model, body)
		}
		if strings.Contains(body, "reasoning_effort") {
			t.Errorf("%s: body still carries reasoning_effort: %s", model, body)
		}
	}
}

func TestDisablingThinkingOnAModelThatOnlyThinks(t *testing.T) {
	t.Parallel()

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

	llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel("qwen3.8-2.4t-a95b"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoningDisabled())

	var unsupported *reasoning.ErrReasoningOffUnsupported
	if !errors.As(err, &unsupported) {
		t.Fatalf("want ErrReasoningOffUnsupported, got err=%v body=%s", err, body)
	}
	if body != "" {
		t.Errorf("the refusal must come before the network, but a request went out: %s", body)
	}
}
