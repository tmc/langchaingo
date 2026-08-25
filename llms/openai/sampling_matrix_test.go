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

func sendForWire(t *testing.T, model string, opts ...llms.CallOption) string {
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

type samplingCase struct {
	name    string
	model   string
	opts    []llms.CallOption
	present []string
	absent  []string
}

func runSamplingCases(t *testing.T, cases []samplingCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := sendForWire(t, tc.model, tc.opts...)
			for _, want := range tc.present {
				if !strings.Contains(body, want) {
					t.Errorf("want %s on the wire, got body: %s", want, body)
				}
			}
			for _, unwanted := range tc.absent {
				if strings.Contains(body, unwanted) {
					t.Errorf("want no %s on the wire, got body: %s", unwanted, body)
				}
			}
		})
	}
}

func TestSamplingMatrix(t *testing.T) {
	t.Parallel()

	temp := llms.WithTemperature(0.3)
	topP := llms.WithTopP(0.9)
	high := llms.WithReasoning(llms.ReasoningHigh, 0)
	off := llms.WithReasoningDisabled()

	runSamplingCases(t, []samplingCase{{
		name:    "accepting model keeps its temperature while thinking",
		model:   "gpt-5.4",
		opts:    []llms.CallOption{temp, topP, high},
		present: []string{`"temperature":0.3`, `"reasoning_effort":"high"`},
		absent:  []string{`"top_p"`},
	}, {
		name:    "non-accepting model is pinned while thinking",
		model:   "gpt-5.5",
		opts:    []llms.CallOption{temp, topP, high},
		present: []string{`"temperature":1`, `"reasoning_effort":"high"`},
		absent:  []string{`"top_p"`},
	}, {
		name:    "disabled thinking leaves both sampling params alone",
		model:   "gpt-5.5",
		opts:    []llms.CallOption{temp, topP, off},
		present: []string{`"temperature":0.3`, `"top_p":0.9`, `"reasoning_effort":"none"`},
	}, {
		name:    "omitted effort falls back to the model for an accepting name",
		model:   "gpt-5.4",
		opts:    []llms.CallOption{temp, topP},
		present: []string{`"temperature":0.3`, `"top_p":0.9`},
		absent:  []string{`"reasoning_effort"`},
	}, {
		name:    "omitted effort pins a name that thinks by default",
		model:   "gpt-5.5",
		opts:    []llms.CallOption{temp, topP},
		present: []string{`"temperature":1`},
		absent:  []string{`"top_p"`, `"reasoning_effort"`},
	}, {
		name:    "a non-reasoning model keeps everything",
		model:   "gpt-4.1",
		opts:    []llms.CallOption{temp, topP},
		present: []string{`"temperature":0.3`, `"top_p":0.9`},
	}})
}

func TestClaudeSamplingMatrix(t *testing.T) {
	t.Parallel()

	temp := llms.WithTemperature(0.3)
	topP := llms.WithTopP(0.9)
	high := llms.WithReasoning(llms.ReasoningHigh, 0)

	runSamplingCases(t, []samplingCase{{
		name:   "a model that rejects sampling gets neither param",
		model:  "claude-opus-4-7",
		opts:   []llms.CallOption{temp, topP},
		absent: []string{`"temperature"`, `"top_p"`},
	}, {
		name:    "mutually exclusive sampling drops top_p and keeps temperature",
		model:   "claude-opus-4-5",
		opts:    []llms.CallOption{temp, topP},
		present: []string{`"temperature":0.3`},
		absent:  []string{`"top_p"`},
	}, {
		name:    "an omitted effort leaves a request-only thinker unpinned",
		model:   "claude-opus-4-5",
		opts:    []llms.CallOption{temp},
		present: []string{`"temperature":0.3`},
		absent:  []string{`"reasoning_effort"`},
	}, {
		name:    "thinking pins the temperature even when top_p forced a drop",
		model:   "claude-sonnet-4-5",
		opts:    []llms.CallOption{temp, topP, high},
		present: []string{`"temperature":1`, `"reasoning_effort":"high"`},
		absent:  []string{`"top_p"`},
	}, {
		name:    "an effort above the generation's ceiling is lowered",
		model:   "claude-opus-4-6",
		opts:    []llms.CallOption{llms.WithReasoning(llms.ReasoningXHigh, 0)},
		present: []string{`"reasoning_effort":"high"`},
		absent:  []string{`"reasoning_effort":"xhigh"`},
	}, {
		name:    "extra body overwrites a param the policy stripped",
		model:   "claude-opus-4-7",
		opts:    []llms.CallOption{temp, topP, WithExtraBody(map[string]any{"temperature": 0.7})},
		present: []string{`"temperature":0.7`},
		absent:  []string{`"top_p"`},
	}})
}

func TestAModelThatRejectsSamplingGetsNoSamplingAtAll(t *testing.T) {
	t.Parallel()

	body := sendForWire(t, "claude-opus-4-7",
		llms.WithTemperature(0.3), llms.WithTopP(0.9), llms.WithTopK(40), llms.WithMinP(0.05))

	for _, field := range []string{`"temperature"`, `"top_p"`, `"top_k"`, `"min_p"`} {
		if strings.Contains(body, field) {
			t.Errorf("want no %s on the wire, got body: %s", field, body)
		}
	}
}

func TestOptInThinkingKeepsSamplingUntilAsked(t *testing.T) {
	t.Parallel()

	runSamplingCases(t, []samplingCase{
		{
			name:    "no effort leaves the caller's sampling alone",
			model:   "mistral-medium-3",
			opts:    []llms.CallOption{llms.WithTemperature(0.3), llms.WithTopP(0.7)},
			present: []string{`"temperature":0.3`, `"top_p":0.7`},
		},
		{
			name:  "an effort puts the request on the thinking terms",
			model: "mistral-medium-3",
			opts: []llms.CallOption{
				llms.WithTemperature(0.3), llms.WithTopP(0.7), llms.WithReasoning(llms.ReasoningMedium, 0),
			},
			present: []string{`"temperature":1`, `"reasoning_effort":"medium"`},
			absent:  []string{`"top_p"`},
		},
		{
			name:    "the provider-prefixed spelling resolves the same",
			model:   "mistralai/mistral-medium-3",
			opts:    []llms.CallOption{llms.WithTemperature(0.3)},
			present: []string{`"temperature":0.3`},
		},
	})
}
