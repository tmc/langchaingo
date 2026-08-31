package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func captureDeepSeekRequest(t *testing.T, model string, opts ...llms.CallOption) map[string]any {
	t.Helper()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"1","object":"chat.completion","created":1,"model":"m",`+
			`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],`+
			`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
	}))
	defer server.Close()

	llm, err := New(WithBaseURL(server.URL), WithToken("token"), WithModel(model))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...); err != nil {
		t.Fatalf("generate: %v", err)
	}
	return body
}

func TestDeepSeekV32KeepsCallerSamplingUntilThinkingIsAsked(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"deepseek-v3.2", "deepseek.v3.2", "us.deepseek.v3.2", "deepseek-v3.2-exp"} {
		body := captureDeepSeekRequest(t, model, llms.WithTemperature(0.25), llms.WithTopP(0.3))

		if got, want := body["temperature"], 0.25; got != want {
			t.Errorf("%s: temperature = %v, want %v", model, got, want)
		}
		if got, ok := body["top_p"]; !ok || got != 0.3 {
			t.Errorf("%s: top_p = %v (present %v), want 0.3", model, got, ok)
		}
	}
}

func TestDeepSeekV32KeepsSamplingEvenWithAnEffortOnTheWire(t *testing.T) {
	t.Parallel()

	body := captureDeepSeekRequest(t, "deepseek-v3.2",
		llms.WithTemperature(0.25), llms.WithTopP(0.3),
		llms.WithReasoning(llms.ReasoningHigh, 0))

	if got, want := body["temperature"], 0.25; got != want {
		t.Errorf("temperature = %v, want %v", got, want)
	}
	if got, ok := body["top_p"]; !ok || got != 0.3 {
		t.Errorf("top_p = %v (present %v), want 0.3", got, ok)
	}
	if got, want := body["reasoning_effort"], "high"; got != want {
		t.Errorf("reasoning_effort = %v, want %v", got, want)
	}
}

func TestChatVariantsKeepTheTemperatureTheCallerSet(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"gpt-5.2-chat", "gpt-5.2-chat-latest", "gpt-5-chat"} {
		body := captureDeepSeekRequest(t, model,
			llms.WithTemperature(0.3), llms.WithReasoning(llms.ReasoningHigh, 0))
		if body["temperature"] != 0.3 {
			t.Errorf("%s must keep the caller's temperature, got body: %v", model, body)
		}
	}
}

func TestPenaltiesStayOffTheDoorsThatRefuseThem(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"grok-4.6", "grok-4-1-fast", "grok-4-1-fast-non-reasoning", "grok-3-mini-beta", "xai/grok-4.6",
	} {
		body := captureDeepSeekRequest(t, model,
			llms.WithFrequencyPenalty(0.5), llms.WithPresencePenalty(0.5))
		if _, ok := body["frequency_penalty"]; ok {
			t.Errorf("%s answers that it does not support the parameter, got body: %v", model, body)
		}
		if _, ok := body["presence_penalty"]; ok {
			t.Errorf("%s answers that it does not support the parameter, got body: %v", model, body)
		}
	}

	for _, model := range []string{"gpt-5.4", "deepseek-v4-pro", "glm-5.2", "mistral-medium-latest"} {
		body := captureDeepSeekRequest(t, model,
			llms.WithFrequencyPenalty(0.5), llms.WithPresencePenalty(0.5))
		if body["frequency_penalty"] != 0.5 || body["presence_penalty"] != 0.5 {
			t.Errorf("%s takes both penalties and must keep receiving them, got body: %v", model, body)
		}
	}
}

func TestStopStillTravelsToTheDoorsThatRefuseIt(t *testing.T) {
	t.Parallel()

	body := captureDeepSeekRequest(t, "grok-4.6", llms.WithStopWords([]string{"STOP"}))
	if _, ok := body["stop"]; !ok {
		t.Fatalf("stop bounds the answer, so it is left to fail loudly rather than dropped, got body: %v", body)
	}
}
