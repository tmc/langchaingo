package reasoning

import (
	"fmt"
	"testing"
)

func capabilityAnswers(model string) string {
	caps := OpenAIReasoningCapsFor(model)
	return fmt.Sprintf(
		"reasoning=%v likely=%v claude=%d openaiKnown=%v openaiEfforts=%v openaiTemp=%v gemini=%v optIn=%v off=%v",
		IsReasoningModel(model), LikelyReasoningModel(model), ClaudeReasoningKindFor(model),
		caps.Known, caps.Efforts, OpenAIAcceptsCustomTemperature(model),
		GeminiSupportsThinking(model), ThinkingOptIn(model),
		ResolveOff(model, ProviderUnknown))
}

func TestEverySpellingResolvesToOneEntry(t *testing.T) {
	t.Parallel()

	for _, group := range [][]string{
		{
			"claude-sonnet-4-5",
			"claude-sonnet-4.5",
			"anthropic/claude-sonnet-4-5",
			"us.anthropic.claude-sonnet-4-5-20250929-v1:0",
			"claude-sonnet-4-5@20250929",
			"claude-sonnet-4-5-20250929",
			"projects/p/locations/l/publishers/anthropic/models/claude-sonnet-4-5",
		},
		{"gpt-5.4", "openai/gpt-5.4", "GPT-5.4", "azure.gpt-5.4"},
		{"o3", "openai/o3", "O3"},
		{"gpt-oss-120b", "openai.gpt-oss-120b-1:0", "openai/gpt-oss-120b"},
		{
			"gemini-2.5-flash",
			"google/gemini-2.5-flash",
			"projects/p/locations/l/publishers/google/models/gemini-2.5-flash",
		},
		{"deepseek-r1", "deepseek.r1-v1:0", "us.deepseek.r1-v1:0"},
		{"deepseek-v3.2", "deepseek.v3.2", "us.deepseek.v3.2"},
		{"glm-4.7", "zai.glm-4.7"},
		{"minimax-m2", "minimax.minimax-m2"},
		{"kimi-k2-thinking", "moonshot.kimi-k2-thinking"},
		{"mistral-medium-3", "mistralai/mistral-medium-3"},
	} {
		canonical := group[0]
		want := capabilityAnswers(canonical)

		for _, form := range group[1:] {
			t.Run(form, func(t *testing.T) {
				t.Parallel()
				if got := capabilityAnswers(form); got != want {
					t.Errorf("a spelling of %s must resolve to the same entry\n canonical: %s\n      form: %s",
						canonical, want, got)
				}
			})
		}
	}
}

func TestASeparatorOnlyReferenceMatchesNothing(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"us.", "us.anthropic.", "azure.", "openai/", ""} {
		for _, form := range modelSpellings(model) {
			if form == "" {
				t.Errorf("modelSpellings(%q) yields an empty form, which every substring rule matches", model)
			}
		}
		if IsReasoningModel(model) {
			t.Errorf("IsReasoningModel(%q) = true: a reference with no model name cannot be a reasoning model", model)
		}
	}
}
