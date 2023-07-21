package memory

import "github.com/tmc/langchaingo/schema"

// BufferOption is a function for creating new buffer
// with other then the default values.
type BufferOption func(b *Buffer)

// WithChatHistory is an option for providing the chat history store.
func WithChatHistory(chatHistory schema.ChatMessageHistory) BufferOption {
	return func(b *Buffer) {
		b.ChatHistory = chatHistory
	}
}

// WithReturnMessages is an option for specifying should it return messages.
func WithReturnMessages(returnMessages bool) BufferOption {
	return func(b *Buffer) {
		b.ReturnMessages = returnMessages
	}
}

// WithInputKey is an option for specifying the input key.
func WithInputKey(inputKey string) BufferOption {
	return func(b *Buffer) {
		b.InputKey = inputKey
	}
}

// WithOutputKey is an option for specifying the output key.
func WithOutputKey(outputKey string) BufferOption {
	return func(b *Buffer) {
		b.OutputKey = outputKey
	}
}

// WithHumanPrefix is an option for specifying the human prefix.
func WithHumanPrefix(humanPrefix string) BufferOption {
	return func(b *Buffer) {
		b.HumanPrefix = humanPrefix
	}
}

// WithAIPrefix is an option for specifying the AI prefix.
func WithAIPrefix(aiPrefix string) BufferOption {
	return func(b *Buffer) {
		b.AIPrefix = aiPrefix
	}
}

// WithMemoryKey is an option for specifying the memory key.
func WithMemoryKey(memoryKey string) BufferOption {
	return func(b *Buffer) {
		b.MemoryKey = memoryKey
	}
}

func applyBufferOptions(opts ...BufferOption) *Buffer {
	m := &Buffer{
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
