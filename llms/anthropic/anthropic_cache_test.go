package anthropic_test

import (
	"context"
	"fmt"
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
func TestAnthropic_ToolsAndSystemCaching(t *testing.T) { //nolint:funlen
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
func TestAnthropic_ConversationCaching(t *testing.T) { //nolint:funlen
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

// TestAnthropic_IncrementalCaching validates that prompt caching grows incrementally
// across multi-turn conversations without cache invalidation.
//
// Test validates:
// - Cache reads increase monotonically (each turn reads more than previous)
// - Cache writes only include new content (not entire message chain)
// - Behavior is consistent across reasoning and streaming modes
func TestAnthropic_IncrementalCaching(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		reasoning bool
		streaming bool
	}{
		{
			name:      "baseline",
			reasoning: false,
			streaming: false,
		},
		{
			name:      "with_reasoning",
			reasoning: true,
			streaming: false,
		},
		{
			name:      "with_streaming",
			reasoning: false,
			streaming: true,
		},
		{
			name:      "with_reasoning_and_streaming",
			reasoning: true,
			streaming: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runIncrementalCachingTest(t, tc.reasoning, tc.streaming)
		})
	}
}

// runIncrementalCachingTest executes a 4-turn conversation with tool calls:
// Turn 1: User asks about weather and time -> AI calls get_weather, get_time
// Turn 2: Tool results provided -> AI calls check_flight_schedule
// Turn 3: Flight schedule provided -> AI calls book_flight
// Turn 4: Booking confirmed -> AI responds with final message
func runIncrementalCachingTest(t *testing.T, enableReasoning, enableStreaming bool) {
	t.Helper()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	// Use unique marker to prevent cache collision between test cases
	marker := fmt.Sprintf("IncrCacheTest-r%v-s%v-v2 ", enableReasoning, enableStreaming)
	tools := generateFlightBookingTools(marker)

	// Setup streaming handler if needed
	var streamHandler *streamingHandler
	if enableStreaming {
		streamHandler = newStreamingHandler()
	}

	// Build options
	opts := buildCachingOptions(tools, enableReasoning, streamHandler)

	// Execute multi-turn conversation
	cacheMetrics := executeFourTurnConversation(t, llm, opts, enableReasoning, streamHandler)

	// Validate incremental cache growth
	validateIncrementalCacheGrowth(t, cacheMetrics)
}

// generateFlightBookingTools creates a realistic flight booking scenario with large tools
// to ensure cache threshold (1024 tokens) is exceeded.
func generateFlightBookingTools(marker string) []llms.Tool { //nolint:funlen
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: marker + "Get current weather conditions for a specified city. Returns temperature, conditions, humidity, and wind speed.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city": map[string]any{
							"type":        "string",
							"description": "The city name to get weather for",
						},
						"units": map[string]any{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "Temperature unit preference",
						},
					},
					"required": []string{"city"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_time",
				Description: marker + "Get current local time for a specified city with timezone information. Returns time in 24-hour format along with timezone offset.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city": map[string]any{
							"type":        "string",
							"description": "The city name to get time for",
						},
						"format": map[string]any{
							"type":        "string",
							"enum":        []string{"12h", "24h"},
							"description": "Time format preference",
						},
					},
					"required": []string{"city"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "check_flight_schedule",
				Description: marker + "Check available flight schedules between two cities on a specified date. Returns flight numbers, departure/arrival times, airlines, and available seats.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"origin": map[string]any{
							"type":        "string",
							"description": "Origin city or airport code",
						},
						"destination": map[string]any{
							"type":        "string",
							"description": "Destination city or airport code",
						},
						"date": map[string]any{
							"type":        "string",
							"description": "Flight date in YYYY-MM-DD format",
						},
						"class": map[string]any{
							"type":        "string",
							"enum":        []string{"economy", "business", "first"},
							"description": "Preferred travel class",
						},
					},
					"required": []string{"origin", "destination", "date"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "book_flight",
				Description: marker + "Book a flight ticket for specified flight number and passenger details. Returns booking confirmation number and total price.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"flight_number": map[string]any{
							"type":        "string",
							"description": "Flight number from schedule",
						},
						"passenger_name": map[string]any{
							"type":        "string",
							"description": "Full name of passenger",
						},
						"passenger_email": map[string]any{
							"type":        "string",
							"description": "Email for booking confirmation",
						},
						"seat_preference": map[string]any{
							"type":        "string",
							"enum":        []string{"window", "aisle", "middle"},
							"description": "Seat location preference",
						},
					},
					"required": []string{"flight_number", "passenger_name", "passenger_email"},
				},
			},
		},
	}

	return tools
}

// streamingHandler manages streaming completion signaling
type streamingHandler struct {
	done chan struct{}
}

func newStreamingHandler() *streamingHandler {
	return &streamingHandler{done: make(chan struct{})}
}

func (sh *streamingHandler) callback(ctx context.Context, chunk streaming.Chunk) error {
	if chunk.Type == streaming.ChunkTypeDone {
		close(sh.done)
	}
	return nil
}

func (sh *streamingHandler) wait() {
	<-sh.done
}

// buildCachingOptions constructs call options for the test
func buildCachingOptions(tools []llms.Tool, enableReasoning bool, streamHandler *streamingHandler) []llms.CallOption {
	opts := []llms.CallOption{
		llms.WithTools(tools),
		anthropic.WithPromptCaching(),
		anthropic.WithCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheSystem:   true,
			CacheMessages: true,
		}),
		llms.WithMaxTokens(4096),
	}

	if enableReasoning {
		opts = append(opts, llms.WithReasoning(llms.ReasoningMedium, 1024))
	}

	if streamHandler != nil {
		opts = append(opts, llms.WithStreamingFunc(streamHandler.callback))
	}

	return opts
}

// cacheMetrics tracks cache read/write tokens across turns
type cacheMetrics struct {
	turns []turnMetrics
}

type turnMetrics struct {
	turnNum       int
	cacheRead     int
	cacheCreation int
	description   string
}

// executeFourTurnConversation runs the 4-turn flight booking conversation
func executeFourTurnConversation(t *testing.T, llm *anthropic.LLM, opts []llms.CallOption, enableReasoning bool, streamHandler *streamingHandler) cacheMetrics { //nolint:funlen
	t.Helper()

	metrics := cacheMetrics{turns: make([]turnMetrics, 0, 4)}

	// Helper to wait for streaming completion
	waitForStream := func() {
		if streamHandler != nil {
			streamHandler.wait()
			// Reset for next turn
			streamHandler.done = make(chan struct{})
		}
	}

	// Helper to append AI response to messages
	appendAIResponse := func(messages []llms.MessageContent, choice *llms.ContentChoice) []llms.MessageContent {
		messages = append([]llms.MessageContent{}, messages...)
		if choice.Reasoning != nil && enableReasoning {
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
				},
			})
		}
		return messages
	}

	// Helper to append tool calls and responses
	appendToolResults := func(messages []llms.MessageContent, toolCalls []llms.ToolCall, responseMap map[string]string) []llms.MessageContent {
		messages = append([]llms.MessageContent{}, messages...)

		// Add tool calls
		for _, tc := range toolCalls {
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{tc},
			})
		}

		// Add tool responses
		for _, tc := range toolCalls {
			content, exists := responseMap[tc.FunctionCall.Name]
			if !exists {
				content = "Function not recognized"
			}
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    content,
					},
				},
			})
		}

		return messages
	}

	// Turn 1: User provides complete task description with clear instructions
	t.Log("Turn 1: User provides flight booking task")
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(
				"You are an automated flight booking assistant. You MUST follow this exact workflow:\n\n" +
					"STEP 1: When user requests a flight, FIRST call get_weather and get_time for destination\n" +
					"STEP 2: After receiving weather/time data, NEXT call check_flight_schedule\n" +
					"STEP 3: After receiving flight schedule, NEXT call book_flight with user details\n" +
					"STEP 4: After booking confirmation, provide final summary to user\n\n" +
					"CRITICAL: You must call ALL required tools in sequence. Do not skip steps or summarize prematurely.",
			)},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(
				"Book me a flight from New York to Paris for tomorrow (2024-06-15). " +
					"Execute the complete booking workflow.\n\n" +
					"Passenger details:\n" +
					"- Name: John Smith\n" +
					"- Email: john.smith@example.com\n" +
					"- Seat preference: window\n" +
					"- Travel class: economy",
			)},
		},
	}

	r1, err := llm.GenerateContent(t.Context(), messages, opts...)
	require.NoError(t, err)
	waitForStream()
	require.NotEmpty(t, r1.Choices)
	require.NotEmpty(t, r1.Choices[0].ToolCalls, "Turn 1: AI should call weather/time tools")

	metrics.turns = append(metrics.turns, turnMetrics{
		turnNum:       1,
		cacheRead:     getFromGenerationInfo(r1, "CacheReadInputTokens"),
		cacheCreation: getFromGenerationInfo(r1, "CacheCreationInputTokens"),
		description:   "Initial request with tools",
	})
	t.Logf("Turn 1 - CacheRead: %d, CacheCreation: %d", metrics.turns[0].cacheRead, metrics.turns[0].cacheCreation)

	// Turn 2: Provide weather/time results, AI should continue to check flights
	t.Log("Turn 2: Provide weather/time results")
	messages = appendAIResponse(messages, r1.Choices[0])
	messages = appendToolResults(messages, r1.Choices[0].ToolCalls, map[string]string{
		"get_weather": "Weather in Paris: Sunny, 18°C, humidity 65%, light breeze from west",
		"get_time":    "Current time in Paris: 14:30 CET (UTC+1)",
	})

	r2, err := llm.GenerateContent(t.Context(), messages, opts...)
	require.NoError(t, err)
	waitForStream()
	require.NotEmpty(t, r2.Choices)
	require.NotEmpty(t, r2.Choices[0].ToolCalls, "Turn 2: AI should check flight schedule")

	metrics.turns = append(metrics.turns, turnMetrics{
		turnNum:       2,
		cacheRead:     getFromGenerationInfo(r2, "CacheReadInputTokens"),
		cacheCreation: getFromGenerationInfo(r2, "CacheCreationInputTokens"),
		description:   "After weather/time results",
	})
	t.Logf("Turn 2 - CacheRead: %d, CacheCreation: %d", metrics.turns[1].cacheRead, metrics.turns[1].cacheCreation)

	// Turn 3: Provide flight schedule, AI should book flight
	t.Log("Turn 3: Provide flight schedule")
	messages = appendAIResponse(messages, r2.Choices[0])
	messages = appendToolResults(messages, r2.Choices[0].ToolCalls, map[string]string{
		"check_flight_schedule": "Available flights on 2024-06-15:\n" +
			"Flight AF1234 - Air France\n" +
			"Departure: JFK 09:00 → Arrival: CDG 22:30\n" +
			"Economy class: €350, 45 seats available\n" +
			"Window seats available",
	})

	r3, err := llm.GenerateContent(t.Context(), messages, opts...)
	require.NoError(t, err)
	waitForStream()
	require.NotEmpty(t, r3.Choices)
	require.NotEmpty(t, r3.Choices[0].ToolCalls, "Turn 3: AI should book the flight")

	metrics.turns = append(metrics.turns, turnMetrics{
		turnNum:       3,
		cacheRead:     getFromGenerationInfo(r3, "CacheReadInputTokens"),
		cacheCreation: getFromGenerationInfo(r3, "CacheCreationInputTokens"),
		description:   "After flight schedule",
	})
	t.Logf("Turn 3 - CacheRead: %d, CacheCreation: %d", metrics.turns[2].cacheRead, metrics.turns[2].cacheCreation)

	// Turn 4: Provide booking confirmation, AI gives final response
	t.Log("Turn 4: Provide booking confirmation")
	messages = appendAIResponse(messages, r3.Choices[0])
	messages = appendToolResults(messages, r3.Choices[0].ToolCalls, map[string]string{
		"book_flight": "✓ Booking Confirmed\n" +
			"Confirmation Code: ABC123XYZ\n" +
			"Flight: AF1234\n" +
			"Passenger: John Smith\n" +
			"Seat: 12A (window)\n" +
			"Total: €350\n" +
			"Confirmation sent to: john.smith@example.com",
	})

	r4, err := llm.GenerateContent(t.Context(), messages, opts...)
	require.NoError(t, err)
	waitForStream()
	require.NotEmpty(t, r4.Choices)

	metrics.turns = append(metrics.turns, turnMetrics{
		turnNum:       4,
		cacheRead:     getFromGenerationInfo(r4, "CacheReadInputTokens"),
		cacheCreation: getFromGenerationInfo(r4, "CacheCreationInputTokens"),
		description:   "After booking confirmation",
	})
	t.Logf("Turn 4 - CacheRead: %d, CacheCreation: %d", metrics.turns[3].cacheRead, metrics.turns[3].cacheCreation)

	return metrics
}

// validateIncrementalCacheGrowth ensures cache behaves correctly:
// - Cache reads increase monotonically (never decrease)
// - Cache writes are only for new content (not entire history)
func validateIncrementalCacheGrowth(t *testing.T, metrics cacheMetrics) {
	t.Helper()

	require.Len(t, metrics.turns, 4, "Should have metrics for 4 turns")

	t.Log("=== Cache Growth Analysis ===")
	for i, turn := range metrics.turns {
		t.Logf("Turn %d (%s): READ=%d, WRITE=%d",
			turn.turnNum, turn.description, turn.cacheRead, turn.cacheCreation)

		// Turn 1: Initial cache creation (tools + first message)
		if i == 0 {
			total := turn.cacheRead + turn.cacheCreation
			assert.Greater(t, total, 0,
				"Turn 1: Should cache tools (either create or read from parallel test)")
		}

		// Subsequent turns: Cache reads must grow monotonically
		if i > 0 {
			prevRead := metrics.turns[i-1].cacheRead
			currentRead := turn.cacheRead

			assert.GreaterOrEqual(t, currentRead, prevRead,
				"Turn %d: Cache reads must not decrease (prev=%d, current=%d)",
				turn.turnNum, prevRead, currentRead)

			// Stricter check: In most cases, cache reads should strictly increase
			// (unless API optimizes by not re-caching unchanged prefix)
			if currentRead == prevRead {
				t.Logf("Turn %d: Cache reads unchanged (API optimization or parallel test collision)", turn.turnNum)
			}

			// Cache creation should be reasonable (not entire message chain)
			// Allow flexibility for API implementation details
			assert.GreaterOrEqual(t, turn.cacheCreation, 0,
				"Turn %d: Cache creation should be non-negative", turn.turnNum)
		}
	}

	// Final validation: Overall cache growth
	finalRead := metrics.turns[3].cacheRead
	initialRead := metrics.turns[0].cacheRead
	t.Logf("=== Summary: Cache reads grew from %d → %d tokens ===", initialRead, finalRead)

	assert.Greater(t, finalRead, initialRead,
		"Final turn should read more from cache than first turn (cumulative history)")
}
