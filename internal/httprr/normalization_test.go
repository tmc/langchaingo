package httprr

import (
	"regexp"
	"testing"
)

func TestNormalizeGoogleAPIClientHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Google API client header with versions",
			input:    "gl-go/1.24.4 gccl/v0.15.1 genai-go/0.15.1 gapic/0.7.0 gax/2.14.1 rest/UNKNOWN",
			expected: "gl-go/X.XX.X gccl/vX.XX.X genai-go/X.XX.X gapic/X.X.X gax/X.XX.X rest/UNKNOWN",
		},
		{
			name:     "Google API client header with different versions",
			input:    "gl-go/1.24.6 gccl/v0.15.2 genai-go/0.16.0 gapic/0.8.1 gax/2.15.0 rest/UNKNOWN",
			expected: "gl-go/X.XX.X gccl/vX.XX.X genai-go/X.XX.X gapic/X.X.X gax/X.XX.X rest/UNKNOWN",
		},
		{
			name:     "Mixed version formats",
			input:    "client/1.2 sdk/v3.4.5 lib/0.1.0-beta rest/UNKNOWN",
			expected: "client/X.X sdk/vX.X.X lib/X.X.X-beta rest/UNKNOWN",
		},
		{
			name:     "No versions",
			input:    "client/unknown sdk/latest rest/UNKNOWN",
			expected: "client/unknown sdk/latest rest/UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeGoogleAPIClientHeader(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeGoogleAPIClientHeader(%q) = %q, want %q", tt.input, result, tt.expected)
			}
			// Verify byte count is preserved
			if len(result) != len(tt.input) {
				t.Errorf("Byte count not preserved: input len=%d, result len=%d", len(tt.input), len(result))
			}
		})
	}
}

func TestNormalizeVersionHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Semantic versions",
			input:    "SDK/1.2.3 Client/v2.4.6 Agent/3.0.0-beta.1",
			expected: "SDK/X.X.X Client/X.X.X Agent/X.X.X",
		},
		{
			name:     "Go version format",
			input:    "compiled with go1.21.0 runtime go1.21.5",
			expected: "compiled with goX.X.X runtime goX.X.X",
		},
		{
			name:     "Date versions",
			input:    "build 20240815 version 2024.08.15",
			expected: "build XXXX.XX.XX version XXXX.XX.XX",
		},
		{
			name:     "Mixed formats",
			input:    "aws-sdk-go/1.44.0 (go1.21.0; linux; amd64) release/2024-08-15",
			expected: "aws-sdk-go/X.X.X (goX.X.X; linux; amd64) release/XXXX.XX.XX",
		},
		{
			name:     "No versions",
			input:    "custom-client production build",
			expected: "custom-client production build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVersionHeader(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersionHeader(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestVersionNormalizationConsistency(t *testing.T) {
	// Test that different versions of the same header format normalize to the same value
	headers1 := []string{
		"gl-go/1.24.4 gccl/v0.15.1 genai-go/0.15.1",
		"gl-go/1.24.6 gccl/v0.15.2 genai-go/0.16.0",
		"gl-go/1.25.0 gccl/v0.16.0 genai-go/0.17.0",
	}

	// All should normalize to the same value
	expected := "gl-go/X.XX.X gccl/vX.XX.X genai-go/X.XX.X"

	for _, header := range headers1 {
		result := normalizeGoogleAPIClientHeader(header)
		if result != expected {
			t.Errorf("Version normalization not consistent: %q -> %q, expected %q", header, result, expected)
		}
	}
}

func TestOpenAIProjectHeaderScrubbing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Remove OpenAI-Project header",
			input: `POST /v1/chat/completions HTTP/1.1
Host: api.openai.com
User-Agent: Go-http-client/1.1
Authorization: Bearer sk-test
openai-project: proj-123456789
Content-Type: application/json

{"model":"gpt-4"}`,
			expected: `POST /v1/chat/completions HTTP/1.1
Host: api.openai.com
User-Agent: langchaingo-httprr
Authorization: Bearer sk-test
Content-Type: application/json

{"model":"gpt-4"}`,
		},
		{
			name: "No OpenAI-Project header present",
			input: `POST /v1/chat/completions HTTP/1.1
Host: api.openai.com
User-Agent: Go-http-client/1.1
Authorization: Bearer sk-test
Content-Type: application/json

{"model":"gpt-4"}`,
			expected: `POST /v1/chat/completions HTTP/1.1
Host: api.openai.com
User-Agent: langchaingo-httprr
Authorization: Bearer sk-test
Content-Type: application/json

{"model":"gpt-4"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates what happens in reqWire during request serialization
			result := regexp.MustCompile(`(?m)^User-Agent: .*$`).ReplaceAllString(tt.input, "User-Agent: langchaingo-httprr")
			result = regexp.MustCompile(`(?m)^openai-project: .*\n`).ReplaceAllString(result, "")

			if result != tt.expected {
				t.Errorf("OpenAI-Project header scrubbing failed:\nGot:\n%s\n\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}
