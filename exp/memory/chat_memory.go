package memory

import "github.com/tmc/langchaingo/schema"

// ChatMessageHistory stores chat messages.
type ChatMessageHistory struct {
	messages []schema.ChatMessage
}

// NewChatMessageHistory creates a new ChatMessageHistory using chat message options.
func NewChatMessageHistory(options ...NewChatMessageOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
	}

	for _, option := range options {
		option(h)
	}

	return h
}

// Messages is a function that returns all messages stored.
func (h *ChatMessageHistory) Messages() []schema.ChatMessage {
	return h.messages
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(text string) {
	h.messages = append(h.messages, schema.AIChatMessage{Text: text})
}

// AddUserMessage adds an user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, schema.HumanChatMessage{Text: text})
}

// NewChatMessageOption is a function type that can be used when creating a new chat message history.
type NewChatMessageOption func(m *ChatMessageHistory)

// WithPreviousMessages is a function that sets previous messages to the history.
func WithPreviousMessages(previousMessages []schema.ChatMessage) NewChatMessageOption {
	return func(m *ChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
