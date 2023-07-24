package prompts

import (
	"github.com/tmc/langchaingo/schema"
)

var _ schema.PromptValue = StringPromptValue("")

// StringPromptValue is a prompt value that is a string.
type StringPromptValue string

func (v StringPromptValue) String() string {
	return string(v)
}

// Messages returns a single-element ChatMessage slice.
func (v StringPromptValue) Messages() []schema.ChatMessage {
	return []schema.ChatMessage{
		schema.HumanChatMessage{Content: string(v)},
	}
}
