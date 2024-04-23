package prompts

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

var _ llms.PromptValue = ChatPromptValue{}

// ChatPromptValue is a prompt value that is a list of chat messages.
type ChatPromptValue []llms.ChatMessage

// String returns the chat message slice as a buffer string.
func (v ChatPromptValue) String() string {
	s, err := llms.GetBufferString(v, "Human", "AI")
	if err == nil {
		return s
	}
	return fmt.Sprintf("%v", []llms.ChatMessage(v))
}

// Messages returns the ChatMessage slice.
func (v ChatPromptValue) Messages() []llms.ChatMessage {
	return v
}
