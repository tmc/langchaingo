package reasoning

import "testing"

func TestQwenCommercialHybridsThinkOnlyWhenAsked(t *testing.T) {
	t.Parallel()

	cases := []struct {
		model string
		want  bool
	}{
		{"qwen-plus", true},
		{"qwen-flash", true},
		{"qwen3-max", true},
		{"dashscope/qwen3-max", true},
		{"DashScope/Qwen-Plus", true},
		{"qwen3-8b", false},
		{"qwen3.7-plus", false},
		{"qwen3-coder-plus", false},
		{"qwen3-vl-plus", true},
		{"qwen3-vl-flash", true},
		{"openrouter/qwen/qwen-plus", false},
		{"qwen-plus-latest", false},
	}
	for _, tc := range cases {
		if got := QwenThinkingEnabledByFlag(tc.model); got != tc.want {
			t.Errorf("QwenThinkingEnabledByFlag(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestQwenCommercialHybridsAreOptInReasoners(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"qwen-plus", "qwen-flash", "qwen3-max"} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, want true", model)
		}
		if !ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = false, want true", model)
		}
		if got := ResolveOff(model, ProviderOpenAI); got != OffOmit {
			t.Errorf("ResolveOff(%q) = %v, want OffOmit — they do not think unless asked", model, got)
		}
	}
}
