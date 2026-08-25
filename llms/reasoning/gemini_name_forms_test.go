package reasoning

import "testing"

func TestGeminiPredicatesAgreeAcrossNameForms(t *testing.T) {
	t.Parallel()

	for _, group := range [][]string{
		{
			"gemini-2.5-flash",
			"Gemini-2.5-Flash",
			"google/gemini-2.5-flash",
			"models/gemini-2.5-flash",
			"projects/p/locations/us-central1/publishers/google/models/gemini-2.5-flash",
			"projects/claude-team/locations/us-central1/publishers/google/models/gemini-2.5-flash",
			"projects/gemma-1-lab/locations/us-central1/publishers/google/models/gemini-2.5-flash",
			"projects/gemini-3-lab/locations/us-central1/publishers/google/models/gemini-2.5-flash",
		},
		{
			"gemini-2.5-pro",
			"projects/flash-lite-team/locations/us-central1/publishers/google/models/gemini-2.5-pro",
		},
		{
			"gemini-3-pro-preview",
			"projects/p/locations/global/publishers/google/models/gemini-3-pro-preview",
		},
		{
			"gemma-4-27b-it",
			"projects/p/locations/us-central1/publishers/google/models/gemma-4-27b-it",
		},
		{
			"text-bison",
			"projects/gemini-2.5-lab/locations/us-central1/publishers/google/models/text-bison",
		},
	} {
		canonical := group[0]
		want := struct {
			thinking     bool
			level        bool
			canDisable   bool
			offByDefault bool
			nonThinking  bool
			off          OffWire
		}{
			GeminiSupportsThinking(canonical),
			GeminiUsesThinkingLevel(canonical),
			GeminiCanDisable(canonical),
			GeminiThinkingOffByDefault(canonical),
			geminiKnownNonThinking(canonical),
			ResolveOff(canonical, ProviderGoogleAI),
		}

		for _, form := range group[1:] {
			t.Run(form, func(t *testing.T) {
				t.Parallel()
				if got := GeminiSupportsThinking(form); got != want.thinking {
					t.Errorf("GeminiSupportsThinking = %v, want %v (as %s)", got, want.thinking, canonical)
				}
				if got := GeminiUsesThinkingLevel(form); got != want.level {
					t.Errorf("GeminiUsesThinkingLevel = %v, want %v (as %s)", got, want.level, canonical)
				}
				if got := GeminiCanDisable(form); got != want.canDisable {
					t.Errorf("GeminiCanDisable = %v, want %v (as %s)", got, want.canDisable, canonical)
				}
				if got := GeminiThinkingOffByDefault(form); got != want.offByDefault {
					t.Errorf("GeminiThinkingOffByDefault = %v, want %v (as %s)", got, want.offByDefault, canonical)
				}
				if got := geminiKnownNonThinking(form); got != want.nonThinking {
					t.Errorf("geminiKnownNonThinking = %v, want %v (as %s)", got, want.nonThinking, canonical)
				}
				if got := ResolveOff(form, ProviderGoogleAI); got != want.off {
					t.Errorf("ResolveOff = %v, want %v (as %s)", got, want.off, canonical)
				}
			})
		}
	}
}
