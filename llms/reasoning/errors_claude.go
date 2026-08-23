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
