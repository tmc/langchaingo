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
		{"claude-sonnet-5", ProviderAnthropic, OffDisableClaude}, // on by default, disablable
		{"claude-opus-4-8", ProviderAnthropic, OffOmit},          // off by default
		{"claude-sonnet-4-5", ProviderBedrock, OffOmit},          // off by default
		{"claude-3-5-haiku", ProviderAnthropic, OffOmit},         // no thinking
		{"gemini-2.5-flash", ProviderGoogleAI, OffZeroBudget},
		{"gpt-5.5", ProviderOpenAI, OffEffortNone},
		{"o3-mini", ProviderOpenAI, OffUnsupported},
		{"o1-preview", ProviderOpenAI, OffUnsupported},
		{"gpt-4.1", ProviderOpenAI, OffOmit},                   // non-reasoning
		{"some-random-model", ProviderGoogleAI, OffZeroBudget}, // optimistic google
	}
	for _, tc := range cases {
		if got := ResolveOff(tc.model, tc.p); got != tc.want {
			t.Errorf("ResolveOff(%q,%d) = %d, want %d", tc.model, tc.p, got, tc.want)
		}
	}
}
