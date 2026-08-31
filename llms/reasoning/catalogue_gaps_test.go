package reasoning

import "testing"

func TestModelsThatReasonAreInTheCatalogue(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"magistral-medium-latest",
		"mistral/magistral-medium-latest",
		"magistral-small",
		"qwen3.7-max",
		"qwen3.5-35b-a3b",
		"qwen3.5-397b-a17b",
		"qwen3.6-plus",
		"dashscope/qwen3.7-max",
		"gemini-flash-latest",
		"gemini-robotics-er-1.6-preview",
	} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, want true — measured reasoning at the vendor", model)
		}
	}
}

func TestQwenCoderDoesNotReasonUnasked(t *testing.T) {
	t.Parallel()

	if IsReasoningModel("qwen3-coder-plus") {
		t.Error(`IsReasoningModel("qwen3-coder-plus") = true, want false — a plain call returns no reasoning block`)
	}
}

func TestNonChatGoogleSurfacesAreNotThinkingModels(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gemini-2.5-pro-preview-tts",
		"gemini-2.5-flash-preview-tts",
		"gemini-3.5-live-translate-preview",
	} {
		if GeminiSupportsThinking(model) {
			t.Errorf("GeminiSupportsThinking(%q) = true, want false — the endpoint refuses a chat completion", model)
		}
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, want false", model)
		}
	}
}

func TestQwen3StaysOutUntilEnableThinkingIsSent(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"qwen3-32b", "qwen3-235b-a22b", "qwen3-vl-plus"} {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, want false", model)
		}
		if !LikelyReasoningModel(model) {
			t.Errorf("LikelyReasoningModel(%q) = false, want true — the hint should stay optimistic", model)
		}
	}
}

func TestChatGoogleFamiliesStillThink(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-3-pro", "gemma-4-26b"} {
		if !GeminiSupportsThinking(model) {
			t.Errorf("GeminiSupportsThinking(%q) = false, want true", model)
		}
	}
}

func TestQwenTakesNoEffortOnTheWire(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"qwen3-32b", "qwen3-235b-a22b", "qwen3-vl-plus",
		"qwen3.5-35b-a3b", "qwen3.7-max", "dashscope/qwen3.7-max",
	} {
		if AcceptsEffortWire(model) {
			t.Errorf("AcceptsEffortWire(%q) = true, want false", model)
		}
	}
}

func TestMeasuredModelsSplitByWhenTheyReason(t *testing.T) {
	t.Parallel()

	alwaysOn := []string{
		"gpt-latest", "grok-latest", "glm-latest", "kimi-latest",
		"hy4-preview", "qwen3.8-flash", "qwen3.8-max",
		"ling-3.0-flash", "longcat-2.0",
	}
	for _, model := range alwaysOn {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, want true", model)
		}
		if ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = true, but a bare call already returns reasoning", model)
		}
	}

	if !IsReasoningModel("solar-pro4") {
		t.Error(`IsReasoningModel("solar-pro4") = false, want true`)
	}
	if !ThinkingOptIn("solar-pro4") {
		t.Error(`ThinkingOptIn("solar-pro4") = false, but a bare call returns no reasoning until an effort asks`)
	}

	for _, model := range []string{"gpt-mini-latest", "qwen3-max", "qwen3-coder-plus"} {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, but neither a bare call nor an effort returns reasoning", model)
		}
	}
}

func TestModelsMeasuredNotToReasonStayOut(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"deepseek-chat-v3-0324", false},
		{"deepseek-chat-v3.1", true},
		{"minimax-m2-her", false},
		{"minimax-m2", true},
		{"minimax-m2.7", true},
		{"aion-rp-llama-3.1-8b", false},
		{"aion-2.0", true},
		{"aion-3.0", true},
		{"gemini-2.5-flash-image", false},
		{"gemini-2.5-flash", true},
	} {
		if got := IsReasoningModel(tc.model); got != tc.want {
			t.Errorf("IsReasoningModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestMandatoryThinkingCannotBeDisabled(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"glm-5.3", "glm-5.3-flash"} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: the vendor answers "+
				"\"Reasoning is mandatory\" to an explicit disable", model, got)
		}
	}
	for _, model := range []string{"glm-4.6", "glm-5"} {
		if got := ResolveOff(model, ProviderOpenAI); got == OffUnsupported {
			t.Errorf("ResolveOff(%q) = OffUnsupported, but this generation was never measured to refuse a disable", model)
		}
	}
}

func TestGoogleSurfacesThatRefuseThinkingStayOut(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gemini-2.5-flash-preview-tts",
		"gemini-2.5-pro-preview-tts",
		"gemini-3.1-flash-tts-preview",
		"gemini-3.5-transcribe",
		"gemini-3.5-transcribe-live",
	} {
		if GeminiSupportsThinking(model) {
			t.Errorf("GeminiSupportsThinking(%q) = true, but the vendor answers "+
				"\"Thinking level is not supported for this model\"", model)
		}
	}

	for _, model := range []string{"gemini-2.5-flash", "gemini-3.5-flash", "gemini-2.5-pro"} {
		if !GeminiSupportsThinking(model) {
			t.Errorf("GeminiSupportsThinking(%q) = false, want true", model)
		}
	}
}

func TestDeepSeekV31ReasonsOnlyWhenAsked(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"deepseek-v3.1", "deepseek-chat-v3.1", "deepseek-v3.1-terminus",
	} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, but an effort returns reasoning", model)
		}
		if !ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = false, but a bare call returns none", model)
		}
	}

	for _, model := range []string{"deepseek-v3", "deepseek-v3-0324", "deepseek-chat-v3-0324"} {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, but it answers zero both ways", model)
		}
	}
}

func TestGrokModelsThatTakeNoEffortOnTheWire(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"grok-code-fast-1", "grok-code-fast", "grok-4.20", "grok-build-0.1"} {
		if AcceptsEffortWire(model) {
			t.Errorf("AcceptsEffortWire(%q) = true, but the vendor answers "+
				"\"does not support parameter reasoningEffort\" to any effort", model)
		}
	}

	for _, model := range []string{"grok-4", "grok-4.3", "grok-3-mini", "grok-4-1-fast-reasoning"} {
		if !AcceptsEffortWire(model) {
			t.Errorf("AcceptsEffortWire(%q) = false, but the vendor accepts an effort", model)
		}
	}
}

func TestGrokGenerationsThatCannotStopThinking(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"grok-4.5", "grok-4.5-latest", "grok-4.6"} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: the vendor answers "+
				"\"This model does not support reasoning_effort value none\"", model, got)
		}
	}

	for _, model := range []string{"grok-4", "grok-4.3", "grok-3-mini"} {
		if got := ResolveOff(model, ProviderOpenAI); got == OffUnsupported {
			t.Errorf("ResolveOff(%q) = OffUnsupported, but the vendor accepts an explicit off", model)
		}
	}
}

func TestMiniMaxM2CannotStopThinking(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"minimax-m2", "minimax-m2.1", "minimax-m2.5", "minimax-m2.7",
		"minimax-m2.7-highspeed", "minimax.minimax-m2.5",
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: the vendor answers "+
				"\"Reasoning is mandatory for this endpoint and cannot be disabled\", "+
				"and the native route keeps thinking regardless", model, got)
		}
	}

	if got := ResolveOff("minimax-m3", ProviderOpenAI); got == OffUnsupported {
		t.Error(`ResolveOff("minimax-m3") = OffUnsupported, but an explicit off returns no reasoning`)
	}
}

func TestGLMLatestCannotStopThinking(t *testing.T) {
	t.Parallel()

	if got := ResolveOff("glm-latest", ProviderOpenAI); got != OffUnsupported {
		t.Errorf("ResolveOff(\"glm-latest\") = %v, want OffUnsupported: an explicit off "+
			"still returns reasoning, and the alias points at the mandatory 5.3 line", got)
	}

	for _, model := range []string{"glm-4.6", "glm-5", "glm-5.1", "glm-5.2"} {
		if got := ResolveOff(model, ProviderOpenAI); got == OffUnsupported {
			t.Errorf("ResolveOff(%q) = OffUnsupported, but an explicit off returns no reasoning", model)
		}
	}
}

func TestKimiTakesEffortButNotAlwaysAnOff(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"kimi-k2.5", "kimi-k2.6", "kimi-k2.7-code", "kimi-k2-thinking", "kimi-k3",
	} {
		if !AcceptsEffortWire(model) {
			t.Errorf("AcceptsEffortWire(%q) = false, but every route measured accepts an effort", model)
		}
	}

	for _, model := range []string{"kimi-k2-thinking", "kimi-k2.7-code", "kimi-k2.7-code-highspeed"} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: the vendor refuses an explicit off", model, got)
		}
	}

	for _, model := range []string{"kimi-k2.5", "kimi-k2.6"} {
		if got := ResolveOff(model, ProviderOpenAI); got == OffUnsupported {
			t.Errorf("ResolveOff(%q) = OffUnsupported, but an explicit off returns no reasoning", model)
		}
	}
}

func TestNonChatNvidiaSurfacesAreNotReasoningModels(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"nvidia/Nemotron-3-Embed-8B", "nvidia/Nemotron-3-Embed-1B-BF16",
		"nvidia/llama-nemotron-embed-vl-1b-v2", "nvidia/llama-nemotron-rerank-vl-1b-v2",
		"nvidia/Nemotron-3.5-ASR-Streaming-Multilingual-0.6b",
		"nvidia/Nemotron-Content-Safety-3.5", "nvidia/nemotron-3.5-content-safety",
	} {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, but the route serves no chat completion", model)
		}
	}

	for _, model := range []string{
		"nvidia/NVIDIA-Nemotron-Nano-9B-v2", "nvidia/nemotron-3-super-120b-a12b",
		"nvidia/Nemotron-3-Nano-Omni-30B-A3B-Reasoning", "nvidia.nemotron-nano-9b-v2",
	} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, but the chat route still reasons", model)
		}
	}
}

func TestMistralAliasesOfTheSameModelAgree(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"mistral-medium-latest", "mistral-medium-2604", "mistral-medium", "mistral-medium-3-5",
		"mistral-small-latest", "mistral-small-2603", "mistralai/mistral-medium-latest",
	} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, but the vendor reasons on an explicit effort", model)
		}
		if !ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = false, but a bare call returns no reasoning", model)
		}
	}

	for _, model := range []string{
		"mistral-medium-2505", "mistral-medium-2508", "mistral-large-latest", "mistral-large-2512",
	} {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, but the vendor answers 400 reasoning_effort is not enabled", model)
		}
	}
}

func TestMagistralExpressesItsOffByOmission(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"magistral-medium-latest", "magistral-small-latest", "mistral/magistral-medium-latest"} {
		if !ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = false, but a bare call returns no reasoning", model)
		}
		if got := ResolveOff(model, ProviderOpenAI); got != OffOmit {
			t.Errorf("ResolveOff(%q) = %v, want OffOmit: the none token starts a think block instead of stopping one", model, got)
		}
	}
}

func TestGPTOSSCannotStopReasoning(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gpt-oss-120b", "gpt-oss-20b", "openai/gpt-oss-120b",
		"openai.gpt-oss-120b-1:0", "gpt-oss-safeguard-20b",
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: an explicit off still returns reasoning", model, got)
		}
	}
}
