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

func TestQwen3HybridsAreInTheCatalogueNowThatOffIsExpressible(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"qwen3-32b", "qwen3-235b-a22b", "qwen3-vl-plus"} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, want true", model)
		}
	}
	for _, model := range []string{"qwen3-coder-plus"} {
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
		"qwen3-32b", "qwen3-235b-a22b",
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

	for _, model := range []string{"qwen-plus", "qwen-flash", "qwen3-max"} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, want true", model)
		}
		if !ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = false, but a bare call returns no reasoning until the vendor flag asks", model)
		}
	}

	for _, model := range []string{"gpt-mini-latest", "qwen3-coder-plus"} {
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
		"XiaomiMiMo/MiMo-V2.5-tts", "XiaomiMiMo/MiMo-V2.5-tts-voiceclone",
		"XiaomiMiMo/MiMo-V2.5-tts-voicedesign",
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
		"XiaomiMiMo/MiMo-V2.5",
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

func TestOllamaTagSpellingResolvesLikeTheHyphenOne(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"gpt-oss:120b", "gpt-oss:20b", "ollama_cloud/gpt-oss:120b"} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, but the hyphen spelling of the same model is true", model)
		}
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported like the hyphen spelling", model, got)
		}
	}
}

func TestFamiliesThatRefuseAnExplicitOff(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"aion-2.0", "aion-3.0", "aion-3.0-mini",
		"step-3.5-flash", "step-3.7-flash",
		"reka-flash-3", "fugu-ultra", "nex-n2-mini", "nex-n2-pro",
		"arcee-ai/trinity-large-thinking",
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: the door refuses or ignores an explicit off", model, got)
		}
	}

	if IsReasoningModel("aion-rp-llama-3.1-8b") {
		t.Error("aion-rp is not a reasoning model and must stay out of the mandatory set")
	}
}

func TestFamiliesMeasuredAgainstTheirBareCall(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"lfm-2.5-2.6b", "liquid/lfm-2.5-2.6b:free", "seed-2-1-turbo", "bytedance-seed/seed-2-1-turbo",
		"inkling", "inkling-small", "thinkingmachines/inkling", "openrouter/thinkingmachines/inkling-small:free",
	} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, but a bare call already returns a reasoning block", model)
		}
		if ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = true, but the model reasons without being asked", model)
		}
	}

	if got := ResolveOff("lfm-2.5-2.6b", ProviderOpenAI); got != OffUnsupported {
		t.Errorf("ResolveOff(lfm-2.5) = %v, want OffUnsupported: the door answers 400 on an explicit off", got)
	}
	if got := ResolveOff("seed-2-1-turbo", ProviderOpenAI); got != OffEffortNone {
		t.Errorf("ResolveOff(seed-2-1-turbo) = %v, want OffEffortNone: the off returns no reasoning", got)
	}

	if !ThinkingOptIn("ernie-4.5-vl-424b-a47b") {
		t.Error("ThinkingOptIn(ernie-4.5) = false, but a bare call returns no reasoning and an effort turns it on")
	}

	if got := ResolveOff("thinkingmachines/inkling", ProviderOpenAI); got != OffEffortNone {
		t.Errorf("ResolveOff(inkling) = %v, want OffEffortNone: the off returns no reasoning", got)
	}
}

func TestTheThinkingMarkerNeedsAWordBoundary(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"kimi-k2-thinking", "qwen3-30b-a3b-thinking-2507", "Qwen3-235B-A22B-Thinking-2507",
		"arcee-ai/trinity-large-thinking", "lfm-2-thinking",
	} {
		if !ThinkingMarkedInName(model) {
			t.Errorf("ThinkingMarkedInName(%q) = false, but the name carries the marker as its own token", model)
		}
	}

	for _, model := range []string{
		"thinkingmachines/inkling", "openrouter/thinkingmachines/inkling-small:free",
		"gpt-5.4", "claude-sonnet-4-5",
	} {
		if ThinkingMarkedInName(model) {
			t.Errorf("ThinkingMarkedInName(%q) = true, but the marker only appears inside a longer word", model)
		}
	}
}

func TestDeepSeekR1CannotStopReasoning(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"deepseek-r1", "deepseek-r1-0528", "deepseek/deepseek-r1", "deepseek-reasoner",
		"deepseek-ai/DeepSeek-R1-Turbo", "azure/deepseek-r1",
		"deepseek-ai/DeepSeek-R1-Distill-Qwen-32B",
	} {
		if got := ResolveOff(model, ProviderOpenAI); got != OffUnsupported {
			t.Errorf("ResolveOff(%q) = %v, want OffUnsupported: the off is either refused or ignored", model, got)
		}
	}

	for _, model := range []string{"deepseek-v3.1", "deepseek-v3.2", "deepseek-chat-v3.1"} {
		if got := ResolveOff(model, ProviderOpenAI); got == OffUnsupported {
			t.Errorf("ResolveOff(%q) = OffUnsupported, but the hybrids reason only when asked", model)
		}
	}
}

func TestEveryChatSpellingOfTheGPTLineAgrees(t *testing.T) {
	t.Parallel()

	spellings := []string{
		"gpt-5-chat", "gpt-5-chat-latest", "gpt-5.1-chat-latest", "gpt-5.2-chat",
		"gpt-5.2-chat-latest", "gpt-5.3-chat-latest", "gpt-chat-latest",
		"openai/gpt-5.2-chat", "azure/gpt-5-chat",
	}
	for _, model := range spellings {
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true, but the chat variants answer without reasoning", model)
		}
		if caps := OpenAIReasoningCapsFor(model); caps.Known {
			t.Errorf("OpenAIReasoningCapsFor(%q).Known = true, so the door would pin the caller's temperature", model)
		}
	}

	for _, model := range []string{"deepseek-chat-v3.1", "deepseek-chat-v3.2", "gpt-5.2", "gpt-5.4"} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false: the chat filter reached past the GPT chat variants", model)
		}
	}
}

func TestTheMistralListingNamesWeMissed(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"mistral-vibe-cli-fast", "mistral-vibe-cli-latest", "mistral-vibe-cli-with-tools",
	} {
		if !IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = false, but the vendor reports reasoning and a live "+
				"call with effort returns it", model)
		}
		if !ThinkingOptIn(model) {
			t.Errorf("ThinkingOptIn(%q) = false, but a bare call returns no reasoning", model)
		}
	}

	if !IsReasoningModel("zai-glm-5-2") {
		t.Error("IsReasoningModel(\"zai-glm-5-2\") = false, but it is the same model as glm-5-2 " +
			"under the spelling Mistral serves it")
	}
	if IsReasoningModel("zai-glm-5-2") != IsReasoningModel("glm-5-2") {
		t.Error("the two live spellings of one model disagree")
	}
}

func TestAliasNamesTheVendorListingsCarry(t *testing.T) {
	t.Parallel()

	if !IsReasoningModel("labs-leanstral-1-5") || !IsReasoningModel("labs-leanstral-1-5-1") {
		t.Error("the Mistral listing marks labs-leanstral as reasoning, and the drift test " +
			"carries it as an accounted divergence only because the vendor will not serve it")
	}

	if !IsReasoningModel("glm-flash-latest") {
		t.Error("glm-flash-latest is the alias of glm-5.3-flash, which reasons")
	}
	if got := ResolveOff("glm-flash-latest", ProviderOpenAI); got != OffUnsupported {
		t.Errorf("ResolveOff(glm-flash-latest) = %v, want unsupported: its target answers "+
			"an explicit disable with 159 characters of reasoning", got)
	}
}
