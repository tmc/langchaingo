package llms

import (
	"slices"
	"testing"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestReasoningSupportForNewerGeneration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model    string
		provider reasoning.Provider
		want     bool
	}{
		{"gpt-6", reasoning.ProviderOpenAI, true},
		{"claude-opus-6", reasoning.ProviderAnthropic, true},
		{"gemini-4-pro", reasoning.ProviderGoogleAI, true},
		{"gpt-4o", reasoning.ProviderOpenAI, false},
		{"claude-3-5-sonnet-latest", reasoning.ProviderAnthropic, false},
	}
	for _, tc := range cases {
		s := ReasoningSupportFor(tc.model, tc.provider)
		if s.Supported != tc.want {
			t.Errorf("ReasoningSupportFor(%q).Supported = %v, want %v", tc.model, s.Supported, tc.want)
		}
		if s.Known {
			t.Errorf("ReasoningSupportFor(%q).Known = true, want false: the tiers are a guess, not a fact", tc.model)
		}
	}
}

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
		gateway := ReasoningSupportFor("anthropic/claude-sonnet-5", reasoning.ProviderOpenAI)
		eq(t, "CannotDisable(OpenAI gateway)", gateway.CannotDisable, true)
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

	t.Run("an unclassified OpenAI model names no effort tiers", func(t *testing.T) {
		s := ReasoningSupportFor("gpt-5.7", reasoning.ProviderOpenAI)
		eq(t, "Supported", s.Supported, true)
		if s.Efforts != nil {
			t.Errorf("Efforts = %v, want nil: the tiers of gpt-5.7 are not classified", s.Efforts)
		}
	})

	t.Run("a classified OpenAI generation names its effort tiers", func(t *testing.T) {
		s := ReasoningSupportFor("gpt-5.5", reasoning.ProviderOpenAI)
		want := []ReasoningEffort{"low", "medium", "high", "xhigh"}
		if !slices.Equal(s.Efforts, want) {
			t.Errorf("Efforts = %v, want %v", s.Efforts, want)
		}
	})

	t.Run("mechanism follows the Claude generation", func(t *testing.T) {
		cases := []struct {
			model string
			want  ReasoningMechanism
		}{
			{"claude-sonnet-5", ReasoningMechanismAdaptive},
			{"us.anthropic.claude-opus-4-8", ReasoningMechanismAdaptive},
			{"claude-opus-4-6", ReasoningMechanismAdaptiveAndBudget},
			{"claude-sonnet-4-5", ReasoningMechanismBudget},
			{"claude-haiku-4-5", ReasoningMechanismBudget},
		}
		for _, tc := range cases {
			got := ReasoningSupportFor(tc.model, reasoning.ProviderAnthropic).Mechanism
			if got != tc.want {
				t.Errorf("Mechanism(%q) = %v, want %v", tc.model, got, tc.want)
			}
		}
		eq(t, "gpt-5.5 Mechanism",
			ReasoningSupportFor("gpt-5.5", reasoning.ProviderOpenAI).Mechanism, ReasoningMechanismAdaptive)
		eq(t, "unclassified model Mechanism",
			ReasoningSupportFor("gpt-5.3", reasoning.ProviderOpenAI).Mechanism, ReasoningMechanismUnknown)
	})

	t.Run("a model outside the capability table is not reported as classified", func(t *testing.T) {
		classified := ReasoningSupportFor("gpt-5.5", reasoning.ProviderOpenAI)
		eq(t, "gpt-5.5 Known", classified.Known, true)
		eq(t, "gpt-5.5 Efforts present", len(classified.Efforts) > 0, true)

		unclassified := ReasoningSupportFor("gpt-5.3", reasoning.ProviderOpenAI)
		eq(t, "gpt-5.3 Supported", unclassified.Supported, true)
		eq(t, "gpt-5.3 Known", unclassified.Known, false)
		eq(t, "gpt-5.3 Efforts empty", len(unclassified.Efforts), 0)
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

	t.Run("Gemini Flash-Lite is off until asked", func(t *testing.T) {
		for _, model := range []string{"gemini-2.5-flash-lite", "gemini-3.1-flash-lite", "gemini-3.5-flash-lite"} {
			s := ReasoningSupportFor(model, reasoning.ProviderGoogleAI)
			eq(t, model+" Supported", s.Supported, true)
			eq(t, model+" CannotDisable", s.CannotDisable, false)
			if s.DefaultOn == nil || *s.DefaultOn {
				t.Errorf("%s DefaultOn = %v, want false", model, s.DefaultOn)
			}
		}
		flash := ReasoningSupportFor("gemini-3.6-flash", reasoning.ProviderGoogleAI)
		if flash.DefaultOn == nil || !*flash.DefaultOn {
			t.Errorf("gemini-3.6-flash DefaultOn = %v, want true", flash.DefaultOn)
		}
		eq(t, "gemini-3.6-flash CannotDisable", flash.CannotDisable, true)
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
			{"gemini-3.5-flash", false},
			{"gemini-3.1-pro-preview", true},
			{"gemini-3.1-pro-preview-customtools", true},
			{"gemini-3.1-flash-lite", false},
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

func TestOpenAIHintSurvivesTheTransportLabel(t *testing.T) {
	t.Parallel()

	labels := []reasoning.Provider{
		reasoning.ProviderUnknown,
		reasoning.ProviderGoogleAI,
		reasoning.ProviderBedrock,
		reasoning.ProviderAnthropic,
	}

	for _, model := range []string{"gpt-5-pro", "gpt-5", "gpt-5-mini", "o3", "o4-mini", "gpt-5.1", "gpt-5.2"} {
		want := ReasoningSupportFor(model, reasoning.ProviderOpenAI)
		if !want.Known || len(want.Efforts) == 0 {
			t.Fatalf("%s: the OpenAI label must know this model, got %+v", model, want)
		}
		for _, p := range labels {
			got := ReasoningSupportFor(model, p)
			if !got.Known {
				t.Errorf("%s behind label %v: Known = false, want the same hint as the OpenAI label", model, p)
			}
			if !slices.Equal(got.Efforts, want.Efforts) {
				t.Errorf("%s behind label %v: Efforts = %v, want %v", model, p, got.Efforts, want.Efforts)
			}
			if got.CannotDisable != want.CannotDisable {
				t.Errorf("%s behind label %v: CannotDisable = %v, want %v", model, p, got.CannotDisable, want.CannotDisable)
			}
		}
	}
}

func TestNonOpenAIModelsKeepTheirOwnBranch(t *testing.T) {
	t.Parallel()

	gemini := ReasoningSupportFor("gemini-2.5-flash", reasoning.ProviderGoogleAI)
	if gemini.DefaultOn == nil {
		t.Fatal("gemini on its own provider must still report DefaultOn from the Google branch")
	}
	if len(gemini.Efforts) != 0 {
		t.Fatalf("gemini has no effort ladder, got %v", gemini.Efforts)
	}

	for _, model := range []string{"grok-4", "glm-5", "gemini-2.5-flash"} {
		if got := ReasoningSupportFor(model, reasoning.ProviderUnknown); got.Known {
			t.Errorf("%s behind an unknown label must stay unclassified, got %+v", model, got)
		}
	}
}

func TestCannotDisableFollowsTheResolverOnEveryProvider(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		p     reasoning.Provider
		want  bool
	}{
		{"claude-sonnet-5", reasoning.ProviderAnthropic, false},
		{"claude-sonnet-5", reasoning.ProviderBedrock, true},
		{"claude-sonnet-5", reasoning.ProviderUnknown, false},
		{"claude-fable-5", reasoning.ProviderAnthropic, true},
		{"claude-fable-5", reasoning.ProviderBedrock, true},
		{"claude-fable-5", reasoning.ProviderUnknown, true},
		{"claude-opus-4-6", reasoning.ProviderUnknown, false},
	} {
		if got := ReasoningSupportFor(tc.model, tc.p).CannotDisable; got != tc.want {
			t.Errorf("CannotDisable(%q, provider=%d) = %v, want %v", tc.model, tc.p, got, tc.want)
		}
	}
}

func TestGoogleHintNamesItsMechanism(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  ReasoningMechanism
	}{
		{"gemini-3-pro-preview", ReasoningMechanismAdaptive},
		{"gemini-2.5-flash", ReasoningMechanismBudget},
		{"gemini-2.5-pro", ReasoningMechanismBudget},
	} {
		got := ReasoningSupportFor(tc.model, reasoning.ProviderGoogleAI)
		if !got.Known {
			t.Fatalf("%q is a classified Google thinking model", tc.model)
		}
		if got.Mechanism != tc.want {
			t.Errorf("ReasoningSupportFor(%q).Mechanism = %v, want %v", tc.model, got.Mechanism, tc.want)
		}
	}
}

func TestABudgetModelWithNoEffortTierIsKnownAndEmpty(t *testing.T) {
	t.Parallel()

	support := ReasoningSupportFor("claude-sonnet-4-5", reasoning.ProviderAnthropic)

	if !support.Known {
		t.Fatal("the model is classified, so Known must stay true")
	}
	if len(support.Efforts) != 0 {
		t.Fatalf("a model that takes no effort must offer none, got %v", support.Efforts)
	}
	if support.Mechanism != ReasoningMechanismBudget {
		t.Fatalf("Mechanism must tell the caller what to offer instead, got %v", support.Mechanism)
	}
}

func TestAdvertisedEffortsFollowTheWireOfEachProvider(t *testing.T) {
	t.Parallel()

	const opus45 = "us.anthropic.claude-opus-4-5-20251101-v1:0"

	for _, tc := range []struct {
		provider reasoning.Provider
		want     bool
	}{
		{reasoning.ProviderAnthropic, true},
		{reasoning.ProviderBedrock, false},
	} {
		sends := reasoning.ClaudeSupportsEffortWithBudget(opus45, tc.provider)
		if sends != tc.want {
			t.Fatalf("wire assumption changed: ClaudeSupportsEffortWithBudget(%v) = %v", tc.provider, sends)
		}

		efforts := ReasoningSupportFor(opus45, tc.provider).Efforts
		if sends && len(efforts) == 0 {
			t.Errorf("%v sends an effort with the budget, so the hint must list some", tc.provider)
		}
		if !sends && len(efforts) > 0 {
			t.Errorf("%v never sends an effort for this model, yet the hint advertises %v", tc.provider, efforts)
		}
	}
}

func TestTheHintDoesNotOfferOffOnAModelThatAlwaysReasons(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model             string
		wantCannotDisable bool
	}{
		{"kimi-k3", true},
		{"kimi-k2.7-code", true},
		{"kimi-k2-thinking", true},
		{"kimi-k2.6", false},
	} {
		s := ReasoningSupportFor(tc.model, reasoning.ProviderOpenAI)
		if !s.Supported {
			t.Errorf("%s reasons, so the hint must say so", tc.model)
		}
		if s.CannotDisable != tc.wantCannotDisable {
			t.Errorf("%s CannotDisable = %v, want %v", tc.model, s.CannotDisable, tc.wantCannotDisable)
		}
	}
}

func TestTheBedrockHintReportsTheRefusalTheDoorWillGive(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model             string
		wantCannotDisable bool
	}{
		{"moonshot.kimi-k2-thinking", true},
		{"us.deepseek.r1-v1:0", true},
		{"openai.gpt-oss-120b-1:0", true},
		{"us.xai.grok-4.6", true},
		{"us.amazon.nova-2-lite-v1:0", false},
		{"us.amazon.nova-pro-v1:0", false},
	} {
		t.Run(tc.model, func(t *testing.T) {
			t.Parallel()

			wire := reasoning.ResolveOff(tc.model, reasoning.ProviderBedrock)
			if got := wire == reasoning.OffUnsupported; got != tc.wantCannotDisable {
				t.Fatalf("%s: wire refuses = %v, want %v — the case no longer covers what it claims",
					tc.model, got, tc.wantCannotDisable)
			}
			if got := ReasoningSupportFor(tc.model, reasoning.ProviderBedrock).CannotDisable; got != tc.wantCannotDisable {
				t.Errorf("%s: hint CannotDisable = %v, want %v — a consumer building its interface from the hint "+
					"would offer a control that the door answers with a typed error",
					tc.model, got, tc.wantCannotDisable)
			}
		})
	}
}

func TestTheHintNamesTheMechanismTheDoorActuallyUses(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  ReasoningMechanism
	}{
		{"gemma-4-31b-it", ReasoningMechanismAdaptive},
		{"gemma-4-26b-a4b-it", ReasoningMechanismAdaptive},
		{"gemini-2.5-flash", ReasoningMechanismBudget},
		{"gemini-3-pro", ReasoningMechanismAdaptive},
	} {
		t.Run(tc.model, func(t *testing.T) {
			t.Parallel()
			if got := ReasoningSupportFor(tc.model, reasoning.ProviderGoogleAI).Mechanism; got != tc.want {
				t.Errorf("%s: hint Mechanism = %v, want %v — a consumer builds its control from this field, "+
					"and the door drops a budget it advertises", tc.model, got, tc.want)
			}
		})
	}
}
