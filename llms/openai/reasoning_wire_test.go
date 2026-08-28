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
