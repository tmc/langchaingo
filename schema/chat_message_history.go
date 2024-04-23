package schema

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// ChatMessageHistory is the interface for chat history in memory/store.
type ChatMessageHistory interface {
	// AddMessage adds a message to the store.
	AddMessage(ctx context.Context, message llms.ChatMessage) error

	// AddUserMessage is a convenience method for adding a human message string
	// to the store.
	AddUserMessage(ctx context.Context, message string) error

	// AddAIMessage is a convenience method for adding an AI message string to
	// the store.
	AddAIMessage(ctx context.Context, message string) error

	// Clear removes all messages from the store.
	Clear(ctx context.Context) error

	// Messages retrieves all messages from the store
	Messages(ctx context.Context) ([]llms.ChatMessage, error)

	// SetMessages replaces existing messages in the store
	SetMessages(ctx context.Context, messages []llms.ChatMessage) error
}
