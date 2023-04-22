package memory

import "github.com/tmc/langchaingo/schema"

// ChatMessageHistory stores chat messages.
type ChatMessageHistory struct {
	messages []schema.ChatMessage
}

// Creates new ChatMessageHistory.
func NewChatMessageHistory(options ...NewChatMessageOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
	}

	for _, option := range options {
		option(h)
	}

	return h
}

// Returns all messages stored.
func (h *ChatMessageHistory) Messages() []schema.ChatMessage {
	return h.messages
}

// Adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(text string) {
	h.messages = append(h.messages, schema.AIChatMessage{Text: text})
}

// Adds an user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, schema.HumanChatMessage{Text: text})
}

// Option for creating new chat message history.
type NewChatMessageOption func(m *ChatMessageHistory)

// Option for NewChatMessageHistory adding previous messages to the history.
func WithPreviousMessages(previousMessages []schema.ChatMessage) NewChatMessageOption {
	return func(m *ChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
