package googleai

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// GOOGLE GEMINI CACHING TESTS
// ============================================================================
//
// This file contains tests for both Explicit and Implicit caching mechanisms
// provided by Google Gemini API.
//
// KEY FINDINGS (as of January 2026):
//
// 1. EXPLICIT CACHING (via WithCachedContent):
//    ✓ Works correctly for multi-turn conversations
//    ✓ Requires manual cache creation via CachingHelper.CreateCachedContent
//    ✓ Provides guaranteed cost savings (75% discount on cached tokens)
//    ✓ Requires minimum 32,768 tokens (~24,000 words)
//    ✓ Has hourly storage charges based on TTL
//
// 2. IMPLICIT CACHING (automatic):
//    ✓ Works for completely IDENTICAL requests (byte-for-byte HTTP body match)
//    ✓ Works for conversation continuation with gemini-3-flash-preview + reasoning
//    ✗ Does NOT work for conversation continuation WITHOUT reasoning/signature preservation
//    ✓ Requires minimum 1,024 tokens for Gemini 2.5 Flash, 4,096 for Gemini 3 Flash Preview
//    ✓ No storage charges (ephemeral, managed by Google)
//    ⚠️  Requires 15+ seconds delay between requests to establish cache
//
// 3. CRITICAL INSIGHT FOR CONVERSATION CONTINUATION:
//    - Use gemini-3-flash-preview with llms.WithReasoning()
//    - MUST preserve reasoning/signature when adding AI responses:
//      ✓ llms.TextPartWithReasoning(choice.Content, choice.Reasoning)
//      ✗ llms.TextPart(choice.Content) <- will NOT cache prefix!
//    - With proper signature preservation: 89% of tokens cached (~4058/4534)
//
// 4. RECOMMENDATION:
//    - Use IMPLICIT CACHING for multi-turn conversations with gemini-3-flash-preview
//    - Use EXPLICIT CACHING for longer conversations or when using gemini-2.5-flash
//    - See TestGoogleAI_ImplicitCaching_ConversationContinuation for working example
//
// ============================================================================

// TestGoogleAI_ExplicitCaching tests explicit caching with httprr
func TestGoogleAI_ExplicitCaching(t *testing.T) { //nolint:funlen
	// Skip if no credentials and no recording
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)

	// Scrub request bodies and API keys
	rr.ScrubReq(httprr.JsonCompactScrubBody)
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	ctx := t.Context()

	// Create caching helper with httprr client
	helper, err := NewCachingHelper(ctx,
		WithAPIKey(apiKey),
		WithRest(),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	// Create cached content with a large system prompt
	// Note: Google AI requires at least 32,768 tokens (~24,000 words) for explicit caching
	baseContext := `You are an expert code reviewer with deep knowledge of Go best practices.
You should always consider the following aspects in your reviews:

1. Performance: Analyze algorithmic complexity, memory usage, and potential bottlenecks.
2. Security: Check for common vulnerabilities like SQL injection, XSS, buffer overflows.
3. Maintainability: Evaluate code readability, documentation, and adherence to conventions.
4. Testing: Assess test coverage and quality of test cases.
5. Error handling: Verify proper error propagation and handling.
6. Concurrency: Review goroutine usage, channel operations, and synchronization.
7. Resource management: Check for proper cleanup of resources (files, connections, etc).
8. API design: Evaluate interface design and package structure.

`
	// Repeat to reach minimum cache size (~32k tokens requires ~24k words)
	longContext := strings.Repeat(baseContext, 500)

	cached, err := helper.CreateCachedContent(ctx, "gemini-2.0-flash",
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextPart(longContext),
				},
			},
		},
		5*time.Minute,
		"go-code-review-expert",
	)
	require.NoError(t, err, "Failed to create cached content")
	t.Logf("Created cached content: %s", cached.Name)

	// Only try to delete if we're recording (not replaying)
	if os.Getenv("HTTPRR_RECORD") != "" {
		defer func() {
			if err := helper.DeleteCachedContent(ctx, cached.Name); err != nil {
				t.Logf("Failed to delete cached content: %v", err)
			}
		}()
	}

	// Verify cache metadata
	assert.NotEmpty(t, cached.Name)
	assert.Greater(t, cached.UsageMetadata.TotalTokenCount, int32(0), "Cache should contain tokens")
	t.Logf("Cache metadata: TotalTokens=%d", cached.UsageMetadata.TotalTokenCount)

	// Use the cached content in a request
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithRest(),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What are the key things to look for in a Go code review?"),
			},
		},
	}

	resp, err := client.GenerateContent(ctx, messages,
		WithCachedContent(cached.Name),
		llms.WithMaxTokens(200),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	// Verify that cached tokens are reported and non-zero
	genInfo := resp.Choices[0].GenerationInfo
	require.NotNil(t, genInfo, "Generation info should be present")

	cachedTokens, hasCachedTokens := genInfo["PromptCachedTokens"].(int)
	assert.True(t, hasCachedTokens, "CachedTokens should be present in response metadata")
	assert.Greater(t, cachedTokens, 0, "CachedTokens should be greater than 0 for explicit caching")

	t.Logf("Successfully used %d cached tokens", cachedTokens)
	t.Logf("Full response metadata: %+v", genInfo)

	// Verify response content
	assert.NotEmpty(t, resp.Choices[0].Content, "Response content should not be empty")
	t.Logf("Response preview: %s...", resp.Choices[0].Content[:min(100, len(resp.Choices[0].Content))])
}

// TestGoogleAI_ImplicitCaching tests that implicit caching works correctly
// Implicit caching is automatically enabled and reports cache hits via CachedTokens
//
// IMPORTANT DISCOVERY: Based on empirical testing, Google Gemini's implicit caching appears
// to work ONLY for completely IDENTICAL requests (same HTTP body byte-for-byte).
// It does NOT reliably cache prefixes for multi-turn conversation continuation.
//
// This test verifies that at least identical requests trigger implicit caching.
func TestGoogleAI_ImplicitCaching(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)
	rr.ScrubReq(httprr.JsonCompactScrubBody)
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	ctx := t.Context()
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithRest(),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ImplicitTest-v1 "
	largeContext := strings.Repeat(marker+"Go is a compiled language with garbage collection.\n", 500)

	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("Say 'hello'")}},
	}

	// Request 1: Initial request
	r1, err := client.GenerateContent(ctx, msgs, llms.WithModel("gemini-2.5-flash"), llms.WithMaxTokens(50), llms.WithTemperature(0.0))
	require.NoError(t, err)
	c1 := 0
	if ct, ok := r1.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}

	// Request 1: In fresh recording should be 0, but in replay may have cache from previous runs
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens")

	if rr.Recording() {
		// Increased delay to 30 seconds to ensure Google establishes the implicit cache
		t.Log("Waiting 30 seconds for Google to establish implicit cache...")
		time.Sleep(30 * time.Second)
	}

	// Request 2: IDENTICAL to request 1 - should hit cache
	r2, err := client.GenerateContent(ctx, msgs, llms.WithModel("gemini-2.5-flash"), llms.WithMaxTokens(50), llms.WithTemperature(0.0))
	require.NoError(t, err)
	c2 := 0
	if ct, ok := r2.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 MUST have cached tokens for identical request
	assert.Greater(t, c2, 0, "Request 2 (identical) must have cached tokens")

	if rr.Recording() {
		t.Log("Waiting 30 seconds...")
		time.Sleep(30 * time.Second)
	}

	// Request 3: IDENTICAL again - should still hit cache
	r3, err := client.GenerateContent(ctx, msgs, llms.WithModel("gemini-2.5-flash"), llms.WithMaxTokens(50), llms.WithTemperature(0.0))
	require.NoError(t, err)
	c3 := 0
	if ct, ok := r3.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c3 = ct
	}

	// Assert: Request 3 MUST also have cached tokens
	assert.Greater(t, c3, 0, "Request 3 (identical) must have cached tokens")

	// Verify cache consistency
	assert.Equal(t, c2, c3, "Identical requests should have same cached token count")
}

// TestGoogleAI_ImplicitCaching_Streaming tests implicit caching with streaming
// This test mirrors TestGoogleAI_ImplicitCaching but uses streaming mode
func TestGoogleAI_ImplicitCaching_Streaming(t *testing.T) { //nolint:funlen
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)
	rr.ScrubReq(httprr.JsonCompactScrubBody)
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	ctx := t.Context()
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithRest(),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "StreamTest-v1 "
	largeContext := strings.Repeat(marker+"Rust ensures memory safety without garbage collection.\n", 500)

	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("Say 'hello'")}},
	}

	streamFunc := func(content *strings.Builder) llms.CallOption {
		return llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeText {
				content.WriteString(chunk.Content)
			}
			return nil
		})
	}

	// Request 1: Initial streaming request
	var s1 strings.Builder
	r1, err := client.GenerateContent(ctx, msgs, llms.WithModel("gemini-2.5-flash"), llms.WithMaxTokens(50), llms.WithTemperature(0.0), streamFunc(&s1))
	require.NoError(t, err)
	require.NotNil(t, r1.Choices[0].GenerationInfo, "GenerationInfo should be present in streaming")
	c1 := 0
	if ct, ok := r1.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	// Request 1: In fresh recording should be 0, but in replay may have cache from previous runs
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens")

	if rr.Recording() {
		// Increased delay to 30 seconds to ensure Google establishes the implicit cache
		t.Log("Waiting 30 seconds for Google to establish implicit cache...")
		time.Sleep(30 * time.Second)
	}

	// Request 2: IDENTICAL to request 1 - should hit cache
	var s2 strings.Builder
	r2, err := client.GenerateContent(ctx, msgs, llms.WithModel("gemini-2.5-flash"), llms.WithMaxTokens(50), llms.WithTemperature(0.0), streamFunc(&s2))
	require.NoError(t, err)
	require.NotNil(t, r2.Choices[0].GenerationInfo, "GenerationInfo should be present in streaming")
	c2 := 0
	if ct, ok := r2.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 MUST have cached tokens for identical streaming request
	assert.Greater(t, c2, 0, "Request 2 (identical streaming) must have cached tokens")

	if rr.Recording() {
		t.Log("Waiting 30 seconds...")
		time.Sleep(30 * time.Second)
	}

	// Request 3: IDENTICAL again - should still hit cache
	var s3 strings.Builder
	r3, err := client.GenerateContent(ctx, msgs, llms.WithModel("gemini-2.5-flash"), llms.WithMaxTokens(50), llms.WithTemperature(0.0), streamFunc(&s3))
	require.NoError(t, err)
	require.NotNil(t, r3.Choices[0].GenerationInfo, "GenerationInfo should be present in streaming")
	c3 := 0
	if ct, ok := r3.Choices[0].GenerationInfo["PromptCachedTokens"].(int); ok {
		c3 = ct
	}

	// Assert: Request 3 MUST also have cached tokens
	assert.Greater(t, c3, 0, "Request 3 (identical streaming) must have cached tokens")

	// Verify cache consistency
	assert.Equal(t, c2, c3, "Identical streaming requests should have same cached token count")

	// Verify streamed content is not empty
	assert.NotEmpty(t, s1.String(), "Stream 1 should have content")
	assert.NotEmpty(t, s2.String(), "Stream 2 should have content")
	assert.NotEmpty(t, s3.String(), "Stream 3 should have content")
}

// TestGoogleAI_ImplicitCaching_ConversationContinuation tests implicit caching for multi-turn conversations
//
// UPDATED APPROACH: Use gemini-3-flash-preview with proper signature/reasoning preservation.
// The key insight is that conversation continuation requires preserving the model's thought
// signatures when adding AI responses to the conversation history.
//
// Expected behavior:
//   - Request 1: [system, user1] -> establishes cache
//   - Request 2: [system, user1, AI1+signature, user2] -> prefix should be cached
//   - Request 3: [system, user1, AI1+signature, user2, AI2+signature, user3] -> longer prefix cached
func TestGoogleAI_ImplicitCaching_ConversationContinuation(t *testing.T) { //nolint:funlen
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)
	rr.ScrubReq(httprr.JsonCompactScrubBody)
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	ctx := t.Context()
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithRest(),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvTest-v1 "
	largeContext := marker + strings.Repeat("Go is a compiled language with garbage collection. ", 500)

	// Request 1: Initial conversation turn [system, user1]
	msgs := []llms.MessageContent{ //nolint:prealloc
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(largeContext)}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What is 10 + 5?")}},
	}

	r1, err := client.GenerateContent(ctx, msgs,
		llms.WithModel("gemini-3-flash-preview"),
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.0),
		llms.WithReasoning(llms.ReasoningMedium, 512),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r1.Choices)

	choice1 := r1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	// Request 1: first request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 should have 0 cached tokens (first request)")
	assert.Contains(t, choice1.Content, "15", "Response should contain the answer 15")
	assert.NotNil(t, choice1.Reasoning, "Gemini 3 should return reasoning")

	if rr.Recording() {
		t.Log("Waiting 15 seconds for Google to establish implicit cache...")
		time.Sleep(15 * time.Second)
	}

	// Request 2: Continue conversation [system, user1, AI1+signature, user2]
	// CRITICAL: Must preserve reasoning/signature when adding AI response
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

	r2, err := client.GenerateContent(ctx, msgs,
		llms.WithModel("gemini-3-flash-preview"),
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.0),
		llms.WithReasoning(llms.ReasoningMedium, 512),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r2.Choices)

	choice2 := r2.Choices[0]
	c2 := 0
	if ct, ok := choice2.GenerationInfo["PromptCachedTokens"].(int); ok {
		c2 = ct
	}

	// Assert: Request 2 MUST have cached tokens from prefix [system, user1]
	assert.Greater(t, c2, 0, "Request 2 must have cached tokens from prefix [system, user1]")
	assert.Contains(t, choice2.Content, "30", "Response should contain the answer 30")
	assert.NotNil(t, choice2.Reasoning, "Gemini 3 should return reasoning")

	if rr.Recording() {
		t.Log("Waiting 15 seconds for cache update...")
		time.Sleep(15 * time.Second)
	}

	// Request 3: Continue conversation [system, user1, AI1+sig, user2, AI2+sig, user3]
	msgs = append(msgs,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice2.Content, choice2.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now divide that by 5")},
		},
	)

	r3, err := client.GenerateContent(ctx, msgs,
		llms.WithModel("gemini-3-flash-preview"),
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.0),
		llms.WithReasoning(llms.ReasoningMedium, 512),
	)
	require.NoError(t, err)
	require.NotEmpty(t, r3.Choices)

	choice3 := r3.Choices[0]
	c3 := 0
	if ct, ok := choice3.GenerationInfo["PromptCachedTokens"].(int); ok {
		c3 = ct
	}

	// Assert: Request 3 MUST have cached tokens from longer prefix
	assert.Greater(t, c3, 0, "Request 3 must have cached tokens from prefix [system, user1, AI1, user2]")
	assert.Contains(t, choice3.Content, "6", "Response should contain the answer 6")
	assert.NotNil(t, choice3.Reasoning, "Gemini 3 should return reasoning")
}

// TestGoogleAI_ImplicitCaching_ConversationContinuation_Streaming tests implicit caching
// for multi-turn conversations in STREAMING mode with proper signature/reasoning preservation.
//
// This test verifies that:
// 1. Conversation continuation works with streaming
// 2. Reasoning/signature preservation works in streaming mode
// 3. Cached tokens are correctly reported in streaming responses
func TestGoogleAI_ImplicitCaching_ConversationContinuation_Streaming(t *testing.T) { //nolint:funlen
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)
	rr.ScrubReq(httprr.JsonCompactScrubBody)
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	ctx := t.Context()
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithRest(),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	// Marker: change to "v2", "v3" etc. if re-recording
	marker := "ConvStreamTest-v1 "
	largeContext := marker + strings.Repeat("Go is a compiled language with garbage collection. ", 500)

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
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("What is 8 + 7?")}},
	}

	var s1, r1Content strings.Builder
	resp1, err := client.GenerateContent(ctx, msgs,
		llms.WithModel("gemini-3-flash-preview"),
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.0),
		llms.WithReasoning(llms.ReasoningMedium, 512),
		streamFunc(&s1, &r1Content),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.Choices)

	choice1 := resp1.Choices[0]
	c1 := 0
	if ct, ok := choice1.GenerationInfo["PromptCachedTokens"].(int); ok {
		c1 = ct
	}
	// Request 1: first request, no cache expected
	assert.Equal(t, 0, c1, "Request 1 (streaming) should have 0 cached tokens")
	assert.Contains(t, choice1.Content, "15", "Response should contain the answer 15")
	assert.NotNil(t, choice1.Reasoning, "Gemini 3 streaming should return reasoning")
	assert.NotEmpty(t, s1.String(), "Streamed content should not be empty")

	if rr.Recording() {
		t.Log("Waiting 15 seconds for Google to establish implicit cache...")
		time.Sleep(15 * time.Second)
	}

	// Request 2: Continue conversation [system, user1, AI1+signature, user2]
	msgs = append(msgs,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice1.Content, choice1.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now multiply that by 3")},
		},
	)

	var s2, r2Content strings.Builder
	resp2, err := client.GenerateContent(ctx, msgs,
		llms.WithModel("gemini-3-flash-preview"),
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.0),
		llms.WithReasoning(llms.ReasoningMedium, 512),
		streamFunc(&s2, &r2Content),
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
	assert.Contains(t, choice2.Content, "45", "Response should contain the answer 45")
	assert.NotNil(t, choice2.Reasoning, "Gemini 3 streaming should return reasoning")
	assert.NotEmpty(t, s2.String(), "Streamed content should not be empty")

	if rr.Recording() {
		t.Log("Waiting 15 seconds for cache update...")
		time.Sleep(15 * time.Second)
	}

	// Request 3: Continue conversation [system, user1, AI1+sig, user2, AI2+sig, user3]
	msgs = append(msgs,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice2.Content, choice2.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now divide that by 5")},
		},
	)

	var s3, r3Content strings.Builder
	resp3, err := client.GenerateContent(ctx, msgs,
		llms.WithModel("gemini-3-flash-preview"),
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.0),
		llms.WithReasoning(llms.ReasoningMedium, 512),
		streamFunc(&s3, &r3Content),
	)
	require.NoError(t, err)
	require.NotEmpty(t, resp3.Choices)

	choice3 := resp3.Choices[0]
	c3 := 0
	if ct, ok := choice3.GenerationInfo["PromptCachedTokens"].(int); ok {
		c3 = ct
	}

	// Assert: Request 3 MUST have cached tokens from longer prefix in streaming mode
	assert.Greater(t, c3, 0, "Request 3 (streaming) must have cached tokens from longer prefix")
	assert.Contains(t, choice3.Content, "9", "Response should contain the answer 9")
	assert.NotNil(t, choice3.Reasoning, "Gemini 3 streaming should return reasoning")
	assert.NotEmpty(t, s3.String(), "Streamed content should not be empty")
}
