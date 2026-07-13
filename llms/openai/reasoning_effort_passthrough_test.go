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

// xhigh/max effort must reach the request body unchanged (not downgraded to
// high) on both the legacy reasoning_effort and modern reasoning.effort formats.
func TestReasoningEffortPassthroughToWire(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	cases := []struct {
		name   string
		effort llms.ReasoningEffort
		modern bool
		want   string
	}{
		{"legacy xhigh", llms.ReasoningXHigh, false, `"reasoning_effort":"xhigh"`},
		{"legacy max", llms.ReasoningMax, false, `"reasoning_effort":"max"`},
		{"modern xhigh", llms.ReasoningXHigh, true, `"effort":"xhigh"`},
		{"modern max", llms.ReasoningMax, true, `"effort":"max"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				body = string(b)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, completion)
			}))
			defer srv.Close()

			opts := []Option{WithBaseURL(srv.URL), WithToken("test"), WithModel("reasoning-model")}
			if tc.modern {
				opts = append(opts, WithModernReasoningFormat())
			}
			llm, err := New(opts...)
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			if _, err := llm.GenerateContent(
				context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
				llms.WithReasoning(tc.effort, 0),
			); err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}

			if !strings.Contains(body, tc.want) {
				t.Fatalf("wire body missing %s\nbody: %s", tc.want, body)
			}
		})
	}
}

// WithAdaptiveReasoning with no explicit effort must not silently disable
// reasoning here: it defaults to high, matching the Anthropic/Bedrock paths.
func TestAdaptiveReasoningNoEffortDefaultsToHighOnWire(t *testing.T) {
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
	defer srv.Close()

	llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel("reasoning-model"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, err := llm.GenerateContent(
		context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithAdaptiveReasoning(llms.ReasoningNone),
	); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	if !strings.Contains(body, `"reasoning_effort":"high"`) {
		t.Fatalf("adaptive-no-effort must send reasoning_effort=high, got body: %s", body)
	}
}
