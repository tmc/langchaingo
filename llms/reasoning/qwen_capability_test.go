package reasoning

import "testing"

func TestQwenThinkingRequiresStream(t *testing.T) {
	t.Parallel()

	cases := []struct {
		model string
		want  bool
	}{
		{"qwen3-8b", true},
		{"qwen3-14b", true},
		{"qwen3-32b", true},
		{"qwen3-30b-a3b", true},
		{"qwen3-235b-a22b", true},
		{"qwen3-1.7b", true},
		{"dashscope/qwen3-8b", true},
		{"DashScope/Qwen3-32B", true},
		{"qwen3-coder-plus", false},
		{"qwen3-coder-30b-a3b-instruct", false},
		{"qwen3-vl-plus", false},
		{"qwen3-vl-flash", false},
		{"qwen3-max", false},
		{"qwen3-rerank", false},
		{"qwen3-next-80b-a3b-thinking", false},
		{"qwen3-next-80b-a3b-instruct", false},
		{"qwen3-235b-a22b-thinking-2507", false},
		{"qwen3.5-flash", false},
		{"qwen2.5-7b", false},
		{"qwen/qwen3-32b", false},
		{"openrouter/qwen/qwen3-8b", false},
		{"deepinfra/Qwen/Qwen3-32B", false},
	}
	for _, tc := range cases {
		if got := QwenThinkingRequiresStream(tc.model); got != tc.want {
			t.Errorf("QwenThinkingRequiresStream(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestQwenThinkingOffByFlag(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		flag  bool
	}{
		{"qwen3.7-plus", true},
		{"qwen3.7-max", true},
		{"qwen3.7-max-preview", true},
		{"qwen3.6-plus", true},
		{"qwen3.6-max-preview", true},
		{"qwen3.5-flash", true},
		{"dashscope/qwen3.5-plus", true},
		{"qwen3.8-max", false},
		{"qwen3.5-35b-a3b", true},
		{"qwen3.5-27b", true},
		{"qwen3.5-122b-a10b", true},
		{"qwen3.5-397b-a17b", true},
		{"qwen3.6-27b", true},
		{"qwen3.6-35b-a3b", true},
		{"qwen3-8b", false},
		{"qwen-plus", false},
		{"openrouter/qwen/qwen3.7-plus", false},
		{"deepinfra/Qwen/qwen3.7-plus", false},
	} {
		if got := QwenThinkingOffByFlag(tc.model); got != tc.flag {
			t.Errorf("QwenThinkingOffByFlag(%q) = %v, want %v", tc.model, got, tc.flag)
		}
	}
}

func TestQwenDisableWire(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		off   OffWire
	}{
		{"qwen3.8-max", OffEffortNone},
		{"qwen3.8-flash", OffEffortNone},
		{"qwen3.7-plus", OffDisableDashScope},
		{"qwen3.6-flash", OffDisableDashScope},
		{"qwen3.5-plus", OffDisableDashScope},
		{"qwen3.5-35b-a3b", OffDisableDashScope},
		{"qwen3.6-27b", OffDisableDashScope},
		{"qwen3-8b", OffDisableDashScope},
		{"openrouter/qwen/qwen3.7-plus", OffUnsupported},
	} {
		if got := ResolveOff(tc.model, ProviderOpenAI); got != tc.off {
			t.Errorf("ResolveOff(%q) = %v, want %v", tc.model, got, tc.off)
		}
	}
}
