package history

import "github.com/tmc/langchaingo/schema"

// Statically assert that the types implement the interface.
var _ schema.ChatMessageHistory = (*SimpleChatMessageHistory)(nil)

// SimpleChatMessageHistory is a struct that stores chat messages.
type SimpleChatMessageHistory struct {
	messages []schema.ChatMessage
}

// NewSimpleChatMessageHistory creates a new SimpleChatMessageHistory using chat message options.
func NewSimpleChatMessageHistory(options ...NewSimpleChatMessageOption) *SimpleChatMessageHistory {
	h := &SimpleChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
	}
	for _, option := range options {
		option(h)
	}
	return h
}

// Messages returns all messages stored.
func (h *SimpleChatMessageHistory) Messages() ([]schema.ChatMessage, error) {
	return h.messages, nil
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *SimpleChatMessageHistory) AddMessage(msg schema.ChatMessage) error {
	h.messages = append(h.messages, msg)
	return nil
}

func (h *SimpleChatMessageHistory) Clear() error {
	h.messages = make([]schema.ChatMessage, 0)
	return nil
}

// NewSimpleChatMessageOption is a function for creating new chat message history
// with other then the default values.
type NewSimpleChatMessageOption func(m *SimpleChatMessageHistory)

// WithPreviousMessages is an option for NewSimpleChatMessageHistory for adding
// previous messages to the history.
func WithPreviousMessages(previousMessages []schema.ChatMessage) NewSimpleChatMessageOption {
	return func(m *SimpleChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
