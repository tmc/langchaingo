package memory

import "github.com/tmc/langchaingo/schema"

// ConversationBufferWindowOption is a function for creating new buffer
// with other then the default values.
type ConversationBufferWindowOption func(b *ConversationBufferWindow)

// WithBufferWindowChatHistory is an option for providing the chat history store.
func WithBufferWindowChatHistory(chatHistory schema.ChatMessageHistory) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.ChatHistory = chatHistory
	}
}

// WithBufferWindowReturnMessages is an option for specifying should it return messages.
func WithBufferWindowReturnMessages(returnMessages bool) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.ReturnMessages = returnMessages
	}
}

// WithBufferWindowInputKey is an option for specifying the input key.
func WithBufferWindowInputKey(inputKey string) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.InputKey = inputKey
	}
}

// WithBufferWindowOutputKey is an option for specifying the output key.
func WithBufferWindowOutputKey(outputKey string) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.OutputKey = outputKey
	}
}

// WithBufferWindowHumanPrefix is an option for specifying the human prefix.
func WithBufferWindowHumanPrefix(humanPrefix string) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.HumanPrefix = humanPrefix
	}
}

// WithBufferWindowAIPrefix is an option for specifying the AI prefix.
func WithBufferWindowAIPrefix(aiPrefix string) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.AIPrefix = aiPrefix
	}
}

// WithBufferWindowMemoryKey is an option for specifying the memory key.
func WithBufferWindowMemoryKey(memoryKey string) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.MemoryKey = memoryKey
	}
}

// WithBufferWindowK is an option for specifying the memory key.
func WithBufferWindowK(k int) ConversationBufferWindowOption {
	return func(b *ConversationBufferWindow) {
		b.K = k
	}
}

func applyBufferWindowOptions(opts ...ConversationBufferWindowOption) *ConversationBufferWindow {
	m := &ConversationBufferWindow{
		ReturnMessages: false,
		InputKey:       "",
		OutputKey:      "",
		HumanPrefix:    "Human",
		AIPrefix:       "AI",
		MemoryKey:      "history",
		K:              5,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.ChatHistory == nil {
		m.ChatHistory = NewChatMessageHistory()
	}

	return m
}
