package prompts

import (
	"fmt"

	"github.com/tmc/langchaingo/schema"
)

var _ schema.PromptValue = ChatPromptValue{}

// ChatPromptValue is a prompt value that is a list of chat messages.
type ChatPromptValue []schema.ChatMessage

// String returns the chat message slice as a buffer string.
func (v ChatPromptValue) String() string {
	s, err := schema.GetBufferString(v, "Human", "AI")
	if err == nil {
		return s
	}
	return fmt.Sprintf("%v", []schema.ChatMessage(v))
}

// Messages returns the ChatMessage slice.
func (v ChatPromptValue) Messages() []schema.ChatMessage {
	return []schema.ChatMessage(v)
}
