package reasoning

import "testing"

func TestNovaReasoningModelIsTheSecondGenerationOnly(t *testing.T) {
	for _, model := range []string{
		"amazon.nova-2-lite-v1:0", "us.amazon.nova-2-lite-v1:0",
		"global.amazon.nova-2-pro-preview-20251202-v1:0",
	} {
		if !IsNovaReasoningModel(model) {
			t.Errorf("%s takes reasoningConfig", model)
		}
	}
	for _, model := range []string{
		"amazon.nova-pro-v1:0", "us.amazon.nova-lite-v1:0", "amazon.nova-micro-v1:0",
		"us.anthropic.claude-opus-5", "openai.gpt-oss-120b-1:0",
	} {
		if IsNovaReasoningModel(model) {
			t.Errorf("%s does not take reasoningConfig", model)
		}
	}
}

func TestNovaEffortCollapsesOntoThreeLevels(t *testing.T) {
	cases := map[string]string{
		"minimal": "low", "low": "low", "medium": "medium",
		"high": "high", "xhigh": "high", "max": "high", "": "high",
	}
	for in, want := range cases {
		if got := NovaEffort(in); got != want {
			t.Errorf("NovaEffort(%q) = %q, want %q — Bedrock rejects anything outside low/medium/high", in, got, want)
		}
	}
}

func TestNovaRejectsSamplingOnlyAtTopEffort(t *testing.T) {
	if !NovaClearsInferenceConfigAt("max") || !NovaClearsInferenceConfigAt("high") {
		t.Error("temperature, topP and topK are an error at the top effort")
	}
	if NovaClearsInferenceConfigAt("low") || NovaClearsInferenceConfigAt("medium") {
		t.Error("sampling params are allowed below the top effort")
	}
}

func TestGrokIsRecognisedByItsBedrockName(t *testing.T) {
	if !IsGrokModel("us.xai.grok-4.6") || !IsGrokModel("xai.grok-4.3") {
		t.Fatal("both generations are Grok")
	}
	if IsGrokModel("us.amazon.nova-2-lite-v1:0") {
		t.Error("Nova is not Grok")
	}
}

func TestGrokEffortFollowsTheGeneration(t *testing.T) {
	// xAI publishes the accepted set per model: 4.3 keeps none, 4.6 dropped it,
	// and every generation takes xhigh.
	cases := []struct{ model, in, want string }{
		{"xai.grok-4.3", "none", "none"},
		{"xai.grok-4.3", "max", "xhigh"},
		{"xai.grok-4.3", "xhigh", "xhigh"},
		{"us.xai.grok-4.6", "none", ""},
		{"us.xai.grok-4.6", "max", "xhigh"},
		{"us.xai.grok-4.6", "xhigh", "xhigh"},
		{"us.xai.grok-4.6", "medium", "medium"},
		{"us.xai.grok-4.6", "minimal", "low"},
	}
	for _, c := range cases {
		if got := GrokEffort(c.model, c.in); got != c.want {
			t.Errorf("GrokEffort(%s, %q) = %q, want %q", c.model, c.in, got, c.want)
		}
	}
}

func TestAlwaysReasoningModelCannotExpressOffOnAnyDoor(t *testing.T) {
	for _, model := range []string{"us.xai.grok-4.6", "xai.grok-4.6"} {
		for _, p := range []Provider{ProviderBedrock, ProviderOpenAI} {
			if got := ResolveOff(model, p); got != OffUnsupported {
				t.Errorf("ResolveOff(%s, %v) = %v, want OffUnsupported — omitting leaves thinking on",
					model, p, got)
			}
		}
	}
	// Nova defaults to disabled, so omitting really is off there.
	if got := ResolveOff("us.amazon.nova-2-lite-v1:0", ProviderBedrock); got != OffOmit {
		t.Errorf("ResolveOff(nova, bedrock) = %v, want OffOmit", got)
	}
}

func TestMechanismSeparatesTheBedrockFamilies(t *testing.T) {
	cases := []struct {
		model        string
		claudeFamily bool
		want         Mechanism
	}{
		{"us.amazon.nova-2-lite-v1:0", false, MechanismNovaReasoningConfig},
		{"us.xai.grok-4.6", false, MechanismGrokEffort},
		{"us.anthropic.claude-opus-4-5-20251101-v1:0", true, MechanismBudget},
	}
	for _, c := range cases {
		if got := ResolveMechanism(c.model, false, c.claudeFamily, true); got != c.want {
			t.Errorf("ResolveMechanism(%s) = %d, want %d", c.model, got, c.want)
		}
	}
}
