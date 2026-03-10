package openai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// OPENAI IMPLICIT CACHING TESTS
// ============================================================================
//
// This file contains tests for OpenAI's automatic prompt caching mechanism.
//
// KEY FINDINGS (as of January 2026):
//
// 1. IMPLICIT CACHING (automatic):
//    ✓ Automatically enabled for all OpenAI models (gpt-4.1-mini, o3-mini, etc.)
//    ✓ Works for completely IDENTICAL requests (byte-for-byte HTTP body match)
//    ✓ Works for conversation continuation with proper message history
//    ✓ Requires minimum 1,024 tokens for caching
//    ✓ Cache TTL is typically 5-10 minutes
//    ✓ No configuration required - fully automatic
//    ✓ Cached tokens reported in usage.prompt_tokens_details.cached_tokens
//    ✗ cache_write_tokens field is NOT returned by OpenAI API (always 0 in our mapping)
//
// 2. IMPORTANT NOTES:
//    - OpenAI does NOT support explicit caching (no WithCachedContent option)
//    - OpenAI does NOT return cache_write_tokens in API responses
//    - All caching is implicit and automatic
//    - For reasoning models (o1, o3), reasoning content must be preserved in conversation
//    - Cache hits significantly reduce latency and cost
//
// 3. TESTING STRATEGY:
//    - Test identical requests (non-streaming and streaming)
//    - Test conversation continuation (non-streaming and streaming)
//    - Test both non-reasoning models (gpt-4.1-mini) and reasoning models (o3-mini)
//    - Verify cached tokens are reported correctly in subsequent requests
//
// ============================================================================

// TestOpenAI_ImplicitCaching_IdenticalRequests_NonReasoning tests that implicit caching
// works correctly for identical requests with a non-reasoning model
func TestOpenAI_ImplicitCaching_IdenticalRequests_NonReasoning(t *testing.T) {
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v3", "v4" etc. if re-recording
	marker := "IdenticalNonReasoning-v2 "
	largeContext := strings.Repeat(marker+"Go is a statically typed, compiled programming language. ", 200)

	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("Say 'hello'")}},
	}

	// Request 1: Initial request - establishes cache
	r1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithMaxTokens(50),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r1.Choices)

	c1 := 0
	if ct, ok := r1.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}

	// Request 1: First request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens (first request)")

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: IDENTICAL to request 1 - should hit cache
	r2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithMaxTokens(50),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r2.Choices)

	c2 := 0
	if ct, ok := r2.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 MUST have cached tokens for identical request
	assert.Greater(t, c2, 0, "Request 2 (identical) must have cached tokens")
	// Note: OpenAI does not return cache_write_tokens, so CacheCreationInputTokens is always 0

	// Verify content is not empty
	assert.NotEmpty(t, r1.Choices[0].Content, "Response 1 content should not be empty")
	assert.NotEmpty(t, r2.Choices[0].Content, "Response 2 content should not be empty")
}

// TestOpenAI_ImplicitCaching_IdenticalRequests_Reasoning tests that implicit caching
// works correctly for identical requests with a reasoning model
func TestOpenAI_ImplicitCaching_IdenticalRequests_Reasoning(t *testing.T) {
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v4", "v5" etc. if re-recording
	marker := "IdenticalReasoning-v3 "
	largeContext := strings.Repeat(marker+"Python is a high-level, interpreted programming language. ", 200)

	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What is 7 * 8? Think about it step by step.")}},
	}

	// Request 1: Initial request with reasoning - establishes cache
	r1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithMaxTokens(4096),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r1.Choices)

	choice1 := r1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	rr1 := 0
	if r, ok := choice1.GenerationInfo["ReasoningTokens"].(int); ok {
		rr1 = r
	}

	// Request 1: First request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens (first request)")
	assert.Contains(t, choice1.Content, "56", "Response should contain the answer 56")
	// TODO: return reasoning is not supported yet for OpenAI /chat/completions endpoint
	// assert.NotNil(t, choice1.Reasoning, "Reasoning model should return reasoning")
	assert.Greater(t, rr1, 0, "Request 1 (reasoning) should return reasoning")

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: IDENTICAL to request 1 - should hit cache
	r2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithMaxTokens(4096),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r2.Choices)

	choice2 := r2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}
	rr2 := 0
	if r, ok := choice2.GenerationInfo["ReasoningTokens"].(int); ok {
		rr2 = r
	}

	// Assert: Request 2 MUST have cached tokens for identical request
	assert.Greater(t, c2, 0, "Request 2 (identical) must have cached tokens")
	// Note: OpenAI does not return cache_write_tokens, so CacheCreationInputTokens is always 0
	assert.Contains(t, choice2.Content, "56", "Response should contain the answer 56")
	// TODO: return reasoning is not supported yet for OpenAI /chat/completions endpoint
	// assert.NotNil(t, choice2.Reasoning, "Reasoning model should return reasoning")
	assert.Greater(t, rr2, 0, "Request 2 (reasoning) should return reasoning")
}

// TestOpenAI_ImplicitCaching_Streaming_NonReasoning tests implicit caching
// with streaming for a non-reasoning model
func TestOpenAI_ImplicitCaching_Streaming_NonReasoning(t *testing.T) {
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "StreamNonReasoning-v1 "
	largeContext := strings.Repeat(marker+"Rust is a systems programming language focused on safety. ", 200)

	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("Say 'world'")}},
	}

	streamFunc := func(content *strings.Builder) llms.CallOption {
		return llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeText {
				content.WriteString(chunk.Content)
			}
			return nil
		})
	}

	// Request 1: Initial streaming request - establishes cache
	var s1 strings.Builder
	r1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithMaxTokens(50),
		streamFunc(&s1),
	)
	require.NoError(t, err)
	require.NotNil(t, r1.Choices[0].GenerationInfo, "GenerationInfo should be present in streaming")

	c1 := 0
	if ct, ok := r1.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}

	// Request 1: First request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 (streaming) should have 0 cached tokens")
	assert.NotEmpty(t, s1.String(), "Streamed content should not be empty")

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: IDENTICAL streaming request - should hit cache
	var s2 strings.Builder
	r2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithMaxTokens(50),
		streamFunc(&s2),
	)
	require.NoError(t, err)
	require.NotNil(t, r2.Choices[0].GenerationInfo, "GenerationInfo should be present in streaming")

	c2 := 0
	if ct, ok := r2.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 MUST have cached tokens for identical streaming request
	assert.Greater(t, c2, 0, "Request 2 (identical streaming) must have cached tokens")
	// Note: OpenAI does not return cache_write_tokens, so CacheCreationInputTokens is always 0
	assert.NotEmpty(t, s2.String(), "Streamed content should not be empty")
}

// TestOpenAI_ImplicitCaching_Streaming_Reasoning tests implicit caching
// with streaming for a reasoning model
func TestOpenAI_ImplicitCaching_Streaming_Reasoning(t *testing.T) {
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v4", "v5" etc. if re-recording
	marker := "StreamReasoning-v3 "
	largeContext := strings.Repeat(marker+"JavaScript is a dynamic programming language. ", 200)

	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What is 12 + 13?")}},
	}

	streamFunc := func(content *strings.Builder, reasoning *strings.Builder) llms.CallOption {
		return llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			switch chunk.Type {
			case streaming.ChunkTypeText:
				content.WriteString(chunk.Content)
			case streaming.ChunkTypeReasoning:
				if chunk.Reasoning != nil {
					reasoning.WriteString(chunk.Reasoning.Content)
				}
			default:
			}
			return nil
		})
	}

	// Request 1: Initial streaming request with reasoning - establishes cache
	var s1, r1Content strings.Builder
	r1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithMaxTokens(4096),
		streamFunc(&s1, &r1Content),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r1.Choices)

	choice1 := r1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	rr1 := 0
	if r, ok := choice1.GenerationInfo["ReasoningTokens"].(int); ok {
		rr1 = r
	}

	// Request 1: First request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 (streaming reasoning) should have 0 cached tokens")
	assert.Contains(t, choice1.Content, "25", "Response should contain the answer 25")
	// TODO: return reasoning is not supported yet for OpenAI /chat/completions endpoint
	// assert.NotNil(t, choice1.Reasoning, "Reasoning model streaming should return reasoning")
	assert.Greater(t, rr1, 0, "Request 1 (streaming reasoning) should return reasoning")
	assert.NotEmpty(t, s1.String(), "Streamed content should not be empty")

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: IDENTICAL streaming request - should hit cache
	var s2, r2Content strings.Builder
	r2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithMaxTokens(4096),
		streamFunc(&s2, &r2Content),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r2.Choices)

	choice2 := r2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}
	rr2 := 0
	if r, ok := choice2.GenerationInfo["ReasoningTokens"].(int); ok {
		rr2 = r
	}
	// Assert: Request 2 MUST have cached tokens for identical streaming request
	assert.Greater(t, c2, 0, "Request 2 (identical streaming reasoning) must have cached tokens")
	// Note: OpenAI does not return cache_write_tokens, so CacheCreationInputTokens is always 0
	assert.Contains(t, choice2.Content, "25", "Response should contain the answer 25")
	// TODO: return reasoning is not supported yet for OpenAI /chat/completions endpoint
	// assert.NotNil(t, choice2.Reasoning, "Reasoning model streaming should return reasoning")
	assert.Greater(t, rr2, 0, "Request 2 (streaming reasoning) should return reasoning")
	assert.NotEmpty(t, s2.String(), "Streamed content should not be empty")
}

// TestOpenAI_ImplicitCaching_Conversation_NonReasoning tests implicit caching
// for multi-turn conversations with tool calls using a non-reasoning model
func TestOpenAI_ImplicitCaching_Conversation_NonReasoning(t *testing.T) { //nolint:funlen
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvToolsNonReason-v1 "
	largeContext := marker + strings.Repeat("You are a helpful assistant with access to weather and calculation tools. ", 200)

	// Define tools for the conversation
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_temperature",
				Description: "Get the current temperature for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city name, e.g. Boston",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Calculate a mathematical expression",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{
							"type":        "string",
							"description": "Mathematical expression to calculate",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	// Request 1: User asks for temperature - LLM should call get_temperature tool
	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What's the temperature in Boston?")}},
	}

	r1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithTools(tools),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r1.Choices)

	choice1 := r1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}

	// Request 1: First request, no cache expected, should have tool call
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens (first request)")
	assert.NotEmpty(t, choice1.ToolCalls, "Request 1 should have tool calls")
	require.Len(t, choice1.ToolCalls, 1, "Request 1 should have exactly 1 tool call")
	assert.Equal(t, "get_temperature", choice1.ToolCalls[0].FunctionCall.Name)

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: Add tool call and response, then ask to calculate with that temperature
	msgs = append(msgs,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				choice1.ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: choice1.ToolCalls[0].ID,
					Name:       choice1.ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "unit": "fahrenheit"}`,
				},
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Convert that temperature to Celsius using the formula (F-32)*5/9")},
		},
	)

	r2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithTools(tools),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r2.Choices)

	choice2 := r2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 may have cached tokens (tool calls may affect caching)
	// Note: OpenAI may not cache prefixes with tool calls as efficiently
	t.Logf("Request 2 cached tokens: %d (may be 0 with tool calls)", c2)
	// Should call calculate tool or respond with answer
	if len(choice2.ToolCalls) > 0 {
		assert.Equal(t, "calculate", choice2.ToolCalls[0].FunctionCall.Name)
	} else {
		assert.Contains(t, choice2.Content, "22", "Response should contain Celsius value ~22")
	}

	// Wait for cache update (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for cache update...")
		time.Sleep(40 * time.Second)
	}

	// Request 3: Continue conversation with final question
	if len(choice2.ToolCalls) > 0 {
		// If tool call was made, add it and response
		msgs = append(msgs,
			llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					choice2.ToolCalls[0],
				},
			},
			llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: choice2.ToolCalls[0].ID,
						Name:       choice2.ToolCalls[0].FunctionCall.Name,
						Content:    `{"result": 22.22}`,
					},
				},
			},
		)
	} else {
		// Otherwise just add AI response
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPart(choice2.Content)},
			},
		)
	}

	msgs = append(msgs, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("Is that temperature comfortable for outdoor activities?")},
	})

	r3, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithTools(tools),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r3.Choices)

	choice3 := r3.Choices[0]
	c3 := 0
	if ct, ok := choice3.GenerationInfo["PromptCachedTokens"].(int); ok {
		c3 = ct
	}

	// Assert: Request 3 may have cached tokens (tool calls may affect caching)
	t.Logf("Request 3 cached tokens: %d (may be 0 with tool calls)", c3)
	assert.NotEmpty(t, choice3.Content, "Response 3 should have content")
	// Verify conversation flow works correctly with tools
	assert.NotNil(t, r3.Choices, "Response 3 should have choices")
}

// TestOpenAI_ImplicitCaching_Conversation_Reasoning tests implicit caching
// for multi-turn conversations with tool calls using a reasoning model.
// IMPORTANT: reasoning content must be preserved when adding AI responses to history
func TestOpenAI_ImplicitCaching_Conversation_Reasoning(t *testing.T) { //nolint:funlen
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvToolsReason-v1 "
	largeContext := marker + strings.Repeat("You are a helpful assistant with access to database query and analysis tools. ", 200)

	// Define tools for the conversation
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "query_database",
				Description: "Query the sales database for information",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "SQL query to execute",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "analyze_trend",
				Description: "Analyze trend in the data",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"data": map[string]any{
							"type":        "string",
							"description": "Data to analyze",
						},
						"period": map[string]any{
							"type":        "string",
							"description": "Time period for analysis",
						},
					},
					"required": []string{"data", "period"},
				},
			},
		},
	}

	// Request 1: User asks for sales data - LLM should call query_database tool
	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("Show me total sales for Q4 2024")}},
	}

	r1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithTools(tools),
		llms.WithMaxTokens(4096),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r1.Choices)

	choice1 := r1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	rr1 := 0
	if r, ok := choice1.GenerationInfo["ReasoningTokens"].(int); ok {
		rr1 = r
	}

	// Request 1: First request, no cache expected, should have tool call
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens (first request)")
	assert.Greater(t, rr1, 0, "Request 1 (reasoning with tools) should return reasoning")
	assert.NotEmpty(t, choice1.ToolCalls, "Request 1 should have tool calls")
	require.Len(t, choice1.ToolCalls, 1, "Request 1 should have exactly 1 tool call")
	assert.Equal(t, "query_database", choice1.ToolCalls[0].FunctionCall.Name)

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: Add tool call with reasoning and response, then ask for trend analysis
	msgs = append(msgs,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				choice1.ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: choice1.ToolCalls[0].ID,
					Name:       choice1.ToolCalls[0].FunctionCall.Name,
					Content:    `{"total_sales": 1250000, "currency": "USD", "quarter": "Q4 2024"}`,
				},
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Analyze the trend compared to Q3 2024 which had 980000 USD")},
		},
	)

	r2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithTools(tools),
		llms.WithMaxTokens(4096),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r2.Choices)

	choice2 := r2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}
	rr2 := 0
	if r, ok := choice2.GenerationInfo["ReasoningTokens"].(int); ok {
		rr2 = r
	}

	// Assert: Request 2 may have cached tokens (tool calls may affect caching)
	// Note: OpenAI may not cache prefixes with tool calls as efficiently
	t.Logf("Request 2 cached tokens: %d (may be 0 with tool calls)", c2)
	assert.Greater(t, rr2, 0, "Request 2 (reasoning with tools) should return reasoning")
	// Should call analyze_trend tool or respond with analysis
	if len(choice2.ToolCalls) > 0 {
		assert.Equal(t, "analyze_trend", choice2.ToolCalls[0].FunctionCall.Name)
	} else {
		assert.NotEmpty(t, choice2.Content, "Response 2 should have analysis content")
	}

	// Wait for cache update (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for cache update...")
		time.Sleep(40 * time.Second)
	}

	// Request 3: Continue conversation with final question
	if len(choice2.ToolCalls) > 0 {
		// If tool call was made, add it with reasoning and response
		msgs = append(msgs,
			llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					choice2.ToolCalls[0],
				},
			},
			llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: choice2.ToolCalls[0].ID,
						Name:       choice2.ToolCalls[0].FunctionCall.Name,
						Content:    `{"trend": "upward", "growth_percent": 27.55, "analysis": "Strong growth quarter over quarter"}`,
					},
				},
			},
		)
	} else {
		// Otherwise just add AI response
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPart(choice2.Content)},
			},
		)
	}

	msgs = append(msgs, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("What should be our target for Q1 2025?")},
	})

	r3, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithTools(tools),
		llms.WithMaxTokens(4096),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r3.Choices)

	choice3 := r3.Choices[0]
	c3 := 0
	if ct, ok := choice3.GenerationInfo["PromptCachedTokens"].(int); ok {
		c3 = ct
	}
	rr3 := 0
	if r, ok := choice3.GenerationInfo["ReasoningTokens"].(int); ok {
		rr3 = r
	}

	// Assert: Request 3 may have cached tokens (tool calls may affect caching)
	t.Logf("Request 3 cached tokens: %d (may be 0 with tool calls)", c3)
	assert.Greater(t, rr3, 0, "Request 3 (reasoning with tools) should return reasoning")
	assert.NotEmpty(t, choice3.Content, "Response 3 should have content")
	// Verify conversation flow works correctly with reasoning and tools
	assert.NotNil(t, r3.Choices, "Response 3 should have choices")
}

// TestOpenAI_ImplicitCaching_Conversation_Streaming_NonReasoning tests implicit caching
// for multi-turn conversations in streaming mode with a non-reasoning model
func TestOpenAI_ImplicitCaching_Conversation_Streaming_NonReasoning(t *testing.T) {
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvStreamNonReasoning-v1 "
	largeContext := marker + strings.Repeat("Java is a class-based object-oriented language. ", 200)

	streamFunc := func(content *strings.Builder) llms.CallOption {
		return llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeText {
				content.WriteString(chunk.Content)
			}
			return nil
		})
	}

	// Request 1: Initial conversation turn [system, user1]
	msgs := []llms.MessageContent{ //nolint:prealloc
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What is 4 + 5?")}},
	}

	var s1 strings.Builder
	resp1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithMaxTokens(100),
		streamFunc(&s1),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.Choices)

	choice1 := resp1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}

	// Request 1: First request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 (streaming) should have 0 cached tokens")
	assert.Contains(t, choice1.Content, "9", "Response should contain the answer 9")
	assert.NotEmpty(t, s1.String(), "Streamed content should not be empty")

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: Continue conversation [system, user1, AI1, user2]
	msgs = append(msgs,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(choice1.Content)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now multiply that by 4")},
		},
	)

	var s2 strings.Builder
	resp2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("gpt-4.1-mini"),
		llms.WithMaxTokens(100),
		streamFunc(&s2),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp2.Choices)

	choice2 := resp2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 MUST have cached tokens from prefix [system, user1] in streaming mode
	assert.Greater(t, c2, 0, "Request 2 (streaming) must have cached tokens from prefix")
	assert.Contains(t, choice2.Content, "36", "Response should contain the answer 36")
	assert.NotEmpty(t, s2.String(), "Streamed content should not be empty")
}

// TestOpenAI_ImplicitCaching_Conversation_Streaming_Reasoning tests implicit caching
// for multi-turn conversations in streaming mode with a reasoning model
func TestOpenAI_ImplicitCaching_Conversation_Streaming_Reasoning(t *testing.T) { //nolint:funlen
	t.Parallel()

	llm := newTestOpenAIClient(t)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvStreamReasoning-v1 "
	largeContext := marker + strings.Repeat("Swift is a powerful programming language for iOS. ", 200)

	streamFunc := func(content *strings.Builder, reasoning *strings.Builder) llms.CallOption {
		return llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			switch chunk.Type {
			case streaming.ChunkTypeText:
				content.WriteString(chunk.Content)
			case streaming.ChunkTypeReasoning:
				if chunk.Reasoning != nil {
					reasoning.WriteString(chunk.Reasoning.Content)
				}
			default:
			}
			return nil
		})
	}

	// Request 1: Initial conversation turn [system, user1]
	msgs := []llms.MessageContent{ //nolint:prealloc
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What is 6 + 7?")}},
	}

	var s1, r1Content strings.Builder
	resp1, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithMaxTokens(4096),
		streamFunc(&s1, &r1Content),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.Choices)

	choice1 := resp1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	r1Reasoning := 0
	if r, ok := choice1.GenerationInfo["ReasoningTokens"].(int); ok {
		r1Reasoning = r
	}

	// Request 1: First request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 (streaming) should have 0 cached tokens")
	assert.Contains(t, choice1.Content, "13", "Response should contain the answer 13")
	// TODO: return reasoning is not supported yet for OpenAI /chat/completions endpoint
	// assert.NotNil(t, choice1.Reasoning, "Reasoning model streaming should return reasoning")
	assert.Greater(t, r1Reasoning, 0, "Request 1 (conversation streaming reasoning) should return reasoning")
	assert.NotEmpty(t, s1.String(), "Streamed content should not be empty")

	// Wait for cache to be established (only in recording mode)
	if os.Getenv("HTTPRR_RECORD") != "" {
		t.Log("Waiting 40 seconds for OpenAI to establish cache...")
		time.Sleep(40 * time.Second)
	}

	// Request 2: Continue conversation [system, user1, AI1+reasoning, user2]
	msgs = append(msgs,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice1.Content, choice1.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now multiply that by 2")},
		},
	)

	var s2, r2Content strings.Builder
	resp2, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithModel("o3-mini"),
		llms.WithReasoning(llms.ReasoningMedium, 1000),
		llms.WithMaxTokens(4096),
		streamFunc(&s2, &r2Content),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp2.Choices)

	choice2 := resp2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}
	r2Reasoning := 0
	if r, ok := choice2.GenerationInfo["ReasoningTokens"].(int); ok {
		r2Reasoning = r
	}

	// Assert: Request 2 MUST have cached tokens from prefix [system, user1] in streaming mode
	assert.Greater(t, c2, 0, "Request 2 (streaming) must have cached tokens from prefix")
	assert.Contains(t, choice2.Content, "26", "Response should contain the answer 26")
	// TODO: return reasoning is not supported yet for OpenAI /chat/completions endpoint
	// assert.NotNil(t, choice2.Reasoning, "Reasoning model streaming should return reasoning")
	assert.Greater(t, r2Reasoning, 0, "Request 2 (conversation streaming reasoning) should return reasoning")
	assert.NotEmpty(t, s2.String(), "Streamed content should not be empty")
}
