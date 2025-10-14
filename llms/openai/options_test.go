package openai

import (
	"testing"

	"github.com/vendasta/langchaingo/llms"
)

func TestWithMaxCompletionTokens(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that WithMaxCompletionTokens sets MaxTokens
	WithMaxCompletionTokens(100)(opts)
	if opts.MaxTokens != 100 {
		t.Errorf("expected MaxTokens=100, got %d", opts.MaxTokens)
	}

	// Test that it can be overridden
	WithMaxCompletionTokens(200)(opts)
	if opts.MaxTokens != 200 {
		t.Errorf("expected MaxTokens=200, got %d", opts.MaxTokens)
	}

	// Test with zero value
	WithMaxCompletionTokens(0)(opts)
	if opts.MaxTokens != 0 {
		t.Errorf("expected MaxTokens=0, got %d", opts.MaxTokens)
	}
}

func TestOptionsCompatibility(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that both llms.WithMaxTokens and WithMaxCompletionTokens
	// set the same field for compatibility
	llms.WithMaxTokens(150)(opts)
	if opts.MaxTokens != 150 {
		t.Errorf("expected MaxTokens=150, got %d", opts.MaxTokens)
	}

	opts2 := &llms.CallOptions{}
	WithMaxCompletionTokens(150)(opts2)
	if opts2.MaxTokens != 150 {
		t.Errorf("expected MaxTokens=150, got %d", opts2.MaxTokens)
	}

	// They should be equivalent
	if opts.MaxTokens != opts2.MaxTokens {
		t.Errorf("WithMaxTokens and WithMaxCompletionTokens should set the same field")
	}
}

func TestWithLegacyMaxTokensField(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that WithLegacyMaxTokensField sets the metadata flag
	WithLegacyMaxTokensField()(opts)
	if opts.Metadata == nil {
		t.Fatal("expected Metadata to be initialized")
	}
	if v, ok := opts.Metadata["openai:use_legacy_max_tokens"].(bool); !ok || !v {
		t.Error("expected openai:use_legacy_max_tokens to be true")
	}

	// Test combining with WithMaxTokens
	opts2 := &llms.CallOptions{}
	llms.WithMaxTokens(200)(opts2)
	WithLegacyMaxTokensField()(opts2)
	if opts2.MaxTokens != 200 {
		t.Errorf("expected MaxTokens=200, got %d", opts2.MaxTokens)
	}
	if v, ok := opts2.Metadata["openai:use_legacy_max_tokens"].(bool); !ok || !v {
		t.Error("expected openai:use_legacy_max_tokens to be true")
	}
}
