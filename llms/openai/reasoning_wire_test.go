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

func sendModernReasoningForWire(t *testing.T, effort llms.ReasoningEffort, tokens int) string {
	return sendModernReasoningForModel(t, "gpt-5", effort, tokens)
}

func sendModernReasoningForModel(t *testing.T, model string, effort llms.ReasoningEffort, tokens int) string {
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

	llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model),
		WithModernReasoningFormat(), WithUsingReasoningMaxTokens())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoning(effort, tokens)); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func sendModernReasoningBody(t *testing.T, model string, effort llms.ReasoningEffort, tokens, maxTokens int) string {
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

	llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model),
		WithModernReasoningFormat(), WithUsingReasoningMaxTokens())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoning(effort, tokens), llms.WithMaxTokens(maxTokens)); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func TestModernReasoningSendsNoBudgetItCannotCompute(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		effort  llms.ReasoningEffort
		tokens  int
		present string
		absent  string
	}{
		{
			name:    "an effort with no budget mapping falls back to the effort itself",
			effort:  llms.ReasoningEffort("minimal"),
			tokens:  -5,
			present: `"effort":"minimal"`,
			absent:  `"max_tokens":-`,
		},
		{
			name:    "a negative budget under a mapped effort uses the effort's budget",
			effort:  llms.ReasoningMedium,
			tokens:  -5,
			present: `"max_tokens":5461`,
			absent:  `"effort"`,
		},
		{
			name:    "an explicit budget is sent as given",
			effort:  llms.ReasoningMedium,
			tokens:  2048,
			present: `"max_tokens":2048`,
			absent:  `"effort"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := sendModernReasoningForWire(t, tc.effort, tc.tokens)
			if !strings.Contains(body, tc.present) {
				t.Errorf("want %s on the wire, got body: %s", tc.present, body)
			}
			if strings.Contains(body, tc.absent) {
				t.Errorf("want no %s on the wire, got body: %s", tc.absent, body)
			}
		})
	}
}

func TestTheBudgetOnTheWireRespectsTheVendorFloor(t *testing.T) {
	t.Parallel()

	body := sendModernReasoningForModel(t, "claude-sonnet-4-5", llms.ReasoningNone, 100)
	if !strings.Contains(body, `"max_tokens":1024`) {
		t.Fatalf("a budget under the vendor floor must be raised to it, got body: %s", body)
	}

	untouched := sendModernReasoningForModel(t, "gpt-5", llms.ReasoningNone, 100)
	if !strings.Contains(untouched, `"max_tokens":100`) {
		t.Fatalf("the floor is Anthropic's alone, got body: %s", untouched)
	}
}

func TestABudgetTravelsToADoorThatRejectsTheEffortField(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"qwen3-235b-a22b-thinking-2507", "qwq-32b"} {
		t.Run(model, func(t *testing.T) {
			t.Parallel()

			body := sendModernReasoningForModel(t, model, llms.ReasoningNone, 4000)

			if !strings.Contains(body, `"reasoning":{"max_tokens":4000}`) {
				t.Fatalf("an explicit budget must reach a door that only rejects the effort field; body: %s", body)
			}
			if strings.Contains(body, `"effort"`) {
				t.Fatalf("the effort field itself must stay off a door that rejects it; body: %s", body)
			}
		})
	}
}

func TestTheEffortFieldStaysOffADoorThatRejectsIt(t *testing.T) {
	t.Parallel()

	body := sendModernReasoningForModel(t, "qwq-32b", llms.ReasoningHigh, 0)

	if strings.Contains(body, `"effort"`) {
		t.Fatalf("an effort-only request must not put the field on a door that rejects it; body: %s", body)
	}
}

func TestBothDeepSeekReasonerNamesRefuseTheOffTheSameWay(t *testing.T) {
	for _, model := range []string{"deepseek-reasoner", "deepseek-r1"} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","created":1,"model":"m",` +
				`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
				`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
		}))
		t.Cleanup(srv.Close)

		llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model))
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		_, err = llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithReasoningDisabled())
		if err == nil {
			t.Fatalf("%s reasons whatever is asked, so the disable must return a typed error", model)
		}
		if !strings.Contains(err.Error(), "reasoning cannot be disabled") {
			t.Fatalf("%s: expected the typed disable error, got: %v", model, err)
		}
	}
}

func TestABudgetLargerThanTheAnswerLimitRaisesTheLimit(t *testing.T) {
	body := sendModernReasoningBody(t, "anthropic/claude-opus-4-5", llms.ReasoningNone, 100, 512)

	if !strings.Contains(body, `"max_tokens":1024`) && !strings.Contains(body, `"max_completion_tokens":1024`) {
		t.Fatalf("the floor raised the budget to 1024, so the answer limit must follow: %s", body)
	}
	if strings.Contains(body, `"max_completion_tokens":512`) || strings.Contains(body, `"max_tokens":512`) {
		t.Fatalf("a 512-token answer limit leaves no room once a 1024-token budget is spent: %s", body)
	}
}

func captureWireWithClient(t *testing.T, model string, clientOpts []Option, callOpts ...llms.CallOption) string {
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

	opts := append([]Option{WithBaseURL(srv.URL), WithToken("test"), WithModel(model)}, clientOpts...)
	llm, err := New(opts...)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, callOpts...); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func TestAnswerLimitClearsTheBudgetTheThinkingWillSpend(t *testing.T) {
	t.Parallel()

	const claude = "anthropic/claude-sonnet-4-5"
	modern := []Option{WithModernReasoningFormat()}
	budgetField := []Option{WithModernReasoningFormat(), WithUsingReasoningMaxTokens()}

	for _, tc := range []struct {
		name   string
		client []Option
		call   []llms.CallOption
		want   string
	}{
		{"legacy format, effort carries the thinking", nil,
			[]llms.CallOption{llms.WithReasoning("", 2048), llms.WithMaxTokens(512)}, `"max_completion_tokens":8192`},
		{"modern format, effort carries the thinking", modern,
			[]llms.CallOption{llms.WithReasoning("", 2048), llms.WithMaxTokens(512)}, `"max_completion_tokens":8192`},
		{"an effort with no budget of its own", nil,
			[]llms.CallOption{llms.WithReasoning(llms.ReasoningHigh, 0), llms.WithMaxTokens(512)}, `"max_completion_tokens":8192`},
		{"a lower effort stands for a smaller budget", nil,
			[]llms.CallOption{llms.WithReasoning(llms.ReasoningLow, 0), llms.WithMaxTokens(512)}, `"max_completion_tokens":2048`},
		{"modern format, budget field carries the thinking", budgetField,
			[]llms.CallOption{llms.WithReasoning("", 2048), llms.WithMaxTokens(512)}, `"max_completion_tokens":2048`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := captureWireWithClient(t, claude, tc.client, tc.call...)
			if strings.Contains(body, `"max_completion_tokens":512`) {
				t.Errorf("a 512-token answer limit leaves no room once the thinking is paid for: %s", body)
			}
			if !strings.Contains(body, tc.want) {
				t.Errorf("want %s on the wire, got body: %s", tc.want, body)
			}
		})
	}
}

func TestAnswerLimitStaysWhereNothingSpendsABudget(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		model string
		call  []llms.CallOption
	}{
		{"a thinking model of another vendor", "gpt-5",
			[]llms.CallOption{llms.WithReasoning(llms.ReasoningHigh, 0), llms.WithMaxTokens(512)}},
		{"no reasoning asked at all", "anthropic/claude-sonnet-4-5",
			[]llms.CallOption{llms.WithMaxTokens(512)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := captureWireWithClient(t, tc.model, nil, tc.call...)
			if !strings.Contains(body, `"max_completion_tokens":512`) {
				t.Errorf("nothing spends a budget here, so the limit must reach the wire verbatim: %s", body)
			}
		})
	}
}

func TestRaisingTheLimitDoesNotEditTheCallersOption(t *testing.T) {
	t.Parallel()

	reused := []llms.CallOption{llms.WithReasoning("", 2048), llms.WithMaxTokens(512)}
	client := []Option{WithModernReasoningFormat(), WithUsingReasoningMaxTokens()}

	first := captureWireWithClient(t, "anthropic/claude-sonnet-4-5", client, reused...)
	second := captureWireWithClient(t, "anthropic/claude-sonnet-4-5", client, reused...)

	if first != second {
		t.Errorf("one call must not edit the option the next one reuses:\nfirst:  %s\nsecond: %s", first, second)
	}
	if !strings.Contains(second, `"reasoning":{"max_tokens":1024}`) {
		t.Errorf("want the vendor floor budget on the reused call, got: %s", second)
	}
}

func TestVerbosityReachesTheWireOnlyWhenAsked(t *testing.T) {
	t.Parallel()

	capture := func(t *testing.T, opts ...llms.CallOption) string {
		t.Helper()

		var body string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			body = string(raw)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","created":1,"model":"m",` +
				`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
				`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
		}))
		t.Cleanup(srv.Close)

		llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel("gpt-5.6"))
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	t.Run("asked", func(t *testing.T) {
		t.Parallel()

		if body := capture(t, llms.WithVerbosity("low")); !strings.Contains(body, `"verbosity":"low"`) {
			t.Fatalf("verbosity must reach the wire, got body: %s", body)
		}
	})

	t.Run("not asked", func(t *testing.T) {
		t.Parallel()

		if body := capture(t); strings.Contains(body, "verbosity") {
			t.Fatalf("an unset verbosity must stay off the wire, got body: %s", body)
		}
	})
}

func TestTheOutputLimitPicksTheFieldTheRouteUnderstands(t *testing.T) {
	t.Parallel()

	capture := func(t *testing.T, model string, opts ...llms.CallOption) string {
		t.Helper()

		var body string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			body = string(raw)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","created":1,"model":"m",` +
				`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
				`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
		}))
		t.Cleanup(srv.Close)

		llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model))
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		all := append([]llms.CallOption{llms.WithMaxTokens(800)}, opts...)
		if _, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, all...); err != nil {
			t.Fatalf("GenerateContent() error: %v", err)
		}
		return body
	}

	for _, model := range []string{
		"grok-4-1-fast", "qwen3.5-35b-a3b", "dashscope/qwen3.7-plus", "qwq-32b",
		"deepseek-v4-pro", "deepseek-v4-flash", "deepseek/deepseek-v4-pro", "deepseek-r1",
	} {
		body := capture(t, model)
		if !strings.Contains(body, `"max_tokens":800`) {
			t.Errorf("%s must receive max_tokens, got body: %s", model, body)
		}
		if strings.Contains(body, "max_completion_tokens") {
			t.Errorf("%s must not receive max_completion_tokens, got body: %s", model, body)
		}
	}

	for _, model := range []string{"gpt-5.4", "gpt-4.1", "o3", "claude-sonnet-4-5", "glm-5.2", "kimi-k3"} {
		body := capture(t, model)
		if !strings.Contains(body, `"max_completion_tokens":800`) {
			t.Errorf("%s must keep max_completion_tokens, got body: %s", model, body)
		}
	}

	body := capture(t, "gpt-5.4", WithLegacyMaxTokensField())
	if !strings.Contains(body, `"max_tokens":800`) {
		t.Errorf("the explicit option must still win, got body: %s", body)
	}
}

func sendDoorDefaultReasoning(t *testing.T, model string, effort llms.ReasoningEffort) string {
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
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoning(effort, 0)); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func TestTheChatCompletionsDoorSendsTheFlatEffortField(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"gpt-5", "anthropic/claude-sonnet-5", "deepseek-v4-pro"} {
		t.Run(model, func(t *testing.T) {
			t.Parallel()

			body := sendDoorDefaultReasoning(t, model, llms.ReasoningHigh)

			if !strings.Contains(body, `"reasoning_effort":"high"`) {
				t.Fatalf("effort must travel as the flat field, got body: %s", body)
			}
			if strings.Contains(body, `"reasoning":{`) {
				t.Fatalf("the nested reasoning object belongs to another door, got body: %s", body)
			}
		})
	}
}
