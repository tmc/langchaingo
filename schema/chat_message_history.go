package schema

// ChatMessageHistory is the interface for chat history in memory/store.
type ChatMessageHistory interface {
	// AddUserMessage Convenience method for adding a human message string to the store.
	AddUserMessage(message string) error

	// AddAIMessage Convenience method for adding an AI message string to the store.
	AddAIMessage(message string) error

	// AddMessage Add a Message object to the store.
	AddMessage(message ChatMessage) error

	// Clear Remove all messages from the store.
	Clear() error

	// Messages get all messages from the store
	Messages() ([]ChatMessage, error)

	// SetMessages replaces existing messages in the store
	SetMessages(messages []ChatMessage) error
}
