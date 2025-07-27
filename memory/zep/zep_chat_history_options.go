package zep

import "github.com/getzep/zep-go/v2"

// ChatMessageHistoryOption is a function for creating new chat message history
// with other than the default values.
type ChatMessageHistoryOption func(m *ChatMessageHistory)

// WithChatHistoryMemoryType specifies zep memory type.
func WithChatHistoryMemoryType(memoryType zep.MemoryType) ChatMessageHistoryOption {
	return func(b *ChatMessageHistory) {
		b.MemoryType = memoryType
	}
}

// WithChatHistoryHumanPrefix is an option for specifying the human prefix. Will be passed as role for the message to zep.
func WithChatHistoryHumanPrefix(humanPrefix string) ChatMessageHistoryOption {
	return func(b *ChatMessageHistory) {
		b.HumanPrefix = humanPrefix
	}
}

// WithChatHistoryAIPrefix is an option for specifying the AI prefix. Will be passed as role for the message to zep.
func WithChatHistoryAIPrefix(aiPrefix string) ChatMessageHistoryOption {
	return func(b *ChatMessageHistory) {
		b.AIPrefix = aiPrefix
	}
}

func applyZepChatHistoryOptions(options ...ChatMessageHistoryOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		MemoryType: zep.MemoryTypePerpetual,
	}

	for _, option := range options {
		option(h)
	}

	return h
}
