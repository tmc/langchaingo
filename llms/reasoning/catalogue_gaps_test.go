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
