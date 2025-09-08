package anthropic_test

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func TestAnthropicPromptCaching(t *testing.T) {
	// This test requires ANTHROPIC_API_KEY environment variable
	if testing.Short() {
		t.Skip("skipping prompt caching test in short mode")
	}

	llm, err := anthropic.New(anthropic.WithModel("claude-3-haiku-20240307"))
	if err != nil {
		t.Skip("failed to create Anthropic LLM, skipping prompt caching test")
	}

	ctx := context.Background()

	// Create a message with cached content
	longContext := "You are an expert software engineer with deep knowledge of Go programming. " +
		"You have extensive experience building distributed systems, microservices, and CLI tools. " +
		"You understand best practices for error handling, testing, and code organization in Go. " +
		"You always write clean, idiomatic Go code following the standard library patterns. " +
		"Please analyze any code I provide and suggest improvements based on Go best practices."

	cachedPart := llms.WithCacheControl(
		llms.TextPart(longContext),
		anthropic.EphemeralCache(),
	)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{cachedPart},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What are the main principles of good Go code?"),
			},
		},
	}

	// Call with Anthropic caching headers
	resp, err := llm.GenerateContent(ctx, messages,
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(100),
	)

	if err != nil {
		// In tests, we expect this to work if API key is available
		// If not available, we skip rather than fail
		t.Skipf("failed to generate content with caching: %v", err)
	}

	if resp == nil || len(resp.Choices) == 0 {
		t.Fatal("expected non-empty response")
	}

	if resp.Choices[0].Content == "" {
		t.Error("expected non-empty content in response")
	}

	// Test with 1-hour cache
	longCachePart := llms.WithCacheControl(
		llms.TextPart(longContext),
		anthropic.EphemeralCacheOneHour(),
	)

	messages2 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{longCachePart},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Give me one Go best practice."),
			},
		},
	}

	resp2, err := llm.GenerateContent(ctx, messages2,
		anthropic.WithPromptCaching(),
		llms.WithMaxTokens(50),
	)

	if err != nil {
		t.Skipf("failed to generate content with 1-hour caching: %v", err)
	}

	if resp2 == nil || len(resp2.Choices) == 0 {
		t.Fatal("expected non-empty response for 1-hour cache test")
	}
}

func TestAnthropicCacheControlInMessages(t *testing.T) {
	// Test that cache control is properly handled in message processing
	// This is a unit test that doesn't require API calls

	cachedText := llms.WithCacheControl(
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

	cached, ok := messages[0].Parts[0].(llms.CachedContent)
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
