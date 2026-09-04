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

	t.Run("o-series clamps xhigh to high", func(t *testing.T) {
		body := capture(t, "o3", llms.ReasoningXHigh)
		if !strings.Contains(body, `"reasoning_effort":"high"`) {
			t.Fatalf("o3 takes only low/medium/high, got body: %s", body)
		}
	})

	t.Run("gpt-5 clamps xhigh to high", func(t *testing.T) {
		body := capture(t, "gpt-5", llms.ReasoningXHigh)
		if !strings.Contains(body, `"reasoning_effort":"high"`) {
			t.Fatalf("gpt-5 tops out at high, got body: %s", body)
		}
	})

	t.Run("unknown model passes max through", func(t *testing.T) {
		body := capture(t, "gpt-5.7", llms.ReasoningMax)
		if !strings.Contains(body, `"reasoning_effort":"max"`) {
			t.Fatalf("unknown model must pass max through, got body: %s", body)
		}
	})
}

func TestReasoningDisabledForClaudeBehindOpenAITransport(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	send := func(t *testing.T, model string) (string, error) {
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
		_, genErr := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithReasoningDisabled())
		return body, genErr
	}

	t.Run("default-on Claude reports unsupported instead of a silent no-op", func(t *testing.T) {
		body, err := send(t, "anthropic/claude-sonnet-5")
		var unsupported *reasoning.ErrReasoningOffUnsupported
		if !errors.As(err, &unsupported) {
			t.Fatalf("want ErrReasoningOffUnsupported, got err=%v body=%s", err, body)
		}
	})

	t.Run("always-on Claude still reports unsupported", func(t *testing.T) {
		_, err := send(t, "anthropic/claude-fable-5")
		var unsupported *reasoning.ErrReasoningOffUnsupported
		if !errors.As(err, &unsupported) {
			t.Fatalf("want ErrReasoningOffUnsupported, got %v", err)
		}
	})

	t.Run("off-by-default Claude omits the field", func(t *testing.T) {
		body, err := send(t, "anthropic/claude-opus-4-7")
		if err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		if strings.Contains(body, "reasoning") {
			t.Fatalf("a model that does not think by default needs no disable token, got body: %s", body)
		}
	})
}

func TestTemperaturePinnedOnlyWhereTheModelDemandsIt(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	send := func(t *testing.T, model string) string {
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
			llms.WithTemperature(0.5)); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	for _, model := range []string{"gpt-5.1", "gpt-5.2", "gpt-5.4", "gpt-5.4-mini", "gpt-5.4-nano"} {
		t.Run("keeps/"+model, func(t *testing.T) {
			if body := send(t, model); !strings.Contains(body, `"temperature":0.5`) {
				t.Fatalf("%s takes the caller's temperature, got body: %s", model, body)
			}
		})
	}
	for _, model := range []string{"gpt-5", "gpt-5-mini", "gpt-5.5", "gpt-5.6-terra", "o3-mini"} {
		t.Run("pins/"+model, func(t *testing.T) {
			if body := send(t, model); !strings.Contains(body, `"temperature":1`) {
				t.Fatalf("%s only accepts the default temperature, got body: %s", model, body)
			}
		})
	}
}

func TestTopPDroppedOnlyWhereTheModelRejectsIt(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	send := func(t *testing.T, model string, opts ...llms.CallOption) string {
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
		opts = append([]llms.CallOption{llms.WithTopP(0.9)}, opts...)
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	for _, model := range []string{"gpt-4.1", "gpt-4o", "gpt-5.1", "gpt-5.2", "gpt-5.4-mini"} {
		t.Run("keeps/"+model, func(t *testing.T) {
			if body := send(t, model); !strings.Contains(body, `"top_p":0.9`) {
				t.Fatalf("%s takes the caller's top_p, got body: %s", model, body)
			}
		})
	}
	for _, model := range []string{"gpt-5", "gpt-5-mini", "gpt-5-nano", "o3", "o4-mini"} {
		t.Run("drops/"+model, func(t *testing.T) {
			if body := send(t, model); strings.Contains(body, "top_p") {
				t.Fatalf("%s rejects top_p outright, got body: %s", model, body)
			}
		})
		t.Run("drops-with-reasoning/"+model, func(t *testing.T) {
			body := send(t, model, llms.WithReasoning(llms.ReasoningHigh, 0))
			if strings.Contains(body, "top_p") {
				t.Fatalf("%s rejects top_p with reasoning too, got body: %s", model, body)
			}
		})
	}
}

func TestClaudeSamplingBehindTheOpenAITransport(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	send := func(t *testing.T, model string) string {
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
			llms.WithTemperature(0.5), llms.WithTopP(0.9)); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	for _, model := range []string{
		"anthropic/claude-sonnet-5", "anthropic/claude-opus-4-7", "anthropic/claude-fable-5",
	} {
		t.Run("drops-both/"+model, func(t *testing.T) {
			body := send(t, model)
			if strings.Contains(body, "temperature") || strings.Contains(body, "top_p") {
				t.Fatalf("%s rejects sampling outright, got body: %s", model, body)
			}
		})
	}

	for _, model := range []string{
		"anthropic/claude-haiku-4-5", "anthropic/claude-sonnet-4-5", "anthropic/claude-opus-4-5",
		"anthropic/claude-sonnet-4-6", "anthropic/claude-opus-4-6",
	} {
		t.Run("keeps-temp-drops-topp/"+model, func(t *testing.T) {
			body := send(t, model)
			if !strings.Contains(body, `"temperature":0.5`) {
				t.Fatalf("%s takes the caller's temperature, got body: %s", model, body)
			}
			if strings.Contains(body, "top_p") {
				t.Fatalf("%s rejects temperature and top_p together, got body: %s", model, body)
			}
		})
	}
}

func TestALevelOutsideTheVendorEnumDoesNotReachTheWire(t *testing.T) {
	t.Parallel()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	for _, tc := range []struct {
		model  string
		asked  llms.ReasoningEffort
		want   string
		reject string
	}{
		{"kimi-k3", llms.ReasoningMedium, `"reasoning_effort":"low"`, `"medium"`},
		{"glm-5.3", llms.ReasoningMedium, `"reasoning_effort":"low"`, `"medium"`},
		{"glm-5.3-flash", llms.ReasoningXHigh, `"reasoning_effort":"high"`, `"xhigh"`},
		{"deepseek-v4-pro", llms.ReasoningMedium, `"reasoning_effort":"medium"`, ""},
		{"glm-5.2", llms.ReasoningMedium, `"reasoning_effort":"medium"`, ""},
	} {
		t.Run(tc.model+"/"+string(tc.asked), func(t *testing.T) {
			t.Parallel()

			var body string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				body = string(b)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, completion)
			}))
			defer srv.Close()

			llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(tc.model))
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}
			if _, err := llm.GenerateContent(
				context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
				llms.WithReasoning(tc.asked, 0),
			); err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}

			if !strings.Contains(body, tc.want) {
				t.Fatalf("%s: wire body missing %s\nbody: %s", tc.model, tc.want, body)
			}
			if tc.reject != "" && strings.Contains(body, tc.reject) {
				t.Fatalf("%s: %s reached the wire although the vendor documents an enum without it\nbody: %s",
					tc.model, tc.reject, body)
			}
		})
	}
}
