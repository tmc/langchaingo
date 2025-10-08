package anthropic_test

import (
	"testing"
	"time"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/anthropic"
)

func TestEphemeralCache(t *testing.T) {
	cache := anthropic.EphemeralCache()

	if cache.Type != "ephemeral" {
		t.Errorf("expected type 'ephemeral', got %q", cache.Type)
	}

	if cache.Duration != 5*time.Minute {
		t.Errorf("expected duration 5m, got %v", cache.Duration)
	}
}

func TestEphemeralCacheOneHour(t *testing.T) {
	cache := anthropic.EphemeralCacheOneHour()

	if cache.Type != "ephemeral" {
		t.Errorf("expected type 'ephemeral', got %q", cache.Type)
	}

	if cache.Duration != time.Hour {
		t.Errorf("expected duration 1h, got %v", cache.Duration)
	}
}

func TestWithPromptCaching(t *testing.T) {
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

	if len(headers) != 1 || headers[0] != "prompt-caching-2024-07-31" {
		t.Errorf("expected ['prompt-caching-2024-07-31'], got %v", headers)
	}
}

func TestWithExtendedOutput(t *testing.T) {
	option := anthropic.WithExtendedOutput()

	var opts llms.CallOptions
	option(&opts)

	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}

	headers, ok := opts.Metadata["anthropic:beta_headers"].([]string)
	if !ok {
		t.Fatal("anthropic:beta_headers should be a []string")
	}

	if len(headers) != 1 || headers[0] != "output-128k-2025-02-19" {
		t.Errorf("expected ['output-128k-2025-02-19'], got %v", headers)
	}
}

func TestWithInterleavedThinking(t *testing.T) {
	option := anthropic.WithInterleavedThinking()

	var opts llms.CallOptions
	option(&opts)

	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}

	headers, ok := opts.Metadata["anthropic:beta_headers"].([]string)
	if !ok {
		t.Fatal("anthropic:beta_headers should be a []string")
	}

	if len(headers) != 1 || headers[0] != "interleaved-thinking-2025-05-14" {
		t.Errorf("expected ['interleaved-thinking-2025-05-14'], got %v", headers)
	}
}

func TestWithBetaHeader(t *testing.T) {
	option := anthropic.WithBetaHeader("custom-feature-2025-01-01")

	var opts llms.CallOptions
	option(&opts)

	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}

	headers, ok := opts.Metadata["anthropic:beta_headers"].([]string)
	if !ok {
		t.Fatal("anthropic:beta_headers should be a []string")
	}

	if len(headers) != 1 || headers[0] != "custom-feature-2025-01-01" {
		t.Errorf("expected ['custom-feature-2025-01-01'], got %v", headers)
	}
}

func TestMultipleBetaHeaders(t *testing.T) {
	var opts llms.CallOptions

	// Apply multiple options
	anthropic.WithPromptCaching()(&opts)
	anthropic.WithExtendedOutput()(&opts)

	headers, ok := opts.Metadata["anthropic:beta_headers"].([]string)
	if !ok {
		t.Fatal("anthropic:beta_headers should be a []string")
	}

	if len(headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(headers))
	}

	expectedHeaders := map[string]bool{
		"prompt-caching-2024-07-31": true,
		"output-128k-2025-02-19":    true,
	}

	for _, h := range headers {
		if !expectedHeaders[h] {
			t.Errorf("unexpected header: %s", h)
		}
	}
}
