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

// captureMessagesRequest drives a real GenerateContent call against a recording
// server and returns the decoded outbound /v1/messages body and headers, so
// assertions cover generateMessagesContent (thinking construction, sampling-param
// handling, beta headers), not a hand-built payload.
func captureMessagesRequest(t *testing.T, callOpts ...llms.CallOption) (map[string]any, http.Header) {
	t.Helper()

	var (
		body   []byte
		header http.Header
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		header = r.Header.Clone()
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
	return payload, header
}

func TestAnthropic_AdaptiveThinkingRequest(t *testing.T) {
	t.Parallel()

	t.Run("adaptive carries output_config.effort and omits sampling params", func(t *testing.T) {
		t.Parallel()

		// Temperature/TopP are set NON-nil; the adaptive path must still drop them.
		payload, _ := captureMessagesRequest(t,
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

		payload, _ := captureMessagesRequest(t,
			llms.WithAdaptiveReasoning(llms.ReasoningNone),
			llms.WithMaxTokens(4096),
		)

		outputConfig, _ := payload["output_config"].(map[string]any)
		assert.Equal(t, "high", outputConfig["effort"])
	})

	t.Run("xhigh effort flows through to output_config", func(t *testing.T) {
		t.Parallel()

		payload, _ := captureMessagesRequest(t,
			llms.WithAdaptiveReasoning(llms.ReasoningXHigh),
			llms.WithMaxTokens(4096),
		)

		outputConfig, _ := payload["output_config"].(map[string]any)
		assert.Equal(t, "xhigh", outputConfig["effort"])
	})
}

func TestAnthropic_BudgetThinkingRequest(t *testing.T) {
	t.Parallel()

	payload, _ := captureMessagesRequest(t,
		llms.WithReasoning(llms.ReasoningMedium, 0),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(0.3),
		llms.WithTopP(0.9),
	)

	thinking, _ := payload["thinking"].(map[string]any)
	assert.Equal(t, "enabled", thinking["type"])
	assert.EqualValues(t, 2048, thinking["budget_tokens"]) // medium => max(4096/3, 2048)
	_, hasDisplay := thinking["display"]
	assert.False(t, hasDisplay, "budget thinking has no display field")

	_, hasOutputConfig := payload["output_config"]
	assert.False(t, hasOutputConfig, "budget thinking has no output_config")

	// Budget thinking pins temperature to 1.0 and keeps top_p as provided.
	assert.EqualValues(t, 1.0, payload["temperature"])
	assert.EqualValues(t, 0.9, payload["top_p"])
	assert.EqualValues(t, 4096, payload["max_tokens"]) // max(budget*2, maxTokens)
}

func TestAnthropic_InterleavedThinkingBetaHeader(t *testing.T) {
	t.Parallel()

	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "probe",
			Description: "test tool",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}}

	t.Run("budget reasoning with tools sends the beta header", func(t *testing.T) {
		t.Parallel()

		_, header := captureMessagesRequest(t,
			llms.WithReasoning(llms.ReasoningMedium, 0),
			llms.WithTools(tools),
			llms.WithMaxTokens(4096),
		)
		assert.Contains(t, header.Get("anthropic-beta"), "interleaved-thinking-2025-05-14")
	})

	t.Run("adaptive reasoning with tools sends no interleaved-thinking header", func(t *testing.T) {
		t.Parallel()

		_, header := captureMessagesRequest(t,
			llms.WithAdaptiveReasoning(llms.ReasoningHigh),
			llms.WithTools(tools),
			llms.WithMaxTokens(4096),
		)
		assert.NotContains(t, header.Get("anthropic-beta"), "interleaved-thinking")
	})
}
