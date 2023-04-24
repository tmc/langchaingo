package prompts

import (
	"github.com/tmc/langchaingo/schema"
)

var _ schema.PromptValue = StringPromptValue{}

// StringPromptValue is a prompt value that is a string.
type StringPromptValue struct {
	value string
}

func (v StringPromptValue) String() string {
	return v.value
}

func (v StringPromptValue) Messages() []schema.ChatMessage {
	return []schema.ChatMessage{
		schema.HumanChatMessage{Text: v.value},
	}
}

// NewStringPromptValue creates a new StringPromptValue.
func NewStringPromptValue(value string) StringPromptValue {
	return StringPromptValue{value: value}
}
