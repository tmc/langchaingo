package llms

import (
	"testing"
	"time"
)

func TestCacheTimingAPI(t *testing.T) {
	tests := []struct {
		name     string
		option   CallOption
		expected map[string]interface{}
	}{
		{
			name:   "basic caching",
			option: WithPromptCaching(true),
			expected: map[string]interface{}{
				"prompt_caching": true,
			},
		},
		{
			name:   "cache duration",
			option: WithCacheDuration(30 * time.Minute),
			expected: map[string]interface{}{
				"prompt_caching": true,
				"prompt_caching_advanced": CacheControlOptions{
					Duration: 30 * time.Minute,
				},
			},
		},
		{
			name:   "cache preset",
			option: WithCachePreset(CacheLong),
			expected: map[string]interface{}{
				"prompt_caching": true,
				"prompt_caching_advanced": CacheControlOptions{
					Preset: CacheLong,
				},
			},
		},
		{
			name: "advanced options",
			option: WithPromptCachingAdvanced(CacheControlOptions{
				Duration: 2 * time.Hour,
				Priority: 80,
				Scope:    "session",
				Tags:     []string{"system-prompt", "v1.2"},
			}),
			expected: map[string]interface{}{
				"prompt_caching": true,
				"prompt_caching_advanced": CacheControlOptions{
					Duration: 2 * time.Hour,
					Priority: 80,
					Scope:    "session",
					Tags:     []string{"system-prompt", "v1.2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &CallOptions{}
			tt.option(opts)

			for key, expectedValue := range tt.expected {
				actualValue, exists := opts.Metadata[key]
				if !exists {
					t.Errorf("Expected metadata key %s not found", key)
					continue
				}

				// Special handling for CacheControlOptions comparison
				if key == "prompt_caching_advanced" {
					expected := expectedValue.(CacheControlOptions)
					actual := actualValue.(CacheControlOptions)

					if expected.Duration != actual.Duration {
						t.Errorf("Duration mismatch: expected %v, got %v", expected.Duration, actual.Duration)
					}
					if expected.Preset != actual.Preset {
						t.Errorf("Preset mismatch: expected %v, got %v", expected.Preset, actual.Preset)
					}
					if expected.Priority != actual.Priority {
						t.Errorf("Priority mismatch: expected %v, got %v", expected.Priority, actual.Priority)
					}
					if expected.Scope != actual.Scope {
						t.Errorf("Scope mismatch: expected %v, got %v", expected.Scope, actual.Scope)
					}
					// Simple tag comparison
					if len(expected.Tags) != len(actual.Tags) {
						t.Errorf("Tags length mismatch: expected %d, got %d", len(expected.Tags), len(actual.Tags))
					}
				} else {
					if actualValue != expectedValue {
						t.Errorf("Metadata %s: expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestResolveCacheDuration(t *testing.T) {
	tests := []struct {
		name        string
		preset      CacheDuration
		providerMax time.Duration
		expected    time.Duration
	}{
		{
			name:     "short preset",
			preset:   CacheShort,
			expected: 5 * time.Minute,
		},
		{
			name:     "medium preset",
			preset:   CacheMedium,
			expected: 1 * time.Hour,
		},
		{
			name:     "long preset",
			preset:   CacheLong,
			expected: 24 * time.Hour,
		},
		{
			name:        "max preset with provider limit",
			preset:      CacheMax,
			providerMax: 12 * time.Hour,
			expected:    12 * time.Hour,
		},
		{
			name:     "max preset without provider limit",
			preset:   CacheMax,
			expected: 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveCacheDuration(tt.preset, tt.providerMax)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}