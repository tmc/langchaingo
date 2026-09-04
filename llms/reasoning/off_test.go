package reasoning

import "testing"

func TestResolveOff(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		p     Provider
		want  OffWire
	}{
		{"claude-fable-5", ProviderAnthropic, OffUnsupported},
		{"us.anthropic.claude-mythos-5", ProviderBedrock, OffUnsupported},
		{"claude-sonnet-5", ProviderAnthropic, OffDisableClaude},          // on by default, disablable on Anthropic
		{"us.anthropic.claude-sonnet-5", ProviderBedrock, OffUnsupported}, // always on on Bedrock, not disablable
		{"claude-opus-5", ProviderAnthropic, OffDisableClaude},            // on by default, disablable on Anthropic
		{"us.anthropic.claude-opus-5", ProviderBedrock, OffUnsupported},   // always on on Bedrock, not disablable
		{"claude-opus-4-8", ProviderAnthropic, OffOmit},                   // off by default
		{"claude-sonnet-4-5", ProviderBedrock, OffOmit},                   // off by default
		{"claude-3-5-haiku", ProviderAnthropic, OffOmit},                  // no thinking
		{"gemini-2.5-flash", ProviderGoogleAI, OffZeroBudget},
		{"gemini-2.5-flash-lite", ProviderGoogleAI, OffOmit},
		{"gemini-2.5-pro", ProviderGoogleAI, OffUnsupported}, // Pro cannot disable thinking
		{"gemini-3.1-pro", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.1-pro-preview", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.1-pro-preview-customtools", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.1-flash-lite", ProviderGoogleAI, OffOmit},
		{"gemini-3-flash", ProviderGoogleAI, OffZeroBudget},
		{"gemini-3-flash-preview", ProviderGoogleAI, OffZeroBudget},
		{"gemini-3.5-flash", ProviderGoogleAI, OffZeroBudget},
		{"models/gemini-3.5-flash", ProviderGoogleAI, OffZeroBudget},
		{"gemini-3.6-flash", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.7-flash", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.5-pro", ProviderGoogleAI, OffUnsupported},
		{"gemma-4-31b-it", ProviderGoogleAI, OffUnsupported},
		{"gemma-4-26b-a4b-it", ProviderGoogleAI, OffUnsupported},
		// Pre-thinking Gemini/Gemma never think: omit rather than send budget:0.
		{"gemini-2.0-flash", ProviderGoogleAI, OffOmit},
		{"gemini-1.5-pro", ProviderGoogleAI, OffOmit},
		{"gemma-3-27b-it", ProviderGoogleAI, OffOmit},
		{"gpt-5.5", ProviderOpenAI, OffEffortNone},
		{"o3-mini", ProviderOpenAI, OffUnsupported},
		{"o1-preview", ProviderOpenAI, OffUnsupported},
		{"gpt-5-pro", ProviderOpenAI, OffUnsupported},          // accepts only high, cannot disable
		{"gpt-4.1", ProviderOpenAI, OffOmit},                   // non-reasoning
		{"some-random-model", ProviderGoogleAI, OffZeroBudget}, // optimistic google
	}
	for _, tc := range cases {
		if got := ResolveOff(tc.model, tc.p); got != tc.want {
			t.Errorf("ResolveOff(%q,%d) = %d, want %d", tc.model, tc.p, got, tc.want)
		}
	}
}

func TestVendorsThatDisableWithAThinkingObject(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"glm-4.5", "glm-4.6", "glm-4.7", "glm-5", "glm-5-turbo", "glm-5.1", "minimax-m3",
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffDisableThinkingObject {
			t.Errorf("ResolveOff(%q) = %v, want the thinking object: the effort token is either "+
				"ignored by the vendor or cut before it", model, got)
		}
	}

	for model, want := range map[string]OffWire{
		"glm-5.2":     OffEffortNone,
		"glm-5.3":     OffUnsupported,
		"minimax-m2":  OffUnsupported,
		"kimi-k3":     OffEffortNone,
		"deepseek-v4": OffEffortNone,
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != want {
			t.Errorf("ResolveOff(%q) = %v, want %v", model, got, want)
		}
	}
}
