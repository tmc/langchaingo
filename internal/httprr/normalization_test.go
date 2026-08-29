package httprr

import (
	"net/http"
	"path/filepath"
	"strings"
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
			expected: "gl-go/X.XX.X gccl/vX.XX.X genai-go/X.XX.X gapic/X.XX.X gax/X.XX.X rest/UNKNOWN",
		},
		{
			name:     "Google API client header with different versions",
			input:    "gl-go/1.24.6 gccl/v0.15.2 genai-go/0.16.0 gapic/0.8.1 gax/2.15.0 rest/UNKNOWN",
			expected: "gl-go/X.XX.X gccl/vX.XX.X genai-go/X.XX.X gapic/X.XX.X gax/X.XX.X rest/UNKNOWN",
		},
		{
			name:     "Mixed version formats",
			input:    "client/1.2 sdk/v3.4.5 lib/0.1.0-beta rest/UNKNOWN",
			expected: "client/X.X sdk/vX.XX.X lib/X.XX.X-beta rest/UNKNOWN",
		},
		{
			name:     "No versions",
			input:    "client/unknown sdk/latest rest/UNKNOWN",
			expected: "client/unknown sdk/latest rest/UNKNOWN",
		},
		{
			name:     "Go toolchain version carries a go prefix",
			input:    "google-genai-sdk/1.29.0 gl-go/go1.26.5",
			expected: "google-genai-sdk/X.XX.X gl-go/goX.XX.X",
		},
		{
			name:     "A newer patch of the same toolchain normalizes the same way",
			input:    "google-genai-sdk/1.29.0 gl-go/go1.26.7",
			expected: "google-genai-sdk/X.XX.X gl-go/goX.XX.X",
		},
		{
			name:     "A two-digit patch normalizes to the same shape as a one-digit one",
			input:    "google-genai-sdk/1.29.0 gl-go/go1.26.10",
			expected: "google-genai-sdk/X.XX.X gl-go/goX.XX.X",
		},
		{
			name:     "A two-digit minor normalizes the same way too",
			input:    "google-genai-sdk/1.29.0 gl-go/go1.100.2",
			expected: "google-genai-sdk/X.XX.X gl-go/goX.XX.X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeGoogleAPIClientHeader(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeGoogleAPIClientHeader(%q) = %q, want %q", tt.input, result, tt.expected)
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
	rr, err := create(filepath.Join(t.TempDir(), "scrub.httprr"), http.DefaultTransport)
	if err != nil {
		t.Fatalf("create() error: %v", err)
	}
	t.Cleanup(func() { _ = rr.Close() })

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions",
		strings.NewReader(`{"model":"gpt-4"}`))
	if err != nil {
		t.Fatalf("NewRequest() error: %v", err)
	}
	req.Header.Set("Authorization", "Bearer sk-test")
	req.Header.Set("openai-project", "proj-123456789")
	req.Header.Set("User-Agent", "Go-http-client/1.1")

	wire, err := rr.reqWire(req)
	if err != nil {
		t.Fatalf("reqWire() error: %v", err)
	}

	if strings.Contains(strings.ToLower(wire), "openai-project") {
		t.Errorf("the project header must not reach the cassette match key:\n%s", wire)
	}
	if strings.Contains(wire, "Go-http-client") {
		t.Errorf("the user agent must be normalized in the match key:\n%s", wire)
	}
	if !strings.Contains(wire, `{"model":"gpt-4"}`) {
		t.Errorf("the body must survive scrubbing:\n%s", wire)
	}
}
