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

func TestGeminiFamilyBoundary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model         string
		nonThinking   bool
		supportsThink bool
		usesLevel     bool
		off           OffWire
	}{
		{"gemini-1.0-pro", true, false, false, OffOmit},
		{"gemini-1.5-pro", true, false, false, OffOmit},
		{"gemini-10-pro", false, false, false, OffZeroBudget},
		{"gemini-15-flash", false, false, false, OffZeroBudget},
		{"gemini-2.0-flash", true, false, false, OffOmit},
		{"gemini-20-flash", false, false, false, OffZeroBudget},
		{"gemma-3-27b", true, false, false, OffOmit},
		{"gemma-30b", false, false, false, OffZeroBudget},
		{"gemini-3-pro", false, true, true, OffUnsupported},
		{"gemini-30-pro", false, false, false, OffZeroBudget},
		{"gemini-2.5-flash", false, true, false, OffZeroBudget},
		{"gemini-25-flash", false, false, false, OffZeroBudget},
	}
	for _, tc := range cases {
		if got := geminiKnownNonThinking(tc.model); got != tc.nonThinking {
			t.Errorf("geminiKnownNonThinking(%q) = %v, want %v", tc.model, got, tc.nonThinking)
		}
		if got := GeminiSupportsThinking(tc.model); got != tc.supportsThink {
			t.Errorf("GeminiSupportsThinking(%q) = %v, want %v", tc.model, got, tc.supportsThink)
		}
		if got := GeminiUsesThinkingLevel(tc.model); got != tc.usesLevel {
			t.Errorf("GeminiUsesThinkingLevel(%q) = %v, want %v", tc.model, got, tc.usesLevel)
		}
		if got := ResolveOff(tc.model, ProviderGoogleAI); got != tc.off {
			t.Errorf("ResolveOff(%q) = %v, want %v", tc.model, got, tc.off)
		}
	}
}

func TestGeminiFlashLiteDisablesByOmitting(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model      string
		canDisable bool
		off        OffWire
	}{
		{"gemini-2.5-flash-lite", true, OffOmit},
		{"gemini-3.1-flash-lite", true, OffOmit},
		{"gemini-3.5-flash-lite", true, OffOmit},
		{"gemini-2.5-flash", true, OffZeroBudget},
		{"gemini-2.5-pro", false, OffUnsupported},
		{"gemini-3.1-pro-preview", false, OffUnsupported},
		{"gemini-3.6-flash", false, OffUnsupported},
	}
	for _, tc := range cases {
		if got := GeminiCanDisable(tc.model); got != tc.canDisable {
			t.Errorf("GeminiCanDisable(%q) = %v, want %v", tc.model, got, tc.canDisable)
		}
		if got := ResolveOff(tc.model, ProviderGoogleAI); got != tc.off {
			t.Errorf("ResolveOff(%q) = %v, want %v", tc.model, got, tc.off)
		}
	}
}
