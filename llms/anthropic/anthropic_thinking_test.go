package anthropic_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func newHTTPRRClient(t *testing.T, opts ...anthropic.Option) *anthropic.LLM {
	t.Helper()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
		return nil
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })

	allOpts := append([]anthropic.Option{
		anthropic.WithHTTPClient(rr.Client()),
	}, opts...)

	llm, err := anthropic.New(allOpts...)
	require.NoError(t, err)
	return llm
}

// captureMessagesRequest drives a real GenerateContent call against a recording
// server and returns the decoded outbound /v1/messages body and headers, so
// assertions cover generateMessagesContent (thinking construction, sampling-param
// handling, beta headers).
func captureMessagesRequest(t *testing.T, callOpts ...llms.CallOption) (map[string]any, http.Header) {
	t.Helper()
	return captureMessagesRequestModel(t, "claude-opus-4-6", callOpts...)
}

// captureMessagesRequestModel is captureMessagesRequest with an explicit client
// model; an empty model creates the client without one, so the request runs on
// the package default and exercises default-model capability resolution.
func captureMessagesRequestModel(t *testing.T, model string, callOpts ...llms.CallOption) (map[string]any, http.Header) {
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

	newOpts := []anthropic.Option{
		anthropic.WithToken("test-key"),
		anthropic.WithBaseURL(srv.URL),
	}
	if model != "" {
		newOpts = append(newOpts, anthropic.WithModel(model))
	}
	llm, err := anthropic.New(newOpts...)
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

// TestAnthropic_AdaptiveThinkingRealAPI verifies the real backend accepts the
// adaptive-thinking wire shape (thinking.type=adaptive + output_config.effort,
// no sampling params) that the unit tests only assert we construct: the request
// must succeed and return an answer. Reasoning is asserted only when present —
// adaptive thinking may legitimately skip thinking on a trivial prompt, so its
// absence is not a failure.
func TestAnthropic_AdaptiveThinkingRealAPI(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-5"))

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Solve: If x + 5 = 12, what is x?")},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithAdaptiveReasoning(llms.ReasoningHigh),
		llms.WithMaxTokens(4096),
	)
	require.NoError(t, err)

	choice := resp.Choices[0]
	assert.Contains(t, choice.Content, "7")
	if choice.Reasoning != nil {
		assert.NotEmpty(t, choice.Reasoning.Signature, "a returned thinking block must carry its signature")
	}
}

// TestAnthropic_ReasoningDisabledRealAPI verifies the real backend accepts an
// explicit disable on a model that thinks by default (Sonnet 5): the request
// must succeed and run without thinking rather than 400.
func TestAnthropic_ReasoningDisabledRealAPI(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-5"))

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Reply with just the word: ready")},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithReasoningDisabled(),
		llms.WithMaxTokens(64),
	)
	require.NoError(t, err)

	choice := resp.Choices[0]
	assert.NotEmpty(t, choice.Content)
	if choice.Reasoning != nil {
		assert.Empty(t, choice.Reasoning.Content, "disabled thinking must not return reasoning content")
	}
}

// TestAnthropic_TextResponseWithThinking tests text response with thinking and roundtrip
func TestAnthropic_TextResponseWithThinking(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Solve: If x + 5 = 12, what is x?"),
			},
		},
	}

	// Request with thinking
	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	choice := resp.Choices[0]
	assert.NotNil(t, choice.Reasoning)
	assert.NotEmpty(t, choice.Reasoning.Content)
	assert.NotEmpty(t, choice.Reasoning.Signature) // KEY ASSERTION
	assert.Contains(t, choice.Content, "7")

	// ROUNDTRIP: Continue conversation with preserved signature
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now solve x + 10 = 25")},
		},
	)

	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)
	assert.Contains(t, resp2.Choices[0].Content, "15")
}

// TestAnthropic_TextResponseWithThinkingStreaming tests streaming with thinking
func TestAnthropic_TextResponseWithThinkingStreaming(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 5 + 3? Think step by step."),
			},
		},
	}

	var streamedReasoning []string
	var streamedText []string

	streamFunc := llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
		switch chunk.Type {
		case streaming.ChunkTypeReasoning:
			if chunk.Reasoning != nil {
				streamedReasoning = append(streamedReasoning, chunk.Reasoning.Content)
			}
		case streaming.ChunkTypeText:
			streamedText = append(streamedText, chunk.Content)
		default:
		}
		return nil
	})

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		streamFunc,
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	// Verify accumulated content matches final response
	assert.Equal(t, strings.Join(streamedText, ""), resp.Choices[0].Content)
	if len(streamedReasoning) > 0 {
		assert.Equal(t, strings.Join(streamedReasoning, ""), resp.Choices[0].Reasoning.Content)
	}
	assert.NotEmpty(t, resp.Choices[0].Reasoning.Signature)
}

// TestAnthropic_SingleToolCallWithThinking tests tool call with thinking
func TestAnthropic_SingleToolCallWithThinking(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "Get weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{"type": "string"},
				},
				"required": []string{"location"},
			},
		},
	}}

	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(), // Explicit beta header
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0), // Required for thinking
	)
	require.NoError(t, err)

	// Reasoning is in choice.Reasoning, not in tool calls
	require.NotEmpty(t, resp.Choices[0].ToolCalls)
	choice := resp.Choices[0]

	assert.NotNil(t, choice.Reasoning)
	assert.NotEmpty(t, choice.Reasoning.Signature)

	// ROUNDTRIP: Send response back preserving reasoning
	// Use TextPartWithReasoning even with empty content to preserve thinking blocks
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
				resp.Choices[0].ToolCalls[0], // Add tool call
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: resp.Choices[0].ToolCalls[0].ID,
					Name:       resp.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		},
	)

	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0), // Required for thinking
	)
	require.NoError(t, err)
	assert.Contains(t, resp2.Choices[0].Content, "72")
}

// TestAnthropic_ParallelToolCallsWithThinking tests parallel tool calls
func TestAnthropic_ParallelToolCallsWithThinking(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{"type": "string"},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_time",
				Description: "Get current time for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{"type": "string"},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the weather and time in Boston?"),
			},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0), // Required for thinking
	)
	require.NoError(t, err)

	require.Equal(t, 2, len(resp.Choices[0].ToolCalls))

	// Reasoning is in choice, not in individual tool calls
	choice := resp.Choices[0]
	assert.NotNil(t, choice.Reasoning, "Response should have reasoning")
	if choice.Reasoning != nil {
		assert.NotEmpty(t, choice.Reasoning.Signature, "Reasoning should have signature")
	}
}

// TestAnthropic_SequentialToolCallsWithThinking tests sequential tool calls
func TestAnthropic_SequentialToolCallsWithThinking(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform a calculation",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{"type": "string"},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Calculate (5 + 3) and then multiply by 2"),
			},
		},
	}

	// First call
	resp1, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0), // Required for thinking
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.Choices[0].ToolCalls)
	require.NotNil(t, resp1.Choices[0].Reasoning)

	sig1 := resp1.Choices[0].Reasoning.Signature

	// Execute first tool and continue
	choice1 := resp1.Choices[0]
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice1.Content, choice1.Reasoning),
				choice1.ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: resp1.Choices[0].ToolCalls[0].ID,
					Name:       resp1.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content:    "8",
				},
			},
		},
	)

	// Second call
	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0), // Required for thinking
	)
	require.NoError(t, err)

	if resp2.Choices[0].Reasoning != nil {
		sig2 := resp2.Choices[0].Reasoning.Signature

		// Signatures should differ (each step has unique context)
		assert.NotEqual(t, string(sig1), string(sig2))
	}
}

// TestAnthropic_InterleavedThinkingMultipleBlocks tests multiple thinking blocks
func TestAnthropic_InterleavedThinkingMultipleBlocks(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "search",
				Description: "Search for information",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{"type": "string"},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform calculation",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{"type": "string"},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Search for the capital of France, then calculate its population density"),
			},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 4096),
		anthropic.WithInterleavedThinking(),
		llms.WithMaxTokens(8192),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	// In interleaved thinking, reasoning is in choice, not in tool calls
	choice := resp.Choices[0]
	if len(choice.ToolCalls) > 0 {
		t.Logf("Got %d tool calls", len(choice.ToolCalls))

		// Verify response has reasoning
		assert.NotNil(t, choice.Reasoning, "Response should have reasoning")
		if choice.Reasoning != nil {
			assert.NotEmpty(t, choice.Reasoning.Signature, "Reasoning should have signature")
		}

		// Log tool calls
		for i, toolCall := range choice.ToolCalls {
			t.Logf("Tool call %d: %s", i, toolCall.FunctionCall.Name)
		}
	}
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

// TestAnthropic_DefaultModelCapabilityResolution locks that a client created
// without a model classifies against the package default (claude-sonnet-5,
// adaptive-only and default-on) — the same model the request runs on — rather
// than an empty string that would read as unknown.
func TestAnthropic_DefaultModelCapabilityResolution(t *testing.T) {
	t.Parallel()

	t.Run("sampling dropped on the adaptive-only default", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequestModel(t, "",
			llms.WithTemperature(0.7), llms.WithTopP(0.9), llms.WithMaxTokens(64))
		_, hasTemp := p["temperature"]
		assert.False(t, hasTemp, "adaptive-only default must drop temperature")
		_, hasTopP := p["top_p"]
		assert.False(t, hasTopP, "adaptive-only default must drop top_p")
	})

	t.Run("budget request resolves to adaptive on the default", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequestModel(t, "",
			llms.WithReasoning(llms.ReasoningMedium, 0), llms.WithMaxTokens(4096))
		th, _ := p["thinking"].(map[string]any)
		assert.Equal(t, "adaptive", th["type"], "budget on adaptive-only default upgrades to adaptive")
	})

	t.Run("disable resolves to thinking disabled on the default-on default", func(t *testing.T) {
		t.Parallel()
		p, _ := captureMessagesRequestModel(t, "",
			llms.WithReasoningDisabled(), llms.WithMaxTokens(64))
		th, _ := p["thinking"].(map[string]any)
		assert.Equal(t, "disabled", th["type"], "default-on default must send thinking:{disabled}")
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

// An explicit ReasoningOn with no effort/tokens must not reach the API as an
// enabled-but-empty thinking block; it defaults to a positive budget.
func TestAnthropic_EmptyReasoningOnDefaultsToBudget(t *testing.T) {
	t.Parallel()

	onWithoutEffort := llms.CallOption(func(o *llms.CallOptions) {
		o.Reasoning = &llms.ReasoningConfig{Mode: llms.ReasoningOn}
	})
	p, _ := captureMessagesRequestModel(t, "claude-sonnet-4-5", onWithoutEffort, llms.WithMaxTokens(4096))

	th, _ := p["thinking"].(map[string]any)
	assert.Equal(t, "enabled", th["type"])
	budget, ok := th["budget_tokens"].(float64)
	assert.True(t, ok && budget > 0, "empty ReasoningOn must carry a positive budget, got %v", th["budget_tokens"])
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

// Haiku 4.5 rejects temperature and top_p together; without thinking the SDK must
// send at most one (keeping temperature) rather than both.
func TestAnthropic_Haiku45MutuallyExclusiveSampling(t *testing.T) {
	t.Parallel()

	p, _ := captureMessagesRequestModel(t, "claude-haiku-4-5",
		llms.WithTemperature(0.5),
		llms.WithTopP(0.9),
		llms.WithMaxTokens(64),
	)

	assert.EqualValues(t, 0.5, p["temperature"], "temperature must be kept")
	_, hasTopP := p["top_p"]
	assert.False(t, hasTopP, "Haiku 4.5 must not send top_p together with temperature")
}

// Opus 4.5 uses budget thinking but also honors effort, so the budget path must
// carry output_config.effort instead of dropping it.
func TestAnthropic_BudgetEffortForOpus45(t *testing.T) {
	t.Parallel()

	payload, _ := captureMessagesRequestModel(t, "claude-opus-4-5",
		llms.WithReasoning(llms.ReasoningHigh, 0),
		llms.WithMaxTokens(4096),
	)

	thinking, _ := payload["thinking"].(map[string]any)
	assert.Equal(t, "enabled", thinking["type"], "Opus 4.5 uses budget thinking")
	budget, ok := thinking["budget_tokens"].(float64)
	assert.True(t, ok && budget > 0, "budget thinking must carry a positive budget")

	outputConfig, _ := payload["output_config"].(map[string]any)
	assert.Equal(t, "high", outputConfig["effort"], "Opus 4.5 budget path must keep effort")
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

	// The header must follow the resolved mechanism, not the raw Adaptive flag:
	// adaptive downgraded to budget on a budget-only model still needs it.
	t.Run("adaptive downgraded to budget on budget-only model sends the header", func(t *testing.T) {
		t.Parallel()

		_, header := captureMessagesRequestModel(t, "claude-haiku-4-5",
			llms.WithAdaptiveReasoning(llms.ReasoningHigh),
			llms.WithTools(tools),
			llms.WithMaxTokens(4096),
		)
		assert.Contains(t, header.Get("anthropic-beta"), "interleaved-thinking-2025-05-14")
	})

	// Budget upgraded to adaptive on an adaptive-only model must not send it.
	t.Run("budget upgraded to adaptive on adaptive-only model omits the header", func(t *testing.T) {
		t.Parallel()

		_, header := captureMessagesRequestModel(t, "claude-sonnet-5",
			llms.WithReasoning(llms.ReasoningMedium, 0),
			llms.WithTools(tools),
			llms.WithMaxTokens(4096),
		)
		assert.NotContains(t, header.Get("anthropic-beta"), "interleaved-thinking")
	})
}
