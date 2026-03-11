package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestWithMaxCompletionTokens(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that WithMaxCompletionTokens sets MaxTokens
	WithMaxCompletionTokens(100)(opts)
	if opts.GetMaxTokens() != 100 {
		t.Errorf("expected MaxTokens=100, got %d", opts.GetMaxTokens())
	}

	// Test that it can be overridden
	WithMaxCompletionTokens(200)(opts)
	if opts.GetMaxTokens() != 200 {
		t.Errorf("expected MaxTokens=200, got %d", opts.GetMaxTokens())
	}

	// Test with zero value
	WithMaxCompletionTokens(0)(opts)
	if opts.GetMaxTokens() != 0 {
		t.Errorf("expected MaxTokens=0, got %d", opts.GetMaxTokens())
	}
}

func TestOptionsCompatibility(t *testing.T) {
	opts := &llms.CallOptions{}

	// Test that both llms.WithMaxTokens and WithMaxCompletionTokens
	// set the same field for compatibility
	llms.WithMaxTokens(150)(opts)
	if opts.GetMaxTokens() != 150 {
		t.Errorf("expected MaxTokens=150, got %d", opts.GetMaxTokens())
	}

	opts2 := &llms.CallOptions{}
	WithMaxCompletionTokens(150)(opts2)
	if opts2.GetMaxTokens() != 150 {
		t.Errorf("expected MaxTokens=150, got %d", opts2.GetMaxTokens())
	}

	// They should be equivalent
	if opts.GetMaxTokens() != opts2.GetMaxTokens() {
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
	if opts2.GetMaxTokens() != 200 {
		t.Errorf("expected MaxTokens=200, got %d", opts2.GetMaxTokens())
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
	require.NotNil(t, result, "expected non-nil result")
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
	require.NotNil(t, result2, "expected non-nil result")
	require.NotNil(t, result2.UserLocation, "expected UserLocation to be set")
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

func TestWithExtraBody(t *testing.T) {
	t.Run("sets extra body in metadata", func(t *testing.T) {
		opts := &llms.CallOptions{}
		extraBody := map[string]any{
			"enable_thinking": false,
			"custom_param":    "value",
		}

		WithExtraBody(extraBody)(opts)

		require.NotNil(t, opts.Metadata)
		stored, ok := opts.Metadata["openai:extra_body"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, false, stored["enable_thinking"])
		assert.Equal(t, "value", stored["custom_param"])
	})

	t.Run("nil extra body is handled", func(t *testing.T) {
		opts := &llms.CallOptions{}
		WithExtraBody(nil)(opts)
		require.NotNil(t, opts.Metadata)
	})

	t.Run("empty extra body is handled", func(t *testing.T) {
		opts := &llms.CallOptions{}
		WithExtraBody(map[string]any{})(opts)
		require.NotNil(t, opts.Metadata)
		stored, ok := opts.Metadata["openai:extra_body"].(map[string]any)
		require.True(t, ok)
		assert.Empty(t, stored)
	})
}

func TestGetExtraBody(t *testing.T) {
	t.Run("returns nil when metadata is nil", func(t *testing.T) {
		opts := &llms.CallOptions{}
		result := getExtraBody(opts)
		assert.Nil(t, result)
	})

	t.Run("returns nil when extra_body not present", func(t *testing.T) {
		opts := &llms.CallOptions{
			Metadata: map[string]any{"other_key": "value"},
		}
		result := getExtraBody(opts)
		assert.Nil(t, result)
	})

	t.Run("returns extra body when present", func(t *testing.T) {
		extraBody := map[string]any{"key": "value"}
		opts := &llms.CallOptions{
			Metadata: map[string]any{
				"openai:extra_body": extraBody,
			},
		}
		result := getExtraBody(opts)
		require.NotNil(t, result)
		assert.Equal(t, "value", result["key"])
	})

	t.Run("returns nil for wrong type", func(t *testing.T) {
		opts := &llms.CallOptions{
			Metadata: map[string]any{
				"openai:extra_body": "not a map",
			},
		}
		result := getExtraBody(opts)
		assert.Nil(t, result)
	})
}
