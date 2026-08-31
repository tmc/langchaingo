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
