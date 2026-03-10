package anthropic_test

import (
	"context"
	"net/http"
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
