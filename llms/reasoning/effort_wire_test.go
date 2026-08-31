package reasoning

import "testing"

func TestAcceptsEffortWire(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model  string
		accept bool
	}{
		{"qwen3-next-80b-a3b-thinking", false},
		{"dashscope/qwen3-next-80b-a3b-thinking", false},
		{"qwen3.7-max", false},
		{"qwq-32b", false},
		{"kimi-k2.7-code-highspeed", true},
		{"moonshot/kimi-k2.7-code", true},
		{"kimi-k2.6", true},
		{"kimi-k2-thinking", true},
		{"kimi-k3", true},
		{"moonshot/kimi-k3", true},
		{"gpt-5.5", true},
		{"o3-mini", true},
		{"deepseek-v4-pro", true},
		{"glm-5-turbo", true},
		{"minimax-m3", true},
	} {
		if got := AcceptsEffortWire(tc.model); got != tc.accept {
			t.Errorf("AcceptsEffortWire(%q) = %v, want %v", tc.model, got, tc.accept)
		}
	}
}

func TestResolveOffOnDoorsThatRejectEffort(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  OffWire
	}{
		{"qwen3-next-80b-a3b-thinking", OffUnsupported},
		{"kimi-k2.7-code-highspeed", OffUnsupported},
		{"kimi-k2.6", OffEffortNone},
		{"kimi-k2-thinking", OffUnsupported},
		{"kimi-k3", OffEffortNone},
		{"deepseek-v4-pro", OffEffortNone},
		{"glm-5-turbo", OffEffortNone},
	} {
		if got := ResolveOff(tc.model, ProviderOpenAI); got != tc.want {
			t.Errorf("ResolveOff(%q, ProviderOpenAI) = %v, want %v", tc.model, got, tc.want)
		}
	}
}
