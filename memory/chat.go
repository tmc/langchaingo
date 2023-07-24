package memory

import "github.com/tmc/langchaingo/schema"

// ChatMessageHistory is a struct that stores chat messages.
type ChatMessageHistory struct {
	messages []schema.ChatMessage
}

// Statically assert that ChatMessageHistory implement the chat message history interface.
var _ schema.ChatMessageHistory = &ChatMessageHistory{}

// NewChatMessageHistory creates a new ChatMessageHistory using chat message options.
func NewChatMessageHistory(options ...ChatMessageHistoryOption) *ChatMessageHistory {
	return applyChatOptions(options...)
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages() []schema.ChatMessage {
	return h.messages
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(text string) {
	h.messages = append(h.messages, schema.AIChatMessage{Content: text})
}

// AddUserMessage adds an user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, schema.HumanChatMessage{Content: text})
}

func (h *ChatMessageHistory) Clear() {
	h.messages = make([]schema.ChatMessage, 0)
}

func (h *ChatMessageHistory) AddMessage(message schema.ChatMessage) {
	h.messages = append(h.messages, message)
}

func (h *ChatMessageHistory) SetMessages(messages []schema.ChatMessage) {
	h.messages = messages
}
