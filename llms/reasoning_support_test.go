package llms

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestReasoningSupportFor(t *testing.T) {
	t.Parallel()

	boolp := func(b bool) *bool { return &b }
	eq := func(t *testing.T, name string, got, want any) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}

	t.Run("adaptive-only Claude cannot disable and rejects sampling", func(t *testing.T) {
		s := ReasoningSupportFor("us.anthropic.claude-fable-5-v1:0", reasoning.ProviderBedrock)
		eq(t, "Supported", s.Supported, true)
		eq(t, "Known", s.Known, true)
		eq(t, "CannotDisable", s.CannotDisable, true)
		eq(t, "RejectsSampling", s.RejectsSampling, true)
		if s.DefaultOn == nil || !*s.DefaultOn {
			t.Errorf("Fable 5 DefaultOn = %v, want true", s.DefaultOn)
		}
		if len(s.Efforts) == 0 {
			t.Error("adaptive-only Claude should advertise effort tiers")
		}
	})

	t.Run("adaptive-only off-by-default Claude is disablable", func(t *testing.T) {
		s := ReasoningSupportFor("claude-opus-4-8", reasoning.ProviderAnthropic)
		eq(t, "CannotDisable", s.CannotDisable, false)
		eq(t, "RejectsSampling", s.RejectsSampling, true)
		if s.DefaultOn == nil || *s.DefaultOn {
			t.Errorf("Opus 4.8 DefaultOn = %v, want false", s.DefaultOn)
		}
	})

	t.Run("budget-only Claude allows sampling and disable", func(t *testing.T) {
		s := ReasoningSupportFor("claude-sonnet-4-5", reasoning.ProviderAnthropic)
		eq(t, "Known", s.Known, true)
		eq(t, "CannotDisable", s.CannotDisable, false)
		eq(t, "RejectsSampling", s.RejectsSampling, false)
		if s.DefaultOn == nil || *s.DefaultOn {
			t.Errorf("Sonnet 4.5 DefaultOn = %v, want false", s.DefaultOn)
		}
	})

	t.Run("OpenAI o-series cannot disable", func(t *testing.T) {
		s := ReasoningSupportFor("o3-mini", reasoning.ProviderOpenAI)
		eq(t, "Supported", s.Supported, true)
		eq(t, "Known", s.Known, true)
		eq(t, "CannotDisable", s.CannotDisable, true)
	})

	t.Run("GPT-5.x can disable", func(t *testing.T) {
		s := ReasoningSupportFor("gpt-5.5", reasoning.ProviderOpenAI)
		eq(t, "Supported", s.Supported, true)
		eq(t, "CannotDisable", s.CannotDisable, false)
		if s.DefaultOn != nil {
			t.Errorf("GPT-5.x DefaultOn should be unknown (nil), got %v", *s.DefaultOn)
		}
	})

	t.Run("Gemini 2.5 defaults on", func(t *testing.T) {
		s := ReasoningSupportFor("gemini-2.5-flash", reasoning.ProviderGoogleAI)
		eq(t, "Supported", s.Supported, true)
		if s.DefaultOn == nil || !*s.DefaultOn {
			t.Errorf("Gemini 2.5 DefaultOn = %v, want true", s.DefaultOn)
		}
	})

	t.Run("unknown model is optimistic", func(t *testing.T) {
		s := ReasoningSupportFor("some-future-model", reasoning.ProviderUnknown)
		eq(t, "Known", s.Known, false)
		eq(t, "CannotDisable", s.CannotDisable, false)
	})

	t.Run("registered override wins", func(t *testing.T) {
		RegisterReasoningSupport("acme-proxy-thinker", ReasoningSupport{
			Supported: true, Known: true, DefaultOn: boolp(true),
			Efforts: []ReasoningEffort{ReasoningLow, ReasoningHigh},
		})
		s := ReasoningSupportFor("acme-proxy-thinker-v2", reasoning.ProviderUnknown)
		eq(t, "Known", s.Known, true)
		eq(t, "Supported", s.Supported, true)
		if s.DefaultOn == nil || !*s.DefaultOn {
			t.Errorf("override DefaultOn = %v, want true", s.DefaultOn)
		}
	})
}
