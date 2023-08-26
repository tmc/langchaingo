package schema

import "context"

// ChatMessageHistory is the interface for chat history in memory/store.
type ChatMessageHistory interface {
	// AddUserMessage Convenience method for adding a human message string to the store.
	AddUserMessage(ctx context.Context, message string) error

	// AddAIMessage Convenience method for adding an AI message string to the store.
	AddAIMessage(ctx context.Context, message string) error

	// AddMessage Add a Message object to the store.
	AddMessage(ctx context.Context, message ChatMessage) error

	// Clear Remove all messages from the store.
	Clear(ctx context.Context) error

	// Messages get all messages from the store
	Messages(ctx context.Context) ([]ChatMessage, error)

	// SetMessages replaces existing messages in the store
	SetMessages(ctx context.Context, messages []ChatMessage) error
}
