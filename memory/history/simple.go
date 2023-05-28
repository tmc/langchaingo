package history

import "github.com/tmc/langchaingo/schema"

// Statically assert that the types implement the interface.
var _ schema.ChatMessageHistory = (*SimpleChatMessageHistory)(nil)

// SimpleChatMessageHistory is a struct that stores chat messages.
type SimpleChatMessageHistory struct {
	messages []schema.ChatMessage
}

// NewSimpleChatMessageHistory creates a new SimpleChatMessageHistory using chat message options.
func NewSimpleChatMessageHistory(options ...NewChatMessageOption) *SimpleChatMessageHistory {
	h := &SimpleChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
	}

	for _, option := range options {
		option(h)
	}

	return h
}

// Messages returns all messages stored.
func (h *SimpleChatMessageHistory) Messages() []schema.ChatMessage {
	return h.messages
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *SimpleChatMessageHistory) AddAIMessage(text string) {
	h.messages = append(h.messages, schema.AIChatMessage{Text: text})
}

// AddUserMessage adds an user to the chat message history.
func (h *SimpleChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, schema.HumanChatMessage{Text: text})
}

func (h *SimpleChatMessageHistory) Clear() {
	h.messages = make([]schema.ChatMessage, 0)
}

// NewChatMessageOption is a function for creating new chat message history
// with other then the default values.
type NewChatMessageOption func(m *SimpleChatMessageHistory)

// WithPreviousMessages is an option for NewSimpleChatMessageHistory for adding
// previous messages to the history.
func WithPreviousMessages(previousMessages []schema.ChatMessage) NewChatMessageOption {
	return func(m *SimpleChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
