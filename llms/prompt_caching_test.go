package llms_test

import (
	"testing"
	"time"

	"github.com/vendasta/langchaingo/llms"
)

func TestWithCacheControl(t *testing.T) {
	textContent := llms.TextPart("Hello world")
	cacheControl := &llms.CacheControl{
		Type:     "test",
		Duration: 10 * time.Minute,
	}

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

func TestCachedContentImplementsContentPart(t *testing.T) {
	textContent := llms.TextPart("Hello world")
	cacheControl := &llms.CacheControl{
		Type:     "test",
		Duration: 10 * time.Minute,
	}
	cached := llms.WithCacheControl(textContent, cacheControl)

	// This should compile - CachedContent implements ContentPart
	var part llms.ContentPart = cached
	_ = part
}
