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

	for _, model := range []string{"kimi-k2.5", "qwen3-235b-a22b-thinking-2507", "qwq-32b"} {
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

	body := sendModernReasoningForModel(t, "kimi-k2.5", llms.ReasoningHigh, 0)

	if strings.Contains(body, `"effort"`) {
		t.Fatalf("an effort-only request must not put the field on a door that rejects it; body: %s", body)
	}
}

func disableReasoningForModel(t *testing.T, model string) string {
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
		llms.WithReasoningDisabled()); err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	return body
}

func TestBothDeepSeekReasonerNamesTurnThinkingOffTheSameWay(t *testing.T) {
	byOwnName := disableReasoningForModel(t, "deepseek-reasoner")
	byFamilyName := disableReasoningForModel(t, "deepseek-r1")

	for _, want := range []string{`"reasoning_effort":"none"`} {
		if !strings.Contains(byFamilyName, want) {
			t.Fatalf("deepseek-r1 request lost %s: %s", want, byFamilyName)
		}
		if !strings.Contains(byOwnName, want) {
			t.Fatalf("deepseek-reasoner must disable thinking like deepseek-r1, got: %s", byOwnName)
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
