package reasoning

import "testing"

func TestGeminiSupportsThinking(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
	}{
		{"gemini-2.5-flash", true},
		{"gemini-2.5-pro", true},
		{"gemini-3-flash-preview", true},
		{"gemini-3.1-pro-preview", true},
		{"gemini-3.5-flash", true},
		{"gemma-4-31b-it", true},
		{"google/gemini-2.5-flash", true},
		{"gemini-1.5-pro", false},
		{"gemini-pro", false},
		{"gpt-5", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := GeminiSupportsThinking(tc.model); got != tc.want {
			t.Errorf("GeminiSupportsThinking(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestGeminiCanDisable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
	}{
		{"gemini-2.5-flash", true},
		{"gemini-2.5-flash-lite", true},
		{"gemma-4-31b-it", true},
		{"gemini-2.5-pro", false},
		{"gemini-3-flash-preview", false},
		{"gemini-3.1-pro-preview", false},
		{"gemini-3.5-flash", false},
		{"some-unknown-google-model", true}, // optimistic
	}
	for _, tc := range cases {
		if got := GeminiCanDisable(tc.model); got != tc.want {
			t.Errorf("GeminiCanDisable(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestGeminiUsesThinkingLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
	}{
		{"gemini-3-flash-preview", true},
		{"gemini-3.1-pro-preview", true},
		{"gemini-2.5-flash", false},
		{"gemini-2.5-pro", false},
		{"gemma-4-31b-it", false},
	}
	for _, tc := range cases {
		if got := GeminiUsesThinkingLevel(tc.model); got != tc.want {
			t.Errorf("GeminiUsesThinkingLevel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}
