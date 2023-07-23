package schema

// ChatMessageHistory is the interface for chat history in memory/store.
type ChatMessageHistory interface {
	// AddUserMessage Convenience method for adding a human message string to the store.
	AddUserMessage(message string)

	// AddAIMessage Convenience method for adding an AI message string to the store.
	AddAIMessage(message string)

	// AddMessage Add a Message object to the store.
	AddMessage(message ChatMessage)

	// Clear Remove all messages from the store.
	Clear()

	// Messages get all messages from the store
	Messages() []ChatMessage
}
