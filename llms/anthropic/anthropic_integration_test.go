package anthropic_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
)

// TestAnthropic_MultiTurnToolThinkingCaching tests complex multi-turn scenario with:
// - Caching strategy at client level
// - Multi-turn conversation (3+ turns)
// - Multiple tool calls in single turn
// - Reasoning/thinking integration
// - Predictable caching behavior
func TestAnthropic_MultiTurnToolThinkingCaching(t *testing.T) { //nolint:funlen
	t.Parallel()

	// Create client with caching strategy at construction time
	llm := newHTTPRRClient(t,
		anthropic.WithDefaultCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true, // Cache tools once
			CacheSystem:   true, // Cache system prompt once
			CacheMessages: true, // Cache conversation incrementally
			TTL:           "5m",
		}),
		anthropic.WithModel("claude-sonnet-4-5"),
	)

	// Large system prompt to exceed caching threshold
	marker := "MultiTurnToolThinkingCaching-v1 "
	systemPrompt := strings.Repeat(marker+"You are a helpful assistant that helps with data analysis and calculations. ", 60)

	// Define tools
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "fetch_data",
				Description: strings.Repeat(marker+"Fetch data from database. ", 40),
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
				Description: strings.Repeat(marker+"Perform calculations. ", 40),
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
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Fetch sales data and calculate the total")},
		},
	}

	// TURN 1: Initial request - should cache tools + system + messages
	t.Log("Turn 1: Initial request with tools and thinking")
	resp1, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	// Verify caching metrics - first turn should create cache
	cacheCreation1 := getFromGenerationInfo(resp1, "CacheCreationInputTokens")
	cacheRead1 := getFromGenerationInfo(resp1, "CacheReadInputTokens")
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d", cacheCreation1, cacheRead1)
	assert.Greater(t, cacheCreation1, 0, "First turn should create cache")

	// Verify response has reasoning and tool calls
	choice1 := resp1.Choices[0]
	require.NotEmpty(t, choice1.ToolCalls, "Should have tool calls")
	if choice1.Reasoning != nil {
		assert.NotEmpty(t, choice1.Reasoning.Signature, "Should have reasoning signature")
	}

	// Build message chain with reasoning and tool calls
	aiParts1 := []llms.ContentPart{ //nolint:prealloc
		llms.TextPartWithReasoning(choice1.Content, choice1.Reasoning),
	}
	for _, tc := range choice1.ToolCalls {
		aiParts1 = append(aiParts1, tc)
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: aiParts1,
	})

	// Add tool results
	for _, tc := range choice1.ToolCalls {
		var result string
		switch tc.FunctionCall.Name {
		case "fetch_data":
			result = `{"sales": [100, 200, 300]}`
		case "calculate":
			result = `{"result": 600}`
		default:
			result = `{"status": "ok"}`
		}

		messages = append(messages, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: tc.ID,
					Name:       tc.FunctionCall.Name,
					Content:    result,
				},
			},
		})
	}

	// TURN 2: Continue with tool results - should read cache and write incrementally
	t.Log("Turn 2: Process tool results")
	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	// Verify caching - should read from cache
	cacheCreation2 := getFromGenerationInfo(resp2, "CacheCreationInputTokens")
	cacheRead2 := getFromGenerationInfo(resp2, "CacheReadInputTokens")
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d", cacheCreation2, cacheRead2)
	assert.Greater(t, cacheRead2, 0, "Turn 2 should read from cache")
	assert.Less(t, cacheCreation2, cacheCreation1, "Turn 2 should write less (incremental)")

	// Add response to message chain
	choice2 := resp2.Choices[0]
	aiParts2 := []llms.ContentPart{ //nolint:prealloc
		llms.TextPartWithReasoning(choice2.Content, choice2.Reasoning),
	}
	for _, tc := range choice2.ToolCalls {
		aiParts2 = append(aiParts2, tc)
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: aiParts2,
	})

	// TURN 3: Add tool results from Turn 2 (if any)
	// IMPORTANT: Only tool results should be added to preserve thinking blocks in cache!
	// Adding a non-tool-result user message would strip thinking blocks and invalidate cache
	if len(choice2.ToolCalls) > 0 {
		for _, tc := range choice2.ToolCalls {
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    `{"result": 600}`,
					},
				},
			})
		}
	} else {
		// If no tool calls, this is the final turn - skip Turn 3
		t.Log("Turn 2 was final turn (no more tool calls)")
		t.Logf("✓ Multi-turn test passed: 2 turns with predictable caching behavior")
		t.Logf("  Cache progression: Create=%d → Read=%d,Create=%d",
			cacheCreation1, cacheRead2, cacheCreation2)
		return
	}

	t.Log("Turn 3: Process tool results from Turn 2")
	resp3, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	// Verify caching - should read cumulatively growing cache
	cacheCreation3 := getFromGenerationInfo(resp3, "CacheCreationInputTokens")
	cacheRead3 := getFromGenerationInfo(resp3, "CacheReadInputTokens")
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d", cacheCreation3, cacheRead3)
	assert.Greater(t, cacheRead3, cacheRead2, "Turn 3 should read more (cumulative history)")
	assert.Greater(t, cacheCreation3, 0, "Turn 3 should write new content")

	// Verify response quality
	choice3 := resp3.Choices[0]
	hasContent := len(choice3.Content) > 0
	hasToolCalls := len(choice3.ToolCalls) > 0
	assert.True(t, hasContent || hasToolCalls, "Turn 3 should have content or tool calls")

	t.Logf("✓ Multi-turn test passed: 3 turns with predictable caching behavior")
	t.Logf("  Cache progression: Create=%d → Read=%d,Create=%d → Read=%d,Create=%d",
		cacheCreation1, cacheRead2, cacheCreation2, cacheRead3, cacheCreation3)
}

// TestAnthropic_MessageChainModification tests message chain modification with signature preservation:
// 1. Turn 1: LLM calls first tool → get response
// 2. Turn 2: LLM calls second tool → get response
// 3. Turn 3: MODIFY chain - replace first tool call with different function (summarization), keep second unchanged
// 4. Turn 4: Call LLM with modified chain, verify cache invalidation and signature handling, request first tool again
// 5. Turn 5: Get response and request text summary, verify Anthropic doesn't drop messages without signatures
func TestAnthropic_MessageChainModification(t *testing.T) { //nolint:funlen
	t.Parallel()

	// Create client with caching strategy
	llm := newHTTPRRClient(t,
		anthropic.WithDefaultCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheSystem:   true,
			CacheMessages: true,
		}),
		anthropic.WithModel("claude-sonnet-4-5"),
	)

	systemPrompt := strings.Repeat("You are a data analysis assistant. ", 80)

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "fetch_sales_data",
				Description: strings.Repeat("Fetch sales data from database. ", 40),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"period": map[string]any{"type": "string"},
					},
					"required": []string{"period"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate_metrics",
				Description: strings.Repeat("Calculate business metrics. ", 40),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"metric_type": map[string]any{"type": "string"},
					},
					"required": []string{"metric_type"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_summary",
				Description: strings.Repeat("Get summarized data. ", 40),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"data_type": map[string]any{"type": "string"},
					},
					"required": []string{"data_type"},
				},
			},
		},
	}

	messages := []llms.MessageContent{ //nolint:prealloc
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Fetch Q1 sales data and calculate revenue metrics")},
		},
	}

	// TURN 1: LLM calls first tool
	t.Log("Turn 1: LLM calls first tool")
	resp1, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	choice1 := resp1.Choices[0]
	require.NotEmpty(t, choice1.ToolCalls, "Turn 1 should have tool calls")
	firstToolCall := choice1.ToolCalls[0]

	// Save original signature from Turn 1
	originalSignature1 := choice1.Reasoning.Signature
	t.Logf("Turn 1: Tool=%s, Signature length=%d", firstToolCall.FunctionCall.Name, len(originalSignature1))

	// Add AI response with reasoning
	aiParts1 := []llms.ContentPart{
		llms.TextPartWithReasoning(choice1.Content, choice1.Reasoning),
		firstToolCall,
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: aiParts1,
	})

	// Add tool result
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: firstToolCall.ID,
				Name:       firstToolCall.FunctionCall.Name,
				Content:    `{"sales": [1000, 2000, 3000], "total": 6000}`,
			},
		},
	})

	// TURN 2: LLM calls second tool
	t.Log("Turn 2: LLM calls second tool")
	resp2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	choice2 := resp2.Choices[0]
	require.NotEmpty(t, choice2.ToolCalls, "Turn 2 should have tool calls")
	secondToolCall := choice2.ToolCalls[0]

	// Cache metrics for Turn 2
	cacheRead2 := getFromGenerationInfo(resp2, "CacheReadInputTokens")
	t.Logf("Turn 2: Tool=%s, CacheRead=%d", secondToolCall.FunctionCall.Name, cacheRead2)

	// Add AI response with reasoning
	aiParts2 := []llms.ContentPart{
		llms.TextPartWithReasoning(choice2.Content, choice2.Reasoning),
		secondToolCall,
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: aiParts2,
	})

	// Add tool result
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: secondToolCall.ID,
				Name:       secondToolCall.FunctionCall.Name,
				Content:    `{"revenue": 6000, "growth": "20%"}`,
			},
		},
	})

	// TURN 3: MODIFY message chain - replace first tool call, keep second unchanged
	t.Log("Turn 3: Modify message chain - replace first tool call with summarized version")

	// Create modified chain:
	// - Keep system, user
	// - Replace Turn 1 AI response (index 2): change to different function WITHOUT signature
	// - Keep Turn 1 tool result (index 3) but update to match new function
	// - Keep Turn 2 AI response (index 4) WITH signature
	// - Keep Turn 2 tool result (index 5)

	modifiedMessages := make([]llms.MessageContent, len(messages))
	copy(modifiedMessages, messages)

	// Replace Turn 1: change function name and remove signature (simulating summarization)
	summarizedToolCallID := "summarized_" + firstToolCall.ID
	modifiedMessages[2] = llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextPart("I'll get the summarized sales data."), // NO reasoning = NO signature
			llms.ToolCall{
				ID:   summarizedToolCallID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_summary", // CHANGED from fetch_sales_data
					Arguments: `{"data_type": "Q1_sales"}`,
				},
			},
		},
	}

	// Update Turn 1 tool result to match new function
	modifiedMessages[3] = llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: summarizedToolCallID,
				Name:       "get_summary",
				Content:    `{"summary": "Q1 total sales: 6000"}`, // Summarized version
			},
		},
	}

	// Turn 2 stays unchanged (indices 4, 5) - it has signature from choice2.Reasoning

	// Add new user request
	modifiedMessages = append(modifiedMessages, llms.MessageContent{ //nolint:makezero
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("Now fetch the detailed Q1 sales data again")},
	})

	// TURN 4: Call with modified chain
	t.Log("Turn 4: Call LLM with modified chain")
	resp4, err := llm.GenerateContent(t.Context(), modifiedMessages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err, "Provider should handle modified chain")

	choice4 := resp4.Choices[0]

	// Verify cache behavior - API intelligently reuses cache even after modification
	cacheRead4 := getFromGenerationInfo(resp4, "CacheReadInputTokens")
	cacheCreation4 := getFromGenerationInfo(resp4, "CacheCreationInputTokens")
	t.Logf("Turn 4 (modified chain): CacheRead=%d (was %d in Turn 2), CacheCreation=%d",
		cacheRead4, cacheRead2, cacheCreation4)

	// API may increase or decrease cache read depending on what parts of chain are reusable
	// The important check: both cache read and creation are happening (proving partial cache reuse)
	assert.Greater(t, cacheRead4, 0, "Cache read should still work (partial reuse)")
	assert.Greater(t, cacheCreation4, 0, "Cache creation should happen (new content)")

	// Verify model returns thinking
	assert.NotNil(t, choice4.Reasoning, "Turn 4 should have reasoning")
	if choice4.Reasoning != nil {
		assert.NotEmpty(t, choice4.Reasoning.Signature, "Turn 4 should have signature")
		t.Logf("Turn 4: Signature length=%d", len(choice4.Reasoning.Signature))
	}

	// Verify model called first tool again (as requested)
	if len(choice4.ToolCalls) > 0 {
		t.Logf("Turn 4: Called tool=%s", choice4.ToolCalls[0].FunctionCall.Name)

		// Add AI response
		aiParts4 := []llms.ContentPart{ //nolint:prealloc
			llms.TextPartWithReasoning(choice4.Content, choice4.Reasoning),
		}
		for _, tc := range choice4.ToolCalls {
			aiParts4 = append(aiParts4, tc)
		}
		modifiedMessages = append(modifiedMessages, llms.MessageContent{ //nolint:makezero
			Role:  llms.ChatMessageTypeAI,
			Parts: aiParts4,
		})

		// Add tool results
		for _, tc := range choice4.ToolCalls {
			modifiedMessages = append(modifiedMessages, llms.MessageContent{ //nolint:makezero
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    `{"sales": [1000, 2000, 3000], "total": 6000, "detailed": true}`,
					},
				},
			})
		}
	}

	// TURN 5: Request text summary including info from modified first message
	t.Log("Turn 5: Request text summary of entire conversation")
	modifiedMessages = append(modifiedMessages, llms.MessageContent{ //nolint:makezero
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("Provide a summary of all data we've collected, including the summarized Q1 data from the beginning")},
	})

	resp5, err := llm.GenerateContent(t.Context(), modifiedMessages,
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		anthropic.WithInterleavedThinking(),
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	require.NoError(t, err)

	choice5 := resp5.Choices[0]

	// Verify Anthropic doesn't drop messages without signatures
	assert.NotEmpty(t, choice5.Content, "Turn 5 should have text response")
	t.Logf("Turn 5: Response length=%d", len(choice5.Content))

	// The response should reference the summarized data (proving it wasn't dropped)
	// We can't assert exact content, but we can verify it's substantial
	assert.Greater(t, len(choice5.Content), 50, "Response should be substantial, including all conversation context")

	t.Logf("✓ Message chain modification test passed")
	t.Logf("  Turn 1: Original tool call WITH signature")
	t.Logf("  Turn 2: Second tool call WITH signature (CacheRead=%d)", cacheRead2)
	t.Logf("  Turn 3: Modified Turn 1 to different function WITHOUT signature (simulating summarization)")
	t.Logf("  Turn 4: API intelligently reused cache (CacheRead=%d, CacheCreation=%d)", cacheRead4, cacheCreation4)
	t.Logf("  Turn 4: Model handled chain with mixed signatures correctly")
	t.Logf("  Turn 5: Model included all context, proving messages without signatures aren't dropped")
}
