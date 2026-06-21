package anthropic_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
)

// captureAdaptiveRequest drives a real GenerateContent call against a recording
// server and returns the decoded outbound /v1/messages body, so assertions cover
// generateMessagesContent (the adaptive branch + sampling-param omission), not a
// hand-built payload.
func captureAdaptiveRequest(t *testing.T, callOpts ...llms.CallOption) map[string]any {
	t.Helper()

	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_test","type":"message","role":"assistant","model":"claude-opus-4-6","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(
		anthropic.WithToken("test-key"),
		anthropic.WithBaseURL(srv.URL),
		anthropic.WithModel("claude-opus-4-6"),
	)
	require.NoError(t, err)

	messages := []llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("hi")},
	}}
	_, err = llm.GenerateContent(t.Context(), messages, callOpts...)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload
}

func TestAnthropic_AdaptiveThinkingRequest(t *testing.T) {
	t.Parallel()

	t.Run("adaptive carries output_config.effort and omits sampling params", func(t *testing.T) {
		t.Parallel()

		// Temperature/TopP are set NON-nil; the adaptive path must still drop them.
		payload := captureAdaptiveRequest(t,
			llms.WithAdaptiveReasoning(llms.ReasoningHigh),
			llms.WithTemperature(0.8),
			llms.WithTopP(0.9),
			llms.WithMaxTokens(4096),
		)

		thinking, _ := payload["thinking"].(map[string]any)
		assert.Equal(t, "adaptive", thinking["type"])
		assert.Equal(t, "summarized", thinking["display"])
		_, hasBudget := thinking["budget_tokens"]
		assert.False(t, hasBudget, "adaptive thinking must not carry a token budget")

		outputConfig, _ := payload["output_config"].(map[string]any)
		assert.Equal(t, "high", outputConfig["effort"])

		_, hasTemp := payload["temperature"]
		assert.False(t, hasTemp, "adaptive must omit temperature even when WithTemperature is set")
		_, hasTopP := payload["top_p"]
		assert.False(t, hasTopP, "adaptive must omit top_p even when WithTopP is set")
	})

	t.Run("empty adaptive effort defaults to high", func(t *testing.T) {
		t.Parallel()

		payload := captureAdaptiveRequest(t,
			llms.WithAdaptiveReasoning(llms.ReasoningNone),
			llms.WithMaxTokens(4096),
		)

		outputConfig, _ := payload["output_config"].(map[string]any)
		assert.Equal(t, "high", outputConfig["effort"])
	})

	t.Run("xhigh effort flows through to output_config", func(t *testing.T) {
		t.Parallel()

		payload := captureAdaptiveRequest(t,
			llms.WithAdaptiveReasoning(llms.ReasoningXHigh),
			llms.WithMaxTokens(4096),
		)

		outputConfig, _ := payload["output_config"].(map[string]any)
		assert.Equal(t, "xhigh", outputConfig["effort"])
	})
}
