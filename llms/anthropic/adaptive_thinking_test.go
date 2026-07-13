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
// handling, beta headers).
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

// TestAnthropic_ModelAwareThinkingResolution locks the deterministic
// (model, request) -> wire mechanism table: a currently-accepted combination
// keeps its mechanism, and a combination the model would reject is redirected
// to the mechanism it accepts.
func TestAnthropic_ModelAwareThinkingResolution(t *testing.T) {
	t.Parallel()

	thinkingType := func(p map[string]any) string {
		th, _ := p["thinking"].(map[string]any)
		s, _ := th["type"].(string)
		return s
	}

	t.Run("budget on adaptive-only model upgrades to adaptive", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-sonnet-5"),
			llms.WithReasoning(llms.ReasoningMedium, 0),
			llms.WithTemperature(0.7), llms.WithTopP(0.9), llms.WithMaxTokens(4096))
		assert.Equal(t, "adaptive", thinkingType(p))
		th, _ := p["thinking"].(map[string]any)
		_, hasBudget := th["budget_tokens"]
		assert.False(t, hasBudget, "upgraded adaptive must not carry a budget")
		_, hasTemp := p["temperature"]
		assert.False(t, hasTemp, "adaptive-only model must drop temperature")
		_, hasTopP := p["top_p"]
		assert.False(t, hasTopP, "adaptive-only model must drop top_p")
	})

	t.Run("adaptive on budget-only model downgrades to budget", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-haiku-4-5"),
			llms.WithAdaptiveReasoning(llms.ReasoningHigh), llms.WithMaxTokens(4096))
		assert.Equal(t, "enabled", thinkingType(p))
		th, _ := p["thinking"].(map[string]any)
		_, hasBudget := th["budget_tokens"]
		assert.True(t, hasBudget, "downgraded budget must carry a token budget")
		_, hasOutputConfig := p["output_config"]
		assert.False(t, hasOutputConfig, "budget thinking has no output_config")
	})

	t.Run("budget on budget-only model stays budget", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-sonnet-4-5"),
			llms.WithReasoning(llms.ReasoningMedium, 0), llms.WithMaxTokens(4096))
		assert.Equal(t, "enabled", thinkingType(p))
	})

	t.Run("adaptive on dual model stays adaptive", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-opus-4-6"),
			llms.WithAdaptiveReasoning(llms.ReasoningHigh), llms.WithMaxTokens(2048))
		assert.Equal(t, "adaptive", thinkingType(p))
	})

	t.Run("sampling on adaptive-only model without reasoning is dropped", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-sonnet-5"),
			llms.WithTemperature(0.7), llms.WithTopP(0.9), llms.WithMaxTokens(64))
		_, hasThinking := p["thinking"]
		assert.False(t, hasThinking, "no reasoning requested")
		_, hasTemp := p["temperature"]
		assert.False(t, hasTemp, "adaptive-only model drops temperature even without thinking")
		_, hasTopP := p["top_p"]
		assert.False(t, hasTopP, "adaptive-only model drops top_p even without thinking")
	})

	t.Run("sampling on budget-capable model without reasoning is kept", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-sonnet-4-5"),
			llms.WithTemperature(0.7), llms.WithMaxTokens(64))
		assert.EqualValues(t, 0.7, p["temperature"], "budget-capable model keeps temperature")
	})
}

// TestAnthropic_ReasoningDisabled locks the explicit-off wire: a default-on model
// sends thinking:{disabled}, a default-off model omits thinking, and a model whose
// thinking cannot be disabled fails with a typed error before any request is sent.
func TestAnthropic_ReasoningDisabled(t *testing.T) {
	t.Parallel()

	t.Run("default-on model sends thinking disabled", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-sonnet-5"),
			llms.WithReasoningDisabled(), llms.WithMaxTokens(64))
		th, _ := p["thinking"].(map[string]any)
		assert.Equal(t, "disabled", th["type"])
		_, hasBudget := th["budget_tokens"]
		assert.False(t, hasBudget, "disabled thinking carries no budget")
	})

	t.Run("default-off model omits thinking", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequest(t,
			llms.WithModel("claude-opus-4-8"),
			llms.WithReasoningDisabled(), llms.WithMaxTokens(64))
		_, hasThinking := p["thinking"]
		assert.False(t, hasThinking, "off on a default-off model omits thinking")
	})

	t.Run("always-on model errors before sending", func(t *testing.T) {
		t.Parallel()
		called := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		llm, err := anthropic.New(
			anthropic.WithToken("test-key"),
			anthropic.WithBaseURL(srv.URL),
			anthropic.WithModel("claude-fable-5"),
		)
		require.NoError(t, err)

		messages := []llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("hi")}}}
		_, err = llm.GenerateContent(t.Context(), messages, llms.WithReasoningDisabled(), llms.WithMaxTokens(64))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be disabled")
		assert.False(t, called, "must not reach the API when off is unsupported")
	})
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

	// Budget thinking pins temperature to 1.0 and drops top_p — the API rejects
	// temperature and top_p together.
	assert.EqualValues(t, 1.0, payload["temperature"])
	_, hasTopP := payload["top_p"]
	assert.False(t, hasTopP, "budget thinking must drop top_p")
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
