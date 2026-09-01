package reasoning

import "testing"

func TestOpenAIProVariantsReasonWithoutBeingAsked(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gpt-5-pro",
		"gpt-5.2-pro",
		"gpt-5.4-pro",
		"gpt-5.5-pro",
		"openai/gpt-5.2-pro",
		"gpt-5.2-pro-2025-12-11",
	} {
		if OpenAIThinkingOptIn(model) {
			t.Errorf("OpenAIThinkingOptIn(%q) = true, but a bare call already returns reasoning", model)
		}
	}
}

func TestOpenAIOptInGenerationsWaitToBeAsked(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gpt-5.1",
		"gpt-5.2",
		"gpt-5.4",
		"openai/gpt-5.4",
		"gpt-5.4-mini",
		"gpt-5.2-2025-12-11",
	} {
		if !OpenAIThinkingOptIn(model) {
			t.Errorf("OpenAIThinkingOptIn(%q) = false, but a bare call returns no reasoning", model)
		}
	}
}

func TestOpenAIGenerationsThatAlwaysReasonAreNotOptIn(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gpt-5",
		"gpt-5-mini",
		"gpt-5-nano",
		"gpt-5.5",
		"gpt-5.6",
		"gpt-5.6-terra",
	} {
		if OpenAIThinkingOptIn(model) {
			t.Errorf("OpenAIThinkingOptIn(%q) = true, but a bare call already returns reasoning", model)
		}
	}
}

func TestOptInModelsCanAlsoDisableReasoning(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"gpt-5", "gpt-5-mini", "gpt-5-nano", "gpt-5-pro",
		"gpt-5.1", "gpt-5.2", "gpt-5.2-pro", "gpt-5.4", "gpt-5.4-pro",
		"gpt-5.5", "gpt-5.5-pro", "gpt-5.6", "gpt-5.6-terra", "gpt-5.6-luna",
	} {
		if !OpenAIThinkingOptIn(model) {
			continue
		}
		if caps := OpenAIReasoningCapsFor(model); !caps.CanDisable {
			t.Errorf("OpenAIThinkingOptIn(%q) = true while CanDisable = false: a model that cannot be "+
				"turned off does not wait to be asked", model)
		}
	}
}
