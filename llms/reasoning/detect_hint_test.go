package reasoning

import "testing"

func TestLikelyReasoningModel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
		note  string
	}{
		{"gpt-5.5", true, "already classified"},
		{"claude-sonnet-5", true, "already classified"},
		{"o3-mini", true, "already classified"},

		{"gpt-6", true, "generation newer than this build"},
		{"gpt-7-turbo", true, "generation newer than this build"},
		{"claude-opus-6", true, "generation newer than this build"},
		{"claude-sonnet-7", true, "generation newer than this build"},
		{"gemini-4-pro", true, "generation newer than this build"},
		{"gemma-5", true, "generation newer than this build"},
		{"o5", true, "o-series successor"},
		{"openai/gpt-6", true, "proxy prefix stripped"},

		{"gpt-4o", false, "pre-reasoning generation"},
		{"gpt-4.1", false, "pre-reasoning generation"},
		{"gpt-3.5-turbo", false, "pre-reasoning generation"},
		{"claude-3-5-sonnet-latest", false, "pre-thinking generation"},
		{"claude-2.1", false, "pre-thinking generation"},
		{"anthropic.claude-instant-v1", false, "pre-thinking generation"},
		{"gemini-1.5-pro", false, "pre-thinking generation"},
		{"gemini-2.0-flash", false, "pre-thinking generation"},
		{"gemma-3", false, "pre-thinking generation"},

		{"", false, "not a model"},
		{"mistral-large-latest", false, "family this package does not classify"},
	}
	for _, tc := range cases {
		if got := LikelyReasoningModel(tc.model); got != tc.want {
			t.Errorf("LikelyReasoningModel(%q) = %v, want %v (%s)", tc.model, got, tc.want, tc.note)
		}
	}
}

func TestNonTextModalitiesAreNotGuessedAsReasoning(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"gpt-image-1", false},
		{"gemini-embedding-001", false},
		{"openai/gpt-image-1", false},
		{"gpt-5.5", true},
		{"gemini-2.5-flash", true},
		{"claude-opus-4-6", true},
		{"o3", true},
	} {
		if got := LikelyReasoningModel(tc.model); got != tc.want {
			t.Errorf("LikelyReasoningModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestFlashLiteIsRecognisedOnlyWithinTheGoogleFamilies(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"gemini-2.5-flash-lite", true},
		{"gemini-3.1-flash-lite", true},
		{"gemini-3.5-flash-lite", true},
		{"gemma-4-flash-lite", true},
		{"models/gemini-2.5-flash-lite-preview-06-17", true},
		{"some-flash-lite-thing", false},
		{"llama-4-flash-lite", false},
		{"gemini-2.5-flash", false},
	} {
		if got := GeminiThinkingOffByDefault(tc.model); got != tc.want {
			t.Errorf("GeminiThinkingOffByDefault(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestQwenHintIsOptimisticWithoutMovingTheWire(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"qwen3-coder-flash", true},
		{"qwen3-coder-plus", true},
		{"qwen3-14b", true},
		{"qwen3.5-flash", true},
		{"qwen3.6-plus", true},
		{"qwen3.7-max", true},
		{"qwen4-max", true},
		{"qwen-plus", true},
		{"qwen-turbo", true},
		{"qwen-max", true},
		{"qwen/qwen3-32b", true},
		{"qwen.qwen3-32b-v1:0", true},
		{"Qwen/Qwen3.6-27B-FP8", true},
		{"qvq-max", true},
		{"qvq-72b-preview", true},
		{"qwen2.5-72b-instruct", false},
		{"qwen2-7b", false},
		{"qwen1.5-14b", false},
		{"qwen-1.8b", false},
		{"qwen-2.5-coder", false},
	} {
		if got := LikelyReasoningModel(tc.model); got != tc.want {
			t.Errorf("LikelyReasoningModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestTheHintDoesNotOfferReasoningToNonTextModels(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gpt-audio", "gpt-realtime", "gpt-4o-transcribe", "whisper-1",
		"tts-1", "text-embedding-3-large", "gpt-image-1",
	} {
		if LikelyReasoningModel(model) {
			t.Errorf("%q answers in another modality, so no reasoning control belongs on it", model)
		}
	}
}

func TestTheHintAgreesWithTheMatchersOnChatLatest(t *testing.T) {
	t.Parallel()

	const model = "gpt-5-chat-latest"
	if LikelyReasoningModel(model) {
		t.Errorf("%q is excluded by the other two matchers, so the hint must exclude it too", model)
	}
}
