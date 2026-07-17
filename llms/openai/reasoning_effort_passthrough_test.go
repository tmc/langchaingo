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

// WithReasoningDisabled must send the model's explicit disable token on a
// reasoning model, error on a model that cannot disable, and stay silent on a
// non-reasoning model.
func TestReasoningDisabledToWire(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	newLLM := func(t *testing.T, model string, modern bool) (*LLM, *string) {
		t.Helper()
		body := new(string)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			*body = string(b)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, completion)
		}))
		t.Cleanup(srv.Close)
		opts := []Option{WithBaseURL(srv.URL), WithToken("test"), WithModel(model)}
		if modern {
			opts = append(opts, WithModernReasoningFormat())
		}
		llm, err := New(opts...)
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		return llm, body
	}

	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	t.Run("legacy disable sends reasoning_effort none", func(t *testing.T) {
		llm, body := newLLM(t, "gpt-5.5", false)
		if _, err := llm.GenerateContent(context.Background(), msgs, llms.WithReasoningDisabled()); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		if !strings.Contains(*body, `"reasoning_effort":"none"`) {
			t.Fatalf("disable must send reasoning_effort=none, got body: %s", *body)
		}
	})

	t.Run("modern disable sends effort none", func(t *testing.T) {
		llm, body := newLLM(t, "gpt-5.5", true)
		if _, err := llm.GenerateContent(context.Background(), msgs, llms.WithReasoningDisabled()); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		if !strings.Contains(*body, `"effort":"none"`) {
			t.Fatalf("disable must send effort=none, got body: %s", *body)
		}
	})

	t.Run("mandatory reasoning model rejects disable", func(t *testing.T) {
		llm, _ := newLLM(t, "o3-mini", false)
		_, err := llm.GenerateContent(context.Background(), msgs, llms.WithReasoningDisabled())
		if err == nil || !strings.Contains(err.Error(), "cannot be disabled") {
			t.Fatalf("o3-mini disable must error with 'cannot be disabled', got: %v", err)
		}
	})

	t.Run("gpt-5-pro rejects disable", func(t *testing.T) {
		llm, _ := newLLM(t, "gpt-5-pro", false)
		_, err := llm.GenerateContent(context.Background(), msgs, llms.WithReasoningDisabled())
		if err == nil || !strings.Contains(err.Error(), "cannot be disabled") {
			t.Fatalf("gpt-5-pro disable must error with 'cannot be disabled', got: %v", err)
		}
	})

	t.Run("non-reasoning model omits reasoning", func(t *testing.T) {
		llm, body := newLLM(t, "gpt-4.1", false)
		if _, err := llm.GenerateContent(context.Background(), msgs, llms.WithReasoningDisabled()); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		if strings.Contains(*body, "reasoning_effort") || strings.Contains(*body, `"reasoning"`) {
			t.Fatalf("non-reasoning model must omit reasoning fields, got body: %s", *body)
		}
	})
}

// An effort above a model's ceiling must be clamped to what the model accepts
// before it reaches the wire, rather than sent verbatim and rejected with a 400.
func TestReasoningEffortClampedPerModel(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	capture := func(t *testing.T, model string, effort llms.ReasoningEffort) string {
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
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithReasoning(effort, 0)); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	t.Run("gpt-5.4-mini clamps max to xhigh", func(t *testing.T) {
		body := capture(t, "gpt-5.4-mini", llms.ReasoningMax)
		if !strings.Contains(body, `"reasoning_effort":"xhigh"`) {
			t.Fatalf("max must clamp to xhigh on gpt-5.4-mini, got body: %s", body)
		}
	})

	t.Run("gpt-5-pro pins low to high", func(t *testing.T) {
		body := capture(t, "gpt-5-pro", llms.ReasoningLow)
		if !strings.Contains(body, `"reasoning_effort":"high"`) {
			t.Fatalf("gpt-5-pro must pin effort to high, got body: %s", body)
		}
	})

	t.Run("unknown model passes max through", func(t *testing.T) {
		body := capture(t, "gpt-5.6", llms.ReasoningMax)
		if !strings.Contains(body, `"reasoning_effort":"max"`) {
			t.Fatalf("unknown model must pass max through, got body: %s", body)
		}
	})
}
