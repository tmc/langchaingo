package llms_test

import (
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms"
)

func TestEphemeralCache(t *testing.T) {
	cache := llms.EphemeralCache()
	
	if cache.Type != "ephemeral" {
		t.Errorf("expected type 'ephemeral', got %q", cache.Type)
	}
	
	if cache.Duration != 5*time.Minute {
		t.Errorf("expected duration 5m, got %v", cache.Duration)
	}
}

func TestEphemeralCacheOneHour(t *testing.T) {
	cache := llms.EphemeralCacheOneHour()
	
	if cache.Type != "ephemeral" {
		t.Errorf("expected type 'ephemeral', got %q", cache.Type)
	}
	
	if cache.Duration != time.Hour {
		t.Errorf("expected duration 1h, got %v", cache.Duration)
	}
}

func TestWithCacheControl(t *testing.T) {
	textContent := llms.TextPart("Hello world")
	cacheControl := llms.EphemeralCache()
	
	cached := llms.WithCacheControl(textContent, cacheControl)
	
	if cached.ContentPart != textContent {
		t.Error("cached content should wrap the original content")
	}
	
	if cached.CacheControl != cacheControl {
		t.Error("cached content should have the provided cache control")
	}
}

func TestWithPromptCaching(t *testing.T) {
	option := llms.WithPromptCaching(true)
	
	var opts llms.CallOptions
	option(&opts)
	
	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}
	
	enabled, ok := opts.Metadata["prompt_caching"].(bool)
	if !ok {
		t.Fatal("prompt_caching should be a bool")
	}
	
	if !enabled {
		t.Error("prompt_caching should be true")
	}
}

func TestWithAnthropicCachingHeaders(t *testing.T) {
	option := llms.WithAnthropicCachingHeaders()
	
	var opts llms.CallOptions
	option(&opts)
	
	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}
	
	headers, ok := opts.Metadata["anthropic:beta_headers"].([]string)
	if !ok {
		t.Fatal("anthropic:beta_headers should be a []string")
	}
	
	if len(headers) != 1 || headers[0] != "prompt-caching-2024-07-31" {
		t.Errorf("expected ['prompt-caching-2024-07-31'], got %v", headers)
	}
}

func TestCachedContentImplementsContentPart(t *testing.T) {
	textContent := llms.TextPart("Hello world")
	cached := llms.WithCacheControl(textContent, llms.EphemeralCache())
	
	// This should compile - CachedContent implements ContentPart
	var part llms.ContentPart = cached
	_ = part
}