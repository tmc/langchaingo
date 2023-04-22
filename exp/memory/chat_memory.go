package memory

import "github.com/tmc/langchaingo/schema"

type ChatMessageHistory struct {
	messages []schema.ChatMessage
}

func (h *ChatMessageHistory) GetMessages() []schema.ChatMessage { return h.messages }
func (h *ChatMessageHistory) AddAiMessage(text string) {
	h.messages = append(h.messages, schema.AIChatMessage{Text: text})
}

func (h *ChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, schema.HumanChatMessage{Text: text})
}

func NewChatMessageHistory(options ...NewChatMessageOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
	}

	for _, option := range options {
		option(h)
	}

	return h
}

type NewChatMessageOption func(m *ChatMessageHistory)

// Option for NewChatMessageHistory adding previous messages to the history.
func WithPreviousMessages(previousMessages []schema.ChatMessage) NewChatMessageOption {
	return func(m *ChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
