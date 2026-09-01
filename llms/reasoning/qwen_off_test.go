package reasoning

import "testing"

func TestQwenHybridsSpellOffWithTheDashScopeFlag(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"qwen3-8b", "qwen3-32b", "qwen3-30b-a3b", "qwen3-235b-a22b",
		"dashscope/qwen3-8b",
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffDisableDashScope {
			t.Errorf("ResolveOff(%q, openai) = %v, want OffDisableDashScope", model, got)
		}
	}
}

func TestQwenHybridsAreReasoningModels(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"qwen3-8b", "qwen3-235b-a22b", "dashscope/qwen3-30b-a3b"} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, want true", model)
		}
	}
	for _, model := range []string{"qwen3-coder-plus", "qwen3-vl-plus"} {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, want false", model)
		}
	}
}

func TestBedrockQwenKeepsItsOwnAnswer(t *testing.T) {
	t.Parallel()

	if got := ResolveOff("qwen.qwen3-32b-v1:0", ProviderBedrock); got != OffOmit {
		t.Errorf("the bedrock spelling must keep OffOmit, got %v", got)
	}
}
