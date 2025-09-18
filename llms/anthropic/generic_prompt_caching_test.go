package anthropic

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// TestGenericPromptCaching verifies that the generic llms.WithPromptCaching option
// works with the Anthropic client.
func TestGenericPromptCaching(t *testing.T) {
	llm := &LLM{}
	opts := &llms.CallOptions{}

	// Apply generic prompt caching
	llms.WithPromptCaching(true)(opts)

	// Extract headers
	headers, _ := extractThinkingOptions(llm, opts)

	// Check that prompt caching header was added
	hasPromptCaching := false
	for _, h := range headers {
		if h == "prompt-caching-2024-07-31" {
			hasPromptCaching = true
			break
		}
	}

	if !hasPromptCaching {
		t.Error("Generic prompt caching did not add the beta header")
	}
}

// TestBothPromptCachingMethods verifies that both methods of enabling prompt caching work
// and don't duplicate headers.
func TestBothPromptCachingMethods(t *testing.T) {
	llm := &LLM{}
	opts := &llms.CallOptions{}

	// Apply both methods
	WithPromptCaching()(opts)           // Anthropic-specific
	llms.WithPromptCaching(true)(opts)  // Generic

	// Extract headers
	headers, _ := extractThinkingOptions(llm, opts)

	// Count occurrences of prompt caching header
	count := 0
	for _, h := range headers {
		if h == "prompt-caching-2024-07-31" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected exactly one prompt caching header, got %d", count)
	}
}