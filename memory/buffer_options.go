package memory

import "github.com/tmc/langchaingo/schema"

// ConversationBufferOption is a function for creating new buffer
// with other than the default values.
type ConversationBufferOption func(b *ConversationBuffer)

// WithChatHistory is an option for providing the chat history store.
func WithChatHistory(chatHistory schema.ChatMessageHistory) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.ChatHistory = chatHistory
	}
}

// WithReturnMessages is an option for specifying should it return messages.
func WithReturnMessages(returnMessages bool) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.ReturnMessages = returnMessages
	}
}

// WithInputKey is an option for specifying the input key.
func WithInputKey(inputKey string) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.InputKey = inputKey
	}
}

// WithOutputKey is an option for specifying the output key.
func WithOutputKey(outputKey string) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.OutputKey = outputKey
	}
}

// WithHumanPrefix is an option for specifying the human prefix.
func WithHumanPrefix(humanPrefix string) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.HumanPrefix = humanPrefix
	}
}

// WithAIPrefix is an option for specifying the AI prefix.
func WithAIPrefix(aiPrefix string) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.AIPrefix = aiPrefix
	}
}

// WithMemoryKey is an option for specifying the memory key.
func WithMemoryKey(memoryKey string) ConversationBufferOption {
	return func(b *ConversationBuffer) {
		b.MemoryKey = memoryKey
	}
}

func applyBufferOptions(opts ...ConversationBufferOption) *ConversationBuffer {
	m := &ConversationBuffer{
		ReturnMessages: false,
		InputKey:       "",
		OutputKey:      "",
		HumanPrefix:    "Human",
		AIPrefix:       "AI",
		MemoryKey:      "history",
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.ChatHistory == nil {
		m.ChatHistory = NewChatMessageHistory()
	}

	return m
}
