package memory

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

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
func (h *ChatMessageHistory) Messages(_ context.Context) ([]schema.ChatMessage, error) {
	return h.messages, nil
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(_ context.Context, text string) error {
	h.messages = append(h.messages, schema.AIChatMessage{Content: text})
	return nil
}

// AddUserMessage adds an user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(_ context.Context, text string) error {
	h.messages = append(h.messages, schema.HumanChatMessage{Content: text})
	return nil
}

func (h *ChatMessageHistory) Clear() error {
	h.messages = make([]schema.ChatMessage, 0)
	return nil
}

func (h *ChatMessageHistory) AddMessage(_ context.Context, message schema.ChatMessage) error {
	h.messages = append(h.messages, message)
	return nil
}

func (h *ChatMessageHistory) SetMessages(_ context.Context, messages []schema.ChatMessage) error {
	h.messages = messages
	return nil
}
