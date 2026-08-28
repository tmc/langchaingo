package reasoning

import "fmt"

// ErrAssistantPrefillUnsupported reports that the model rejects a conversation
// ending with an assistant turn.
type ErrAssistantPrefillUnsupported struct{ Model string }

func (e *ErrAssistantPrefillUnsupported) Error() string {
	return fmt.Sprintf(
		"model %q does not support assistant message prefill; the conversation must end with a user message",
		e.Model)
}

// ErrForcedToolUseWithThinking reports that a forced tool choice was combined
// with manual (budget) thinking.
type ErrForcedToolUseWithThinking struct{ Model string }

func (e *ErrForcedToolUseWithThinking) Error() string {
	return fmt.Sprintf(
		"model %q runs manual thinking, which rejects a forced tool choice; use tool_choice auto or none",
		e.Model)
}

// ErrEffortHasNoBudget reports that reasoning was asked for with an effort that
// does not map to a token budget, on a model whose only thinking wire is one.
type ErrEffortHasNoBudget struct {
	Model  string
	Effort string
}

func (e *ErrEffortHasNoBudget) Error() string {
	return fmt.Sprintf(
		"model %q thinks by token budget, and effort %q maps to none; pass an explicit budget or an effort from low to max",
		e.Model, e.Effort)
}
