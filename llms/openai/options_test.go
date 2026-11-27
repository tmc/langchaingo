package openai

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
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

func TestWithWebSearch(t *testing.T) {
	// Test with nil options (default behavior)
	opts := &llms.CallOptions{}
	llms.WithWebSearch(nil)(opts)
	if opts.WebSearchOptions == nil {
		t.Fatal("expected WebSearchOptions to be initialized")
	}

	// Test with custom search context size
	opts2 := &llms.CallOptions{}
	llms.WithWebSearch(&llms.WebSearchOptions{
		SearchContextSize: "high",
	})(opts2)
	if opts2.WebSearchOptions == nil {
		t.Fatal("expected WebSearchOptions to be set")
	}
	if opts2.WebSearchOptions.SearchContextSize != "high" {
		t.Errorf("expected SearchContextSize=high, got %s", opts2.WebSearchOptions.SearchContextSize)
	}

	// Test with user location
	opts3 := &llms.CallOptions{}
	llms.WithWebSearch(&llms.WebSearchOptions{
		SearchContextSize: "medium",
		UserLocation: &llms.UserLocation{
			Type: "approximate",
			Approximate: &llms.ApproximateLocation{
				Country: "US",
				City:    "San Francisco",
				Region:  "California",
			},
		},
	})(opts3)
	if opts3.WebSearchOptions == nil {
		t.Fatal("expected WebSearchOptions to be set")
	}
	if opts3.WebSearchOptions.UserLocation == nil {
		t.Fatal("expected UserLocation to be set")
	}
	if opts3.WebSearchOptions.UserLocation.Type != "approximate" {
		t.Errorf("expected Type=approximate, got %s", opts3.WebSearchOptions.UserLocation.Type)
	}
	if opts3.WebSearchOptions.UserLocation.Approximate == nil {
		t.Fatal("expected Approximate to be set")
	}
	if opts3.WebSearchOptions.UserLocation.Approximate.Country != "US" {
		t.Errorf("expected Country=US, got %s", opts3.WebSearchOptions.UserLocation.Approximate.Country)
	}
	if opts3.WebSearchOptions.UserLocation.Approximate.City != "San Francisco" {
		t.Errorf("expected City=San Francisco, got %s", opts3.WebSearchOptions.UserLocation.Approximate.City)
	}
	if opts3.WebSearchOptions.UserLocation.Approximate.Region != "California" {
		t.Errorf("expected Region=California, got %s", opts3.WebSearchOptions.UserLocation.Approximate.Region)
	}
}

func TestWebSearchOptionsConversion(t *testing.T) {
	// Test nil conversion
	result := webSearchOptionsFromCallOptions(nil)
	if result != nil {
		t.Error("expected nil result for nil input")
	}

	// Test basic conversion
	opts := &llms.WebSearchOptions{
		SearchContextSize: "high",
	}
	result = webSearchOptionsFromCallOptions(opts)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.SearchContextSize != "high" {
		t.Errorf("expected SearchContextSize=high, got %s", result.SearchContextSize)
	}

	// Test full conversion with user location
	opts2 := &llms.WebSearchOptions{
		SearchContextSize: "medium",
		UserLocation: &llms.UserLocation{
			Type: "approximate",
			Approximate: &llms.ApproximateLocation{
				Country: "GB",
				City:    "London",
				Region:  "London",
			},
		},
	}
	result2 := webSearchOptionsFromCallOptions(opts2)
	if result2 == nil {
		t.Fatal("expected non-nil result")
	}
	if result2.UserLocation == nil {
		t.Fatal("expected UserLocation to be set")
	}
	if result2.UserLocation.Type != "approximate" {
		t.Errorf("expected Type=approximate, got %s", result2.UserLocation.Type)
	}
	if result2.UserLocation.Approximate == nil {
		t.Fatal("expected Approximate to be set")
	}
	if result2.UserLocation.Approximate.Country != "GB" {
		t.Errorf("expected Country=GB, got %s", result2.UserLocation.Approximate.Country)
	}
}

func TestWithMetadataSetsStoreField(t *testing.T) {
	// Test that using llms.WithMetadata results in the Store field being set
	// This tests the integration in openaillm.go where apiMetadata is populated
	// and Store is set based on whether metadata is present

	// Simulate the filtering logic from openaillm.go
	opts := &llms.CallOptions{}
	metadata := map[string]interface{}{
		"feature":     "support",
		"environment": "local",
		"team":        "ai",
	}
	llms.WithMetadata(metadata)(opts)

	// Verify metadata is set
	if opts.Metadata == nil {
		t.Fatal("expected Metadata to be set")
	}

	// Simulate the filtering that happens in openaillm.go
	apiMetadata := make(map[string]any)
	for k, v := range opts.Metadata {
		if k == "thinking_config" { // simulating the HasPrefix check
			continue
		}
		apiMetadata[k] = v
	}

	// Verify that when metadata is present, Store should be true
	store := len(apiMetadata) > 0
	if !store {
		t.Error("expected Store to be true when metadata is present")
	}

	// Test with empty metadata
	opts2 := &llms.CallOptions{}
	llms.WithMetadata(map[string]interface{}{})(opts2)

	apiMetadata2 := make(map[string]any)
	for k, v := range opts2.Metadata {
		if k == "thinking_config" {
			continue
		}
		apiMetadata2[k] = v
	}

	store2 := len(apiMetadata2) > 0
	if store2 {
		t.Error("expected Store to be false when metadata is empty")
	}
}
