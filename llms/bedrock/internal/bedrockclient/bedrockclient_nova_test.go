package bedrockclient

import (
	"testing"
)

func TestGetProvider_NovaModels(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		modelID  string
		expected string
	}{
		{
			name:     "standard_nova_lite",
			modelID:  "amazon.nova-lite-v1:0",
			expected: "nova",
		},
		{
			name:     "inference_profile_nova_lite",
			modelID:  "us.amazon.nova-lite-v1:0",
			expected: "nova",
		},
		{
			name:     "standard_nova_pro",
			modelID:  "amazon.nova-pro-v1:0",
			expected: "nova",
		},
		{
			name:     "inference_profile_nova_pro",
			modelID:  "us.amazon.nova-pro-v1:0",
			expected: "nova",
		},
		{
			name:     "region_specific_nova",
			modelID:  "eu-west-1.amazon.nova-lite-v1:0",
			expected: "nova",
		},
		{
			name:     "anthropic_not_nova",
			modelID:  "anthropic.claude-3-sonnet-20240229-v1:0",
			expected: "anthropic",
		},
		{
			name:     "inference_profile_anthropic",
			modelID:  "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
			expected: "anthropic",
		},
		{
			name:     "amazon_titan_not_nova",
			modelID:  "amazon.titan-text-lite-v1",
			expected: "amazon",
		},
		{
			name:     "meta_model",
			modelID:  "meta.llama3-1-405b-instruct-v1:0",
			expected: "meta",
		},
		{
			name:     "cohere_model",
			modelID:  "cohere.command-r-plus-v1:0",
			expected: "cohere",
		},
		{
			name:     "ai21_model",
			modelID:  "ai21.jamba-1-5-large-v1:0",
			expected: "ai21",
		},
		{
			name:     "unknown_model",
			modelID:  "unknown.model",
			expected: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := getProvider(tc.modelID)
			if result != tc.expected {
				t.Errorf("getProvider(%q) = %q, want %q", tc.modelID, result, tc.expected)
			}
		})
	}
}

func TestGetProvider_EdgeCases(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		modelID  string
		expected string
	}{
		{
			name:     "empty_string",
			modelID:  "",
			expected: "",
		},
		{
			name:     "no_dots",
			modelID:  "modelwithoutdots",
			expected: "modelwithoutdots",
		},
		{
			name:     "nova_in_middle",
			modelID:  "custom.nova-based.model",
			expected: "nova",
		},
		{
			name:     "multiple_dots",
			modelID:  "region.provider.model.version",
			expected: "region",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := getProvider(tc.modelID)
			if result != tc.expected {
				t.Errorf("getProvider(%q) = %q, want %q", tc.modelID, result, tc.expected)
			}
		})
	}
}
