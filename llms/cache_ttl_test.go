package llms

import (
	"testing"
	"time"
)

func TestCacheTTL(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		expected map[string]interface{}
	}{
		{
			name: "5 minute TTL",
			ttl:  5 * time.Minute,
			expected: map[string]interface{}{
				"prompt_caching": true,
				"cache_ttl":      5 * time.Minute,
			},
		},
		{
			name: "1 hour TTL",
			ttl:  1 * time.Hour,
			expected: map[string]interface{}{
				"prompt_caching": true,
				"cache_ttl":      1 * time.Hour,
			},
		},
		{
			name: "custom TTL",
			ttl:  30 * time.Minute,
			expected: map[string]interface{}{
				"prompt_caching": true,
				"cache_ttl":      30 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &CallOptions{}
			WithCacheTTL(tt.ttl)(opts)

			for key, expectedValue := range tt.expected {
				actualValue, exists := opts.Metadata[key]
				if !exists {
					t.Errorf("Expected metadata key %s not found", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Metadata %s: expected %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestBasicPromptCaching(t *testing.T) {
	opts := &CallOptions{}
	WithPromptCaching(true)(opts)

	if enabled, ok := opts.Metadata["prompt_caching"].(bool); !ok || !enabled {
		t.Error("Expected prompt_caching to be true")
	}
}