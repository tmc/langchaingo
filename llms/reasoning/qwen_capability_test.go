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
