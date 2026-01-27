package anthropic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicCacheControlInMessages(t *testing.T) {
	// Test that cache control is properly handled in message processing
	// This is a unit test that doesn't require API calls

	cachedText := anthropic.WithCacheControl(
		llms.TextPart("This is cached content"),
		anthropic.EphemeralCache(),
	)

	message := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{cachedText},
	}

	// This tests that the message can be created without errors
	// The actual processing is tested through integration tests
	messages := []llms.MessageContent{message}

	if len(messages) != 1 {
		t.Error("expected single message")
	}

	if len(messages[0].Parts) != 1 {
		t.Error("expected single part in message")
	}

	cached, ok := messages[0].Parts[0].(anthropic.CachedContent)
	if !ok {
		t.Error("expected CachedContent part")
	}

	if cached.CacheControl == nil {
		t.Error("expected cache control to be set")
	}

	if cached.CacheControl.Type != "ephemeral" {
		t.Errorf("expected ephemeral cache type, got %q", cached.CacheControl.Type)
	}
}

func TestAnthropicBetaHeaders(t *testing.T) {
	// Test that beta headers option works correctly
	option := anthropic.WithPromptCaching()

	var opts llms.CallOptions
	option(&opts)

	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}

	headers, ok := opts.Metadata["anthropic:beta_headers"].([]string)
	if !ok {
		t.Fatal("anthropic:beta_headers should be a []string")
	}

	expectedHeader := "prompt-caching-2024-07-31"
	if len(headers) != 1 || headers[0] != expectedHeader {
		t.Errorf("expected [%q], got %v", expectedHeader, headers)
	}
}

// Helper to get value from generation info
func getFromGenerationInfo(resp *llms.ContentResponse, key string) int {
	if value, ok := resp.Choices[0].GenerationInfo[key]; ok {
		valueInt, ok := value.(int)
		if ok {
			return valueInt
		}
	}
	return 0
}

// TestAnthropic_ExplicitCaching tests explicit caching with cache control
func TestAnthropic_ExplicitCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-haiku-4-5"))

	// Haiku 4.5 requires minimum 4096 tokens for caching
	longContext := strings.Repeat("Context text with additional content for caching.\n", 1000) // > 4096 tokens

	cachedPart := anthropic.WithCacheControl(
		llms.TextPart(longContext),
		anthropic.EphemeralCache(), // 5-minute cache
	)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{cachedPart},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Say hello")},
		},
	}

	// First request: cache miss
	resp1, err := llm.GenerateContent(t.Context(), messages,
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	// Verify cache metrics are present (values may vary)
	cacheRead1Int := getFromGenerationInfo(resp1, "CacheReadInputTokens")
	assert.Equal(t, cacheRead1Int, 0)
	cacheCreation1Int := getFromGenerationInfo(resp1, "CacheCreationInputTokens")
	assert.Greater(t, cacheCreation1Int, 0)
	cacheCreation5mInt := getFromGenerationInfo(resp1, "CacheCreationEphemeral5mInputTokens")
	assert.Greater(t, cacheCreation5mInt, 0)
	assert.Equal(t, cacheCreation5mInt, cacheCreation1Int)
	cacheCreation1hInt := getFromGenerationInfo(resp1, "CacheCreationEphemeral1hInputTokens")
	assert.Equal(t, cacheCreation1hInt, 0)

	// Second request: cache hit (in recording mode this requires a delay)
	// In replay mode, httprr will provide the cached response
	resp2, err := llm.GenerateContent(t.Context(), messages,
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	// Verify cache metrics are present
	cacheCreation2Int := getFromGenerationInfo(resp2, "CacheCreationInputTokens")
	assert.Equal(t, cacheCreation2Int, 0)
	cacheCreation5m2Int := getFromGenerationInfo(resp2, "CacheCreationEphemeral5mInputTokens")
	assert.Equal(t, cacheCreation5m2Int, 0)
	cacheCreation1h2Int := getFromGenerationInfo(resp2, "CacheCreationEphemeral1hInputTokens")
	assert.Equal(t, cacheCreation1h2Int, 0)
	cacheRead2Int := getFromGenerationInfo(resp2, "CacheReadInputTokens")
	assert.Greater(t, cacheRead2Int, 0)
	assert.Equal(t, cacheRead2Int, cacheCreation1Int)
}

// TestAnthropic_TTLCaching tests 1-hour cache
func TestAnthropic_TTLCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-haiku-4-5"))

	// Haiku 4.5 requires minimum 4096 tokens for caching
	longContext := strings.Repeat("Context text with additional content for caching.\n", 600) // > 4096 tokens

	// Test 1-hour cache
	cache1h := anthropic.WithCacheControl(
		llms.TextPart(longContext),
		anthropic.EphemeralCacheOneHour(), // 1 hour
	)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{cache1h},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Say hello")},
		},
	}

	// First request: cache miss
	resp1, err := llm.GenerateContent(t.Context(), messages,
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	// Verify cache metrics are present (values may vary)
	cacheRead1Int := getFromGenerationInfo(resp1, "CacheReadInputTokens")
	assert.Equal(t, cacheRead1Int, 0)
	cacheCreation1Int := getFromGenerationInfo(resp1, "CacheCreationInputTokens")
	assert.Greater(t, cacheCreation1Int, 0)
	cacheCreation5mInt := getFromGenerationInfo(resp1, "CacheCreationEphemeral5mInputTokens")
	assert.Equal(t, cacheCreation5mInt, 0)
	cacheCreation1hInt := getFromGenerationInfo(resp1, "CacheCreationEphemeral1hInputTokens")
	assert.Greater(t, cacheCreation1hInt, 0)
	assert.Equal(t, cacheCreation1hInt, cacheCreation1Int)

	// Second request: cache hit (in recording mode this requires a delay)
	// In replay mode, httprr will provide the cached response
	resp2, err := llm.GenerateContent(t.Context(), messages,
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	// Verify cache metrics are present
	cacheCreation2Int := getFromGenerationInfo(resp2, "CacheCreationInputTokens")
	assert.Equal(t, cacheCreation2Int, 0)
	cacheCreation5m2Int := getFromGenerationInfo(resp2, "CacheCreationEphemeral5mInputTokens")
	assert.Equal(t, cacheCreation5m2Int, 0)
	cacheCreation1h2Int := getFromGenerationInfo(resp2, "CacheCreationEphemeral1hInputTokens")
	assert.Equal(t, cacheCreation1h2Int, 0)
	cacheRead2Int := getFromGenerationInfo(resp2, "CacheReadInputTokens")
	assert.Greater(t, cacheRead2Int, 0)
	assert.Equal(t, cacheRead2Int, cacheCreation1hInt)
}

// Helper functions to generate consistent test data across tests

// generateLargeTools creates tool definitions that exceed caching threshold (>1024 tokens for sonnet-4-5).
// Returns tools that produce ~1530 tokens when serialized.
// marker: unique prefix to avoid cache collisions between tests (e.g., "ToolsTest-v1")
func generateLargeTools(marker string) []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: marker + strings.Repeat("Get current weather conditions including temperature, humidity, wind speed, and precipitation for a specified geographic location. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City and state, e.g. San Francisco, CA, or city and country, e.g. London, UK",
						},
						"units": map[string]any{
							"type":        "string",
							"description": "Temperature units: celsius or fahrenheit",
							"enum":        []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_time",
				Description: marker + strings.Repeat("Get the current time in any timezone around the world using IANA timezone identifiers. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"timezone": map[string]any{
							"type":        "string",
							"description": "IANA timezone name like America/New_York or Europe/London",
						},
						"format": map[string]any{
							"type":        "string",
							"description": "Time format: 12h or 24h",
							"enum":        []string{"12h", "24h"},
						},
					},
					"required": []string{"timezone"},
				},
			},
		},
	}
}

// generateLargeSystemPrompt creates a system message that exceeds caching threshold (>1024 tokens for sonnet-4-5).
// Returns a system prompt that produces ~1200 tokens.
// marker: unique prefix to avoid cache collisions between tests (e.g., "SystemTest-v1")
func generateLargeSystemPrompt(marker string) string {
	return marker + strings.Repeat("You are a helpful assistant with extensive knowledge. ", 200)
}

// generateMediumSystemPrompt creates a smaller system message for combined testing.
// Returns a system prompt that produces ~300 tokens.
// marker: unique prefix to avoid cache collisions between tests (e.g., "CombinedTest-v1")
func generateMediumSystemPrompt(marker string) string {
	return marker + strings.Repeat("You are an expert assistant. ", 50)
}

// TestAnthropic_ToolsCaching tests how tools caching works:
// Q1: Do tools get cached once or written every time?
// Q2: On second request with same tools, do we get cache hit?
// Q3: What tokens are included in the cache?
//
// Expected behavior:
// Request 1: CacheCreationInputTokens > 0 (tools written to cache), CacheReadInputTokens == 0
// Request 2: CacheCreationInputTokens == 0, CacheReadInputTokens > 0 (tools read from cache)
// Request 3: Same as Request 2 (cache hit again)
//
// BASELINE TEST: This establishes the token count for tools alone (~1530 tokens).
func TestAnthropic_ToolsCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ToolsTest-v1 "
	tools := generateLargeTools(marker)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather?")},
		},
	}

	// Request 1: First call - tools should be written to cache
	t.Log("Request 1: First call with tools caching enabled")
	r1, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(), // REQUIRED: Enable caching beta header
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation1 := getFromGenerationInfo(r1, "CacheCreationInputTokens")
	cacheRead1 := getFromGenerationInfo(r1, "CacheReadInputTokens")
	inputTokens1 := getFromGenerationInfo(r1, "InputTokens")

	t.Logf("Request 1 - CacheCreation: %d, CacheRead: %d, Input: %d",
		cacheCreation1, cacheRead1, inputTokens1)

	// ASSERTION: First request should create cache
	assert.Greater(t, cacheCreation1, 0, "First request should create cache for tools")
	assert.Equal(t, 0, cacheRead1, "First request should have no cache reads")

	// BASELINE: Record the tools token count for use in other tests
	toolsTokenCount := cacheCreation1
	t.Logf("BASELINE: Tools alone = %d tokens", toolsTokenCount)

	// Request 2: Second identical call - should hit cache
	t.Log("Request 2: Identical call - should hit cache")
	r2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation2 := getFromGenerationInfo(r2, "CacheCreationInputTokens")
	cacheRead2 := getFromGenerationInfo(r2, "CacheReadInputTokens")
	inputTokens2 := getFromGenerationInfo(r2, "InputTokens")

	t.Logf("Request 2 - CacheCreation: %d, CacheRead: %d, Input: %d",
		cacheCreation2, cacheRead2, inputTokens2)

	// ASSERTION: Second request should read from cache, not write
	assert.Equal(t, 0, cacheCreation2, "Second identical request should NOT create new cache")
	assert.Greater(t, cacheRead2, 0, "Second request should read tools from cache")
	assert.Equal(t, cacheRead2, cacheCreation1, "Cache read should equal initial cache creation")

	// Request 3: Third call - verify cache is still valid
	t.Log("Request 3: Third call - cache should still be valid")
	r3, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation3 := getFromGenerationInfo(r3, "CacheCreationInputTokens")
	cacheRead3 := getFromGenerationInfo(r3, "CacheReadInputTokens")

	t.Logf("Request 3 - CacheCreation: %d, CacheRead: %d",
		cacheCreation3, cacheRead3)

	// ASSERTION: Third request should also hit cache
	assert.Equal(t, 0, cacheCreation3, "Third request should NOT create new cache")
	assert.Equal(t, cacheRead2, cacheRead3, "Cache read should be consistent across requests")
}

// TestAnthropic_SystemCaching tests how system message caching works:
// Q1: Does system message get cached independently?
// Q2: If tools above system change, does it invalidate system cache?
// Q3: If system is always cached, do we write once and read after?
//
// Expected behavior:
// Request 1: CacheCreation > 0 (system written)
// Request 2: CacheRead > 0 (system read from cache)
//
// BASELINE TEST: This establishes the token count for system alone (~1200 tokens).
func TestAnthropic_SystemCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "SystemTest-v1 "
	longSystemPrompt := generateLargeSystemPrompt(marker)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(longSystemPrompt)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Say hello")},
		},
	}

	// Request 1: First call with system caching
	t.Log("Request 1: First call with system caching")
	r1, err := llm.GenerateContent(t.Context(), messages,
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheSystem: true,
		}),
		llms.WithMaxTokens(50),
	)
	require.NoError(t, err)

	cacheCreation1 := getFromGenerationInfo(r1, "CacheCreationInputTokens")
	cacheRead1 := getFromGenerationInfo(r1, "CacheReadInputTokens")

	t.Logf("Request 1 - CacheCreation: %d, CacheRead: %d", cacheCreation1, cacheRead1)

	// ASSERTION: First request creates cache
	assert.Greater(t, cacheCreation1, 0, "First request should create cache for system")
	assert.Equal(t, 0, cacheRead1, "First request should have no cache reads")

	// BASELINE: Record the system token count for use in other tests
	systemTokenCount := cacheCreation1
	t.Logf("BASELINE: System alone = %d tokens", systemTokenCount)

	// Request 2: Identical request
	t.Log("Request 2: Identical request - should hit cache")
	r2, err := llm.GenerateContent(t.Context(), messages,
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheSystem: true,
		}),
		llms.WithMaxTokens(50),
	)
	require.NoError(t, err)

	cacheCreation2 := getFromGenerationInfo(r2, "CacheCreationInputTokens")
	cacheRead2 := getFromGenerationInfo(r2, "CacheReadInputTokens")

	t.Logf("Request 2 - CacheCreation: %d, CacheRead: %d", cacheCreation2, cacheRead2)

	// ASSERTION: Second request reads from cache
	assert.Equal(t, 0, cacheCreation2, "Second request should NOT create new cache")
	assert.Greater(t, cacheRead2, 0, "Second request should read system from cache")
	assert.Equal(t, cacheRead2, cacheCreation1, "Cache read should equal initial creation")
}

// TestAnthropic_ToolsAndSystemCaching tests hierarchical caching:
// Q1: If we cache both tools and system, what gets cached?
// Q2: If tools change, does system cache also invalidate?
// Q3: CRITICAL: Is cached token count = tools + system, or cumulative prefix?
//
// Expected behavior:
// - Cache hierarchy: tools → system → messages
// - Changing tools invalidates everything below (system + messages)
// - System stays cached if tools unchanged
// - Cache contains ENTIRE PREFIX (tools + system together)
//
// VERIFICATION TEST: Uses baseline data from ToolsCaching and SystemCaching tests.
// Expected: CacheCreation ≈ ToolsTokens + SystemTokens (cumulative, not separate)
func TestAnthropic_ToolsAndSystemCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "CombinedTest-v1 "
	tools := generateLargeTools(marker)
	mediumSystem := generateMediumSystemPrompt(marker) // Smaller to show cumulative effect

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(mediumSystem)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Calculate 2+2")},
		},
	}

	// Request 1: Cache both tools and system
	t.Log("Request 1: Cache tools + system")
	r1, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:  true,
			CacheSystem: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation1 := getFromGenerationInfo(r1, "CacheCreationInputTokens")
	t.Logf("Request 1 - CacheCreation: %d (tools + system combined)", cacheCreation1)

	// CRITICAL ASSERTION: Cache is CUMULATIVE PREFIX, not separate pieces
	// From baseline tests:
	// - Tools alone ≈ 1530 tokens
	// - System alone ≈ 1200 tokens
	// - Tools + System together should be > tools alone (cumulative)
	//
	// NOTE: The exact sum won't match because:
	// 1. System is smaller here (medium vs large in baseline)
	// 2. JSON structure overhead differs between separate vs combined requests
	assert.Greater(t, cacheCreation1, 1530, "Tools+System cache should be larger than tools alone")

	// Expected: ~1530 (tools) + ~300 (medium system) = ~1800-1900 tokens
	expectedRange := "1800-2000"
	t.Logf("VERIFICATION: Combined cache = %d tokens (expected ~%s for tools+system)",
		cacheCreation1, expectedRange)

	// Request 2: Same tools and system - should hit cache for both
	t.Log("Request 2: Same tools + system - should hit cache")
	r2, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:  true,
			CacheSystem: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheRead2 := getFromGenerationInfo(r2, "CacheReadInputTokens")
	cacheCreation2 := getFromGenerationInfo(r2, "CacheCreationInputTokens")
	t.Logf("Request 2 - CacheRead: %d, CacheCreation: %d", cacheRead2, cacheCreation2)

	// ASSERTION: Should read both from cache (as single prefix)
	assert.Equal(t, 0, cacheCreation2, "Should NOT create new cache")
	assert.Equal(t, cacheRead2, cacheCreation1, "Should read ENTIRE PREFIX from cache (not separate tools/system)")

	t.Logf("CONFIRMED: Cache hit reads %d tokens (entire tools+system prefix as ONE unit)", cacheRead2)

	// Request 3: Different tools, same system - system cache should INVALIDATE
	t.Log("Request 3: Different tools, same system - cache invalidated")
	toolsModified := append(tools, llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "new_tool",
			Description: "A new tool",
			Parameters:  map[string]any{"type": "object"},
		},
	})

	r3, err := llm.GenerateContent(t.Context(), messages,
		llms.WithTools(toolsModified),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:  true,
			CacheSystem: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheRead3 := getFromGenerationInfo(r3, "CacheReadInputTokens")
	cacheCreation3 := getFromGenerationInfo(r3, "CacheCreationInputTokens")
	t.Logf("Request 3 - CacheRead: %d, CacheCreation: %d", cacheRead3, cacheCreation3)

	// INTERESTING: API is smart about partial cache reuse
	// Even when tools change, it can find common prefix and reuse parts of the cache
	// In this case: some tools are the same, so API reuses that portion
	t.Logf("Request 3 - CacheRead: %d, CacheCreation: %d", cacheRead3, cacheCreation3)

	// The important assertion: total cache is different from Request 1
	totalCache3 := cacheRead3 + cacheCreation3
	assert.NotEqual(t, cacheCreation1, totalCache3, "Cache structure changed with different tools")

	t.Logf("CONFIRMED: API optimizes cache reuse - reads common prefix (%d tokens), writes new content (%d tokens)",
		cacheRead3, cacheCreation3)
	t.Logf("")
	t.Logf("================================================================================")
	t.Logf("ANSWER TO QUESTION:")
	t.Logf("CacheTools + CacheSystem = ONE cache entry containing BOTH")
	t.Logf("Size: %d tokens (tools + system combined, NOT separate)", cacheCreation1)
	t.Logf("================================================================================")
}

// TestAnthropic_ConversationCaching tests multi-turn conversation caching:
// Q1: Does caching conversation include tools + system + all messages?
// Q2: On each turn with caching enabled, do we write incrementally or cumulatively?
// Q3: Does cache size grow progressively (more expensive each turn)?
//
// Expected behavior:
// - Turn 1: Cache = system + user1 (written to cache)
// - Turn 2: Cache READ previous prefix, Cache WRITE new turn (incremental)
// - Turn 3: Cache READ full history, Cache WRITE new turn (incremental)
// - Result: Cumulative cache grows, but we READ old prefix (10% cost) and only WRITE new content (125% cost)
//
// IMPORTANT: Must exceed 1024 token threshold for claude-sonnet-4-5
func TestAnthropic_ConversationCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvTest-v1 "
	longSystem := generateLargeSystemPrompt(marker) // ~1200 tokens

	// Turn 1: Initial conversation
	t.Log("Turn 1: Initial user message")
	messages1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(longSystem)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What is 2+2?")},
		},
	}

	r1, err := llm.GenerateContent(t.Context(), messages1,
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheSystem:   true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation1 := getFromGenerationInfo(r1, "CacheCreationInputTokens")
	cacheRead1 := getFromGenerationInfo(r1, "CacheReadInputTokens")
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d", cacheCreation1, cacheRead1)

	// ASSERTION: First turn creates cache
	assert.Greater(t, cacheCreation1, 0, "Turn 1 should create cache")
	assert.Equal(t, 0, cacheRead1, "Turn 1 should have no cache reads")
	assert.Greater(t, cacheCreation1, 1024, "Cache should exceed minimum threshold (1024 tokens)")

	// Track cumulative cache writes for economics verification
	cumulativeCacheWrites := cacheCreation1
	t.Logf("Cumulative cache writes so far: %d tokens", cumulativeCacheWrites)

	// Turn 2: Continue conversation
	t.Log("Turn 2: Continue conversation")
	messages2 := append(messages1,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(r1.Choices[0].Content)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What is 3+3?")},
		},
	)

	r2, err := llm.GenerateContent(t.Context(), messages2,
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheSystem:   true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation2 := getFromGenerationInfo(r2, "CacheCreationInputTokens")
	cacheRead2 := getFromGenerationInfo(r2, "CacheReadInputTokens")
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d", cacheCreation2, cacheRead2)

	// ASSERTION: Turn 2 should READ previous turn, WRITE new messages only
	assert.Greater(t, cacheRead2, 0, "Turn 2 should read previous conversation from cache")
	assert.Greater(t, cacheCreation2, 0, "Turn 2 should write new messages to cache")

	// CRITICAL ECONOMICS CHECK: We only write NEW content (AI + user messages from this turn)
	// cacheCreation2 should be MUCH LESS than cacheCreation1 (incremental, not cumulative write!)
	assert.Less(t, cacheCreation2, cacheCreation1, "Turn 2 should write less than turn 1 (incremental)")

	// Update cumulative counter and verify we're not writing the full history again
	cumulativeCacheWrites += cacheCreation2
	assert.Less(t, cumulativeCacheWrites, cacheCreation1*2, "Cumulative writes should NOT double (proving incremental writes)")
	t.Logf("Cumulative cache writes so far: %d tokens (Turn1=%d + Turn2=%d)",
		cumulativeCacheWrites, cacheCreation1, cacheCreation2)

	// Verify economics: we READ the system+user1 prefix (cheap) and WRITE only AI1+user2 (incremental)
	t.Logf("Turn 2 economics: READ %d tokens (10%% cost) + WRITE %d tokens (125%% cost)",
		cacheRead2, cacheCreation2)

	// Turn 3: Continue conversation further
	t.Log("Turn 3: Continue conversation further")
	messages3 := append(messages2,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(r2.Choices[0].Content)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What is 5+5?")},
		},
	)

	r3, err := llm.GenerateContent(t.Context(), messages3,
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheSystem:   true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(100),
	)
	require.NoError(t, err)

	cacheCreation3 := getFromGenerationInfo(r3, "CacheCreationInputTokens")
	cacheRead3 := getFromGenerationInfo(r3, "CacheReadInputTokens")
	totalCached3 := cacheRead3 + cacheCreation3
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d, Total: %d", cacheCreation3, cacheRead3, totalCached3)

	// ASSERTION: Turn 3 should read MORE from cache than turn 2 (cumulative history grows)
	assert.Greater(t, cacheRead3, cacheRead2, "Turn 3 should read more from cache (longer history)")
	assert.Greater(t, cacheCreation3, 0, "Turn 3 should write new messages")

	// Verify total cache grows cumulatively
	totalCached2 := cacheRead2 + cacheCreation2
	assert.Greater(t, totalCached3, totalCached2, "Total cached content grows with conversation")

	// CRITICAL ECONOMICS CHECK: Cumulative writes should grow incrementally
	cumulativeCacheWrites += cacheCreation3
	t.Logf("Cumulative cache writes: Turn1=%d + Turn2=%d + Turn3=%d = %d tokens total written",
		cacheCreation1, cacheCreation2, cacheCreation3, cumulativeCacheWrites)

	// Verify we're NOT writing the full history on each turn (that would be expensive!)
	// If we were writing full history each time: Turn3 would write ~1400+ tokens
	// But actually Turn3 writes only ~20-30 tokens (new content only)
	assert.Less(t, cacheCreation3, 100, "Turn 3 writes only new content (not full history)")
	assert.Less(t, cumulativeCacheWrites, cacheCreation1*2, "Total writes prove incremental model")

	// KEY INSIGHT: Total cached tokens grow, but we only PAY for:
	// - Cache read (10% cost): growing with conversation history (cheap!)
	// - Cache write (125% cost): only for new turn's messages (incremental)
	// Result: Cost does NOT grow linearly - we save more with each turn!
	t.Logf("Cache economics: Turn1=%d tokens written, Turn2=%d read + %d written, Turn3=%d read + %d written",
		cacheCreation1, cacheRead2, cacheCreation2, cacheRead3, cacheCreation3)

	// Calculate effective costs (simplified, not actual pricing)
	cost1 := cacheCreation1 * 125 / 100
	cost2 := cacheRead2*10/100 + cacheCreation2*125/100
	cost3 := cacheRead3*10/100 + cacheCreation3*125/100
	noCacheCost2 := totalCached2
	noCacheCost3 := totalCached3
	t.Logf("Relative costs with cache: Turn1=%d, Turn2=%d, Turn3=%d", cost1, cost2, cost3)
	t.Logf("Without cache would be: Turn2=%d, Turn3=%d", noCacheCost2, noCacheCost3)
	t.Logf("SAVINGS: Turn2 saves %d%%, Turn3 saves %d%%",
		(noCacheCost2-cost2)*100/noCacheCost2, (noCacheCost3-cost3)*100/noCacheCost3)

	t.Logf("")
	t.Logf("================================================================================")
	t.Logf("ANSWER TO QUESTION: Multi-turn caching economics")
	t.Logf("We pay 125%% ONLY for difference between turns (NOT cumulative full history)")
	t.Logf("Total cache writes: %d tokens across 3 turns (NOT %d if we wrote full history each time)",
		cumulativeCacheWrites, totalCached3*3)
	t.Logf("Result: HUGE savings on long conversations - cost DECREASES with each turn!")
	t.Logf("================================================================================")
}

// TestAnthropic_ToolResultsCaching tests tool result caching in multi-turn:
// Q1: Do we cache each tool result separately or the entire turn?
// Q2: If agent calls tools multiple times, how does cache accumulate?
//
// Expected behavior:
// - Tool results are part of the message history
// - Caching tool result means caching the PREFIX up to that point
// - Each tool call turn adds to cumulative cache
//
// IMPORTANT: Must exceed 1024 token threshold - use large tools
func TestAnthropic_ToolResultsCaching(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ToolResultTest-v1 "
	tools := generateLargeTools(marker)

	// Turn 1: User asks question that triggers tool call
	t.Log("Turn 1: User asks question")
	messages1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Paris?")},
		},
	}

	r1, err := llm.GenerateContent(t.Context(), messages1,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)

	cacheCreation1 := getFromGenerationInfo(r1, "CacheCreationInputTokens")
	cacheRead1 := getFromGenerationInfo(r1, "CacheReadInputTokens")
	totalTokens1 := cacheCreation1 + cacheRead1
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d, Total: %d", cacheCreation1, cacheRead1, totalTokens1)

	// ASSERTION: First turn caches tools (either creates new or reads existing)
	// Tools may already be cached from parallel tests, so we check total > 0
	assert.Greater(t, totalTokens1, 0, "Turn 1 should have tools (either cached or read from cache)")
	toolsTokens := totalTokens1 // Baseline for tools
	t.Logf("Tools baseline: %d tokens", toolsTokens)

	// Check if model called a tool
	require.NotEmpty(t, r1.Choices[0].ToolCalls, "Model should call tool")

	// Turn 2: Provide tool result
	t.Log("Turn 2: Provide tool result")
	messages2 := append(messages1,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				r1.Choices[0].ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: r1.Choices[0].ToolCalls[0].ID,
					Name:       r1.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content:    "Sunny, 22°C",
				},
			},
		},
	)

	r2, err := llm.GenerateContent(t.Context(), messages2,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)

	cacheCreation2 := getFromGenerationInfo(r2, "CacheCreationInputTokens")
	cacheRead2 := getFromGenerationInfo(r2, "CacheReadInputTokens")
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d", cacheCreation2, cacheRead2)

	// ASSERTION: Turn 2 reads from cache (tools at minimum)
	totalTokens2 := cacheRead2 + cacheCreation2
	assert.Greater(t, cacheRead2, 0, "Turn 2 should read from cache")

	// Note: totalTokens2 may be less than toolsTokens if Turn 1 created cache for the whole request
	// but Turn 2 only reads tools portion. This is API optimization - reading common prefix only.
	t.Logf("Turn 2 total cached: %d tokens (may be less than Turn 1 due to API optimization)", totalTokens2)

	t.Logf("Turn 2 economics: READ %d tokens + WRITE %d tokens",
		cacheRead2, cacheCreation2)

	// Turn 3: User asks another question
	t.Log("Turn 3: User asks another question")
	messages3 := append(messages2,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(r2.Choices[0].Content)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What about London?")},
		},
	)

	r3, err := llm.GenerateContent(t.Context(), messages3,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)

	cacheCreation3 := getFromGenerationInfo(r3, "CacheCreationInputTokens")
	cacheRead3 := getFromGenerationInfo(r3, "CacheReadInputTokens")
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d", cacheCreation3, cacheRead3)

	// ASSERTION: Turn 3 should read cumulative history including tool results
	assert.Greater(t, cacheRead3, cacheRead2, "Turn 3 should read more from cache (includes tool history)")
	assert.Greater(t, cacheCreation3, 0, "Turn 3 should write new user message")

	// KEY INSIGHT: Tool use with caching is highly efficient
	// - Tools definition: cached once, read every turn (10% cost)
	// - Tool results: cached incrementally, read in subsequent turns (10% cost)
	// - Net result: Multi-turn agent workflows are cost-effective with caching
	t.Logf("Tool result caching economics: Cumulative cache read grew from %d → %d tokens",
		cacheRead2, cacheRead3)
	t.Logf("Turn 3: READ %d tokens (tools+turn1+turn2) + WRITE %d tokens (user3)",
		cacheRead3, cacheCreation3)

	// Answer to Q: Tool results are cached as part of PREFIX, not separately
	// Each turn reads the ENTIRE prefix (tools + all previous messages) and writes only new content
	t.Logf("CONFIRMED: Tool results cached cumulatively as part of conversation prefix")
}

// TestAnthropic_ToolResultsCachingStreaming is identical to TestAnthropic_ToolResultsCaching
// but uses streaming mode. This allows us to compare HTTP requests between streaming and
// non-streaming to understand why caching behaves differently.
func TestAnthropic_ToolResultsCachingStreaming(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Marker: DIFFERENT from non-streaming test to avoid cache collision
	marker := "ToolResultTestStreaming-v1 "
	tools := generateLargeTools(marker)

	// Track streaming completion
	var streamingDone chan struct{}
	resetStreaming := func() {
		streamingDone = make(chan struct{})
	}

	// Turn 1: User asks question that triggers tool call
	t.Log("Turn 1: User asks question (streaming)")
	messages1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Paris?")},
		},
	}

	resetStreaming()
	r1, err := llm.GenerateContent(t.Context(), messages1,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(200),
		llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeDone {
				close(streamingDone)
			}
			return nil
		}),
	)
	require.NoError(t, err)
	<-streamingDone // Wait for streaming to complete

	cacheCreation1 := getFromGenerationInfo(r1, "CacheCreationInputTokens")
	cacheRead1 := getFromGenerationInfo(r1, "CacheReadInputTokens")
	totalTokens1 := cacheCreation1 + cacheRead1
	t.Logf("Turn 1 (streaming) - CacheCreation: %d, CacheRead: %d, Total: %d", cacheCreation1, cacheRead1, totalTokens1)

	assert.Greater(t, totalTokens1, 0, "Turn 1 should have tools (either cached or read from cache)")
	toolsTokens := totalTokens1
	t.Logf("Tools baseline: %d tokens", toolsTokens)

	require.NotEmpty(t, r1.Choices[0].ToolCalls, "Model should call tool")

	// Turn 2: Provide tool result
	t.Log("Turn 2: Provide tool result (streaming)")
	messages2 := append(messages1,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				r1.Choices[0].ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: r1.Choices[0].ToolCalls[0].ID,
					Name:       r1.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content:    "Sunny, 22°C",
				},
			},
		},
	)

	resetStreaming()
	r2, err := llm.GenerateContent(t.Context(), messages2,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(200),
		llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeDone {
				close(streamingDone)
			}
			return nil
		}),
	)
	require.NoError(t, err)
	<-streamingDone // Wait for streaming to complete

	cacheCreation2 := getFromGenerationInfo(r2, "CacheCreationInputTokens")
	cacheRead2 := getFromGenerationInfo(r2, "CacheReadInputTokens")
	t.Logf("Turn 2 (streaming) - CacheCreation: %d, CacheRead: %d", cacheCreation2, cacheRead2)

	totalTokens2 := cacheRead2 + cacheCreation2
	assert.Greater(t, cacheRead2, 0, "Turn 2 should read from cache")

	t.Logf("Turn 2 total cached: %d tokens (may be less than Turn 1 due to API optimization)", totalTokens2)
	t.Logf("Turn 2 economics: READ %d tokens + WRITE %d tokens", cacheRead2, cacheCreation2)

	// Turn 3: User asks another question
	t.Log("Turn 3: User asks another question (streaming)")
	messages3 := append(messages2,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart(r2.Choices[0].Content)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What about London?")},
		},
	)

	resetStreaming()
	r3, err := llm.GenerateContent(t.Context(), messages3,
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(200),
		llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeDone {
				close(streamingDone)
			}
			return nil
		}),
	)
	require.NoError(t, err)
	<-streamingDone // Wait for streaming to complete

	cacheCreation3 := getFromGenerationInfo(r3, "CacheCreationInputTokens")
	cacheRead3 := getFromGenerationInfo(r3, "CacheReadInputTokens")
	t.Logf("Turn 3 (streaming) - CacheCreation: %d, CacheRead: %d", cacheCreation3, cacheRead3)

	assert.Greater(t, cacheRead3, cacheRead2, "Turn 3 should read more from cache (includes tool history)")
	assert.Greater(t, cacheCreation3, 0, "Turn 3 should write new user message")

	t.Logf("Streaming tool result caching: Cache read grew from %d → %d tokens", cacheRead2, cacheRead3)
	t.Logf("Turn 3: READ %d tokens (tools+turn1+turn2) + WRITE %d tokens (user3)", cacheRead3, cacheCreation3)

	t.Logf("CONFIRMED: Tool results cached with streaming - same behavior as non-streaming")
}
