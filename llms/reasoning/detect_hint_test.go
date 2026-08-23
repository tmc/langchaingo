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
