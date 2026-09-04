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
		{"qwen3.8-max", true},
		{"qwen3.8-flash", true},
		{"qwen3.8-2.4t-a95b", true},
		{"dashscope/qwen3.8-max", true},
		{"qwen3.5-plus", false},
		{"qwq-32b", false},
		{"kimi-k2.7-code-highspeed", true},
		{"moonshot/kimi-k2.7-code", true},
		{"kimi-k2.6", true},
		{"kimi-k2-thinking", true},
		{"kimi-k3", true},
		{"moonshot/kimi-k3", true},
		{"gpt-5.5", true},
		{"gpt-4o", false},
		{"gpt-4o-mini", false},
		{"gpt-4.1", false},
		{"gpt-4-turbo", false},
		{"gpt-3.5-turbo", false},
		{"openai/gpt-4o", false},
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

func TestRejectsMinP(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model  string
		reject bool
	}{
		{"claude-sonnet-4-5", true},
		{"claude-sonnet-4-5-20250929", true},
		{"anthropic/claude-sonnet-4.5", true},
		{"us.anthropic.claude-opus-4-6-v1:0", true},
		{"claude-opus-4-7", true},
		{"claude-haiku-4-5", true},
		{"gpt-5.5", false},
		{"grok-4.6", false},
		{"qwen3-32b", false},
		{"deepseek-v3.2", false},
	} {
		if got := RejectsMinP(tc.model); got != tc.reject {
			t.Errorf("RejectsMinP(%q) = %v, want %v", tc.model, got, tc.reject)
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
		{"kimi-k3", OffUnsupported},
		{"deepseek-v4-pro", OffEffortNone},
		{"glm-5-turbo", OffDisableThinkingObject},
	} {
		if got := ResolveOff(tc.model, ProviderOpenAI); got != tc.want {
			t.Errorf("ResolveOff(%q, ProviderOpenAI) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestTheGrokBuildAliasThatIsReallyGrok45(t *testing.T) {
	t.Parallel()

	if !AcceptsEffortWire("grok-build-latest") {
		t.Error("grok-build-latest drops the effort field, but the vendor answers low with 47 " +
			"reasoning tokens and high with 73")
	}
	if got := ResolveOff("grok-build-latest", ProviderOpenAI); got != OffUnsupported {
		t.Errorf("ResolveOff(grok-build-latest) = %v, want unsupported: the vendor refuses "+
			"effort none on this model", got)
	}

	for _, model := range []string{"grok-build-0.1", "grok-code-fast-1"} {
		if AcceptsEffortWire(model) {
			t.Errorf("%s carries the effort field, but the vendor refuses it by name", model)
		}
	}
}

func TestEffortWithTools(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		rule  EffortToolsRule
	}{
		{"gpt-5.6-sol", EffortToolsDisable},
		{"gpt-5.6-terra", EffortToolsDisable},
		{"openai/gpt-5.6", EffortToolsDisable},
		{"gpt-5.5", EffortToolsOmit},
		{"gpt-5.4-nano", EffortToolsOmit},
		{"gpt-5.4-mini", EffortToolsOmit},
		{"gpt-5.2", EffortToolsFree},
		{"gpt-5.1", EffortToolsFree},
		{"gpt-5-mini", EffortToolsFree},
		{"gpt-4o", EffortToolsFree},
		{"claude-opus-5", EffortToolsFree},
	} {
		if got := EffortWithTools(tc.model); got != tc.rule {
			t.Errorf("EffortWithTools(%q) = %v, want %v", tc.model, got, tc.rule)
		}
	}
}
