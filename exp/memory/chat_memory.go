package memory

import "github.com/tmc/langchaingo/schema"

// ChatMessageHistory stores chat messages
type ChatMessageHistory struct {
	messages []schema.ChatMessage
}

// NewChatMessageHistory creates new ChatMessageHistory with options.
func NewChatMessageHistory(options ...NewChatMessageOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
	}

	for _, option := range options {
		option(h)
	}

	return h
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages() []schema.ChatMessage {
	return h.messages
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(text string) {
	h.messages = append(h.messages, schema.AIChatMessage{Text: text})
}

// AddUserMessage adds an user to the chat message history
func (h *ChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, schema.HumanChatMessage{Text: text})
}

// NewChatMessageOption is a function for creating new chat message history
// with other then the default values.
type NewChatMessageOption func(m *ChatMessageHistory)

// WithPreviousMessages is an option for NewChatMessageHistory for adding
// previous messages to the history.
func WithPreviousMessages(previousMessages []schema.ChatMessage) NewChatMessageOption {
	return func(m *ChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
