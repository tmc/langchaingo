package googleai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleAI_TextResponseWithSignature(t *testing.T) {
	// Use Gemini 3 model which supports thought signatures
	llm := newHTTPRRClient(t, WithDefaultModel("gemini-3-flash-preview"))

	// Step 1: Get initial response with thinking and signature
	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 15 + 27? Think step by step."),
			},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithMaxTokens(2048),
		llms.WithReasoning(llms.ReasoningMedium, 1024),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	choice := resp.Choices[0]
	assert.Contains(t, choice.Content, "42")

	// For Gemini 3, reasoning should be in ContentChoice (not in ToolCalls)
	assert.Empty(t, choice.ToolCalls, "Expected no tool calls for text-only response")

	// Gemini 3 should return reasoning with signature for text responses
	assert.NotNil(t, choice.Reasoning, "Expected reasoning to be present for Gemini 3")
	assert.NotEmpty(t, choice.Reasoning.Content, "Expected reasoning content to be present")
	assert.NotEmpty(t, choice.Reasoning.Signature, "Expected thought signature for Gemini 3 text response")
	t.Logf("Thought signature: %d bytes, Reasoning content: %d bytes",
		len(choice.Reasoning.Signature), len(choice.Reasoning.Content))

	// Step 2: Test roundtrip - send follow-up message with signature preserved
	// According to docs, signatures in text responses are recommended but not required
	// For Gemini 3, we should preserve the signature in the text response
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
		},
	})
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart("Now multiply that result by 2"),
		},
	})

	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithMaxTokens(1024),
	)
	require.NoError(t, err, "Roundtrip should work with preserved text signature")
	require.NotEmpty(t, resp2.Choices)
	assert.Contains(t, resp2.Choices[0].Content, "84")
	t.Log("Successfully completed text response roundtrip with signature")
}

func TestGoogleAI_TextResponseWithSignatureStreaming(t *testing.T) {
	// Use Gemini 3 model which supports thought signatures
	llm := newHTTPRRClient(t, WithDefaultModel("gemini-3-flash-preview"))

	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Calculate 23 * 17. Show your work."),
			},
		},
	}

	var accumulatedContent, accumulatedReasoning string
	var streamDone bool

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithMaxTokens(2048),
		llms.WithReasoning(llms.ReasoningMedium, 1024),
		llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			switch chunk.Type {
			case streaming.ChunkTypeText:
				accumulatedContent += chunk.Content
			case streaming.ChunkTypeReasoning:
				if chunk.Reasoning != nil {
					accumulatedReasoning += chunk.Reasoning.Content
				} else {
					t.Log("reasoning chunk is nil")
				}
			case streaming.ChunkTypeDone:
				streamDone = true
			default:
			}
			return nil
		}),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	choice := resp.Choices[0]
	assert.True(t, streamDone, "Expected streaming to complete")
	assert.Equal(t, accumulatedContent, choice.Content, "Streamed content should match final content")

	// Verify reasoning is in ContentChoice (not in ToolCalls)
	assert.Empty(t, choice.ToolCalls, "Expected no tool calls for text-only response")

	// Gemini 3 should return reasoning with signature in streaming mode
	assert.NotNil(t, choice.Reasoning, "Expected reasoning to be present for Gemini 3")
	assert.NotEmpty(t, choice.Reasoning.Content, "Expected reasoning content")
	assert.NotEmpty(t, choice.Reasoning.Signature, "Expected thought signature for Gemini 3 streaming")

	// Gemini 3 sends reasoning chunks during streaming
	assert.NotEmpty(t, accumulatedReasoning, "Expected streamed reasoning chunks for Gemini 3")
	t.Logf("Streamed reasoning chunks: %d bytes", len(accumulatedReasoning))
	t.Logf("Streaming - Thought signature: %d bytes, Final reasoning: %d bytes",
		len(choice.Reasoning.Signature), len(choice.Reasoning.Content))

	// Step 2: Test roundtrip - send follow-up message with signature preserved
	// For Gemini 3, we should preserve the signature from streaming response
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
		},
	})
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart("Now divide that result by 23"),
		},
	})

	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithMaxTokens(1024),
	)
	require.NoError(t, err, "Roundtrip should work with preserved streaming signature")
	require.NotEmpty(t, resp2.Choices)
	assert.Contains(t, resp2.Choices[0].Content, "17")
	t.Log("Successfully completed streaming response roundtrip with signature")
}

func TestGoogleAI_SingleFunctionCallWithSignature(t *testing.T) {
	// Use Gemini 3 model which REQUIRES thought signatures for function calling
	llm := newHTTPRRClient(t, WithDefaultModel("gemini-3-flash-preview"))

	// Define a simple calculator tool
	calculatorTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "multiply",
			Description: "Multiply two numbers together",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"a": map[string]any{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]any{
						"type":        "number",
						"description": "Second number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
	}

	// Step 1: Get function call with signature
	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Use the multiply function to calculate 7 times 9"),
			},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools([]llms.Tool{calculatorTool}),
		llms.WithMaxTokens(1024),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	choice := resp.Choices[0]

	// Expect at least one tool call
	require.NotEmpty(t, choice.ToolCalls, "Expected at least one tool call")

	firstToolCall := choice.ToolCalls[0]
	assert.NotEmpty(t, firstToolCall.ID)
	assert.Equal(t, "multiply", firstToolCall.FunctionCall.Name)

	// Parse arguments to verify numbers
	var args map[string]any
	err = json.Unmarshal([]byte(firstToolCall.FunctionCall.Arguments), &args)
	require.NoError(t, err)
	assert.NotNil(t, args["a"])
	assert.NotNil(t, args["b"])

	// For Gemini 3 function calls, signature MUST be in the first ToolCall
	assert.NotNil(t, firstToolCall.Reasoning, "Expected reasoning in first tool call for Gemini 3")
	assert.NotEmpty(t, firstToolCall.Reasoning.Signature, "Expected thought signature in first tool call for Gemini 3")
	t.Logf("First tool call - Signature: %d bytes, Reasoning: %d bytes",
		len(firstToolCall.Reasoning.Signature), len(firstToolCall.Reasoning.Content))

	// ContentChoice.Reasoning should be nil when tool calls are present
	assert.Nil(t, choice.Reasoning, "Expected reasoning to be in ToolCall, not in ContentChoice")

	// Step 2: Test roundtrip - send tool response with signature preserved
	// This is REQUIRED for Gemini 3, otherwise we get 400 error
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			firstToolCall, // This includes the signature via extractThoughtSignature
		},
	})

	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: firstToolCall.ID,
				Name:       firstToolCall.FunctionCall.Name,
				Content:    "63",
			},
		},
	})

	// Step 3: Get final response - this should work if signature was preserved correctly
	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithMaxTokens(512),
	)
	require.NoError(t, err, "Roundtrip should work with preserved signature")
	require.NotEmpty(t, resp2.Choices)
	assert.Contains(t, resp2.Choices[0].Content, "63")
	t.Log("Successfully completed single function call roundtrip with signature")
}

func TestGoogleAI_ParallelFunctionCallsWithSignature(t *testing.T) {
	// Use Gemini 3 model which supports thought signatures
	llm := newHTTPRRClient(t, WithDefaultModel("gemini-3-flash-preview"))

	// Define a weather tool
	weatherTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_temperature",
			Description: "Get the current temperature for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The city name, e.g. Paris",
					},
				},
				"required": []string{"location"},
			},
		},
	}

	// Step 1: Get parallel function calls with signature
	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the temperature in Paris and London?"),
			},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools([]llms.Tool{weatherTool}),
		llms.WithMaxTokens(1024),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	choice := resp.Choices[0]

	// Expect two parallel tool calls
	require.GreaterOrEqual(t, len(choice.ToolCalls), 2, "Expected at least 2 parallel tool calls")
	t.Logf("Got %d parallel tool calls", len(choice.ToolCalls))

	// According to Gemini 3 docs: signature ONLY on the FIRST function call in parallel calls
	assert.NotNil(t, choice.ToolCalls[0].Reasoning, "Expected reasoning in first tool call")
	assert.NotEmpty(t, choice.ToolCalls[0].Reasoning.Signature, "Expected signature ONLY in first tool call")
	t.Logf("First tool call signature: %d bytes", len(choice.ToolCalls[0].Reasoning.Signature))

	// Subsequent parallel tool calls should NOT have signature
	for i := 1; i < len(choice.ToolCalls); i++ {
		assert.Nil(t, choice.ToolCalls[i].Reasoning, "Tool call %d should NOT have reasoning in parallel calls", i)
	}

	// ContentChoice.Reasoning should be nil when tool calls are present
	assert.Nil(t, choice.Reasoning, "Expected reasoning to be in ToolCall, not in ContentChoice")

	// Step 2: Test roundtrip with parallel function calls
	// Add all tool calls to history (signature only in first one)
	var toolCallParts []llms.ContentPart //nolint:prealloc
	for _, tc := range choice.ToolCalls {
		toolCallParts = append(toolCallParts, tc)
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: toolCallParts,
	})

	// Add all tool responses
	var toolResponseParts []llms.ContentPart //nolint:prealloc
	for _, tc := range choice.ToolCalls {
		toolResponseParts = append(toolResponseParts, llms.ToolCallResponse{
			ToolCallID: tc.ID,
			Name:       tc.FunctionCall.Name,
			Content:    "22°C",
		})
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: toolResponseParts,
	})

	// Step 3: Get final response
	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithMaxTokens(512),
	)
	require.NoError(t, err, "Parallel function calls roundtrip should work with signature")
	require.NotEmpty(t, resp2.Choices)
	assert.NotEmpty(t, resp2.Choices[0].Content)
	t.Log("Successfully completed parallel function calls roundtrip with signature")
}

func TestGoogleAI_SequentialFunctionCallsWithSignatures(t *testing.T) { //nolint:funlen
	// Use Gemini 3 model for sequential function calling with signatures
	llm := newHTTPRRClient(t, WithDefaultModel("gemini-3-flash-preview"))

	// Define two tools for sequential calling
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_number",
				Description: "Get a specific number",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"which": map[string]any{
							"type":        "string",
							"description": "Which number to get (first or second)",
						},
					},
					"required": []string{"which"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "add_numbers",
				Description: "Add two numbers together",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"a": map[string]any{
							"type":        "number",
							"description": "First number",
						},
						"b": map[string]any{
							"type":        "number",
							"description": "Second number",
						},
					},
					"required": []string{"a", "b"},
				},
			},
		},
	}

	// Step 1: Initial request
	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("First get the number 10, then get the number 32, then add them together."),
			},
		},
	}

	// Step 2: First function call
	resp1, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithMaxTokens(1024),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.Choices)
	require.NotEmpty(t, resp1.Choices[0].ToolCalls, "Expected first tool call")

	firstCall := resp1.Choices[0].ToolCalls[0]
	assert.NotNil(t, firstCall.Reasoning, "Expected reasoning in first call")
	assert.NotEmpty(t, firstCall.Reasoning.Signature, "Expected signature in first call")
	firstSignature := firstCall.Reasoning.Signature
	t.Logf("Step 1 - Tool: %s, Signature: %d bytes", firstCall.FunctionCall.Name, len(firstSignature))

	// Add first function call and response to history
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{firstCall}, // Includes signature
	})
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: firstCall.ID,
				Name:       firstCall.FunctionCall.Name,
				Content:    "10",
			},
		},
	})

	// Step 3: Second function call
	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithMaxTokens(1024),
	)
	require.NoError(t, err, "Sequential call should work with first signature preserved")
	require.NotEmpty(t, resp2.Choices)

	// The model might call another get_number or might call add_numbers
	// Either way, it should have a signature
	if len(resp2.Choices[0].ToolCalls) > 0 {
		secondCall := resp2.Choices[0].ToolCalls[0]
		assert.NotNil(t, secondCall.Reasoning, "Expected reasoning in second call")
		assert.NotEmpty(t, secondCall.Reasoning.Signature, "Expected signature in second call")
		secondSignature := secondCall.Reasoning.Signature
		t.Logf("Step 2 - Tool: %s, Signature: %d bytes", secondCall.FunctionCall.Name, len(secondSignature))

		// Signatures in sequential calls should be different
		assert.NotEqual(t, firstSignature, secondSignature, "Sequential call signatures should be different")

		t.Log("Successfully completed sequential function calls with unique signatures")
	} else {
		t.Log("Model returned final answer instead of second tool call")
	}
}
