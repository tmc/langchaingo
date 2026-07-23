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
		{"claude-opus-4-8", ProviderAnthropic, OffOmit},                   // off by default
		{"claude-sonnet-4-5", ProviderBedrock, OffOmit},                   // off by default
		{"claude-3-5-haiku", ProviderAnthropic, OffOmit},                  // no thinking
		{"gemini-2.5-flash", ProviderGoogleAI, OffZeroBudget},
		{"gemini-2.5-flash-lite", ProviderGoogleAI, OffZeroBudget},
		{"gemini-2.5-pro", ProviderGoogleAI, OffUnsupported}, // Pro cannot disable thinking
		{"gemini-3.1-pro", ProviderGoogleAI, OffUnsupported}, // 3.x has no full off
		{"gemini-3.1-pro-preview", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.1-pro-preview-customtools", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.1-flash-lite", ProviderGoogleAI, OffUnsupported},
		{"gemini-3.5-flash", ProviderGoogleAI, OffUnsupported},
		{"gemini-3-flash", ProviderGoogleAI, OffUnsupported},
		{"gemma-4-31b-it", ProviderGoogleAI, OffZeroBudget},
		{"gemma-4-26b-a4b-it", ProviderGoogleAI, OffZeroBudget},
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
