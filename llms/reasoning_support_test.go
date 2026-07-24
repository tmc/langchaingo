package llms

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestReasoningSupportFor(t *testing.T) { //nolint:funlen // table-driven test
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

	t.Run("Sonnet 5 disable hint is provider-aware", func(t *testing.T) {
		// Disablable on the Anthropic API, not disablable on Bedrock (always on).
		anthropic := ReasoningSupportFor("claude-sonnet-5", reasoning.ProviderAnthropic)
		eq(t, "CannotDisable(Anthropic)", anthropic.CannotDisable, false)
		bedrock := ReasoningSupportFor("us.anthropic.claude-sonnet-5", reasoning.ProviderBedrock)
		eq(t, "CannotDisable(Bedrock)", bedrock.CannotDisable, true)
	})

	t.Run("Opus 5 defaults on but is disablable, like Sonnet 5 unlike Opus 4.8", func(t *testing.T) {
		s := ReasoningSupportFor("claude-opus-5", reasoning.ProviderAnthropic)
		eq(t, "Known", s.Known, true)
		eq(t, "CannotDisable", s.CannotDisable, false)
		eq(t, "RejectsSampling", s.RejectsSampling, true)
		if s.DefaultOn == nil || !*s.DefaultOn {
			t.Errorf("Opus 5 DefaultOn = %v, want true", s.DefaultOn)
		}
		// Provider-aware, same rule as Sonnet 5: not disablable on Bedrock.
		bedrock := ReasoningSupportFor("us.anthropic.claude-opus-5", reasoning.ProviderBedrock)
		eq(t, "CannotDisable(Bedrock)", bedrock.CannotDisable, true)
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

	t.Run("OpenAI effort set is model-dependent", func(t *testing.T) {
		pro := ReasoningSupportFor("gpt-5-pro", reasoning.ProviderOpenAI)
		eq(t, "gpt-5-pro Efforts len", len(pro.Efforts), 1)
		if len(pro.Efforts) == 1 {
			eq(t, "gpt-5-pro only high", pro.Efforts[0], ReasoningHigh)
		}
		eq(t, "gpt-5-pro CannotDisable", pro.CannotDisable, true)

		mini := ReasoningSupportFor("gpt-5.4-mini", reasoning.ProviderOpenAI)
		hasXHigh := false
		hasMax := false
		for _, e := range mini.Efforts {
			hasXHigh = hasXHigh || e == ReasoningXHigh
			hasMax = hasMax || e == ReasoningMax
		}
		if !hasXHigh {
			t.Errorf("gpt-5.4-mini efforts must include xhigh, got %v", mini.Efforts)
		}
		if hasMax {
			t.Errorf("gpt-5.4-mini efforts must not include max, got %v", mini.Efforts)
		}
		eq(t, "gpt-5.4-mini CannotDisable", mini.CannotDisable, false)
	})

	t.Run("Gemini 2.5 defaults on", func(t *testing.T) {
		s := ReasoningSupportFor("gemini-2.5-flash", reasoning.ProviderGoogleAI)
		eq(t, "Supported", s.Supported, true)
		if s.DefaultOn == nil || !*s.DefaultOn {
			t.Errorf("Gemini 2.5 DefaultOn = %v, want true", s.DefaultOn)
		}
	})

	t.Run("Gemini 2.5 disable hint is model-dependent", func(t *testing.T) {
		flash := ReasoningSupportFor("gemini-2.5-flash", reasoning.ProviderGoogleAI)
		eq(t, "2.5 Flash CannotDisable", flash.CannotDisable, false)
		pro := ReasoningSupportFor("gemini-2.5-pro", reasoning.ProviderGoogleAI)
		eq(t, "2.5 Pro CannotDisable", pro.CannotDisable, true)
	})

	t.Run("production Gemini 3.x / Gemma 4 are Known and classified", func(t *testing.T) {
		// Mirrors pentagi backend/pkg/providers/gemini/models.yml.
		cases := []struct {
			model         string
			cannotDisable bool
		}{
			{"gemini-3.5-flash", true},
			{"gemini-3.1-pro-preview", true},
			{"gemini-3.1-pro-preview-customtools", true},
			{"gemini-3.1-flash-lite", true},
			{"gemini-2.5-pro", true},
			{"gemini-2.5-flash", false},
			{"gemini-2.5-flash-lite", false},
			{"gemma-4-31b-it", false},
			{"gemma-4-26b-a4b-it", false},
		}
		for _, tc := range cases {
			s := ReasoningSupportFor(tc.model, reasoning.ProviderGoogleAI)
			eq(t, tc.model+" Supported", s.Supported, true)
			eq(t, tc.model+" Known", s.Known, true)
			eq(t, tc.model+" CannotDisable", s.CannotDisable, tc.cannotDisable)
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
