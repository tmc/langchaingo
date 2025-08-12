package memory

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// EnhancedTokenBufferOption is a function that configures an EnhancedTokenBuffer.
type EnhancedTokenBufferOption func(*EnhancedTokenBuffer)

// WithTokenLimit sets the maximum number of tokens to keep in memory.
func WithTokenLimit(limit int) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.TokenLimit = limit
	}
}

// WithEncodingModel sets the model name used for token counting.
func WithEncodingModel(model string) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.EncodingModel = model
	}
}

// WithTokenCounter sets a custom token counter implementation.
func WithTokenCounter(counter TokenCounter) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.TokenCounter = counter
	}
}

// WithLLM sets the LLM to use for token counting (fallback option).
func WithLLM(llm llms.Model) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.LLM = llm
	}
}

// WithTrimStrategy sets the strategy used for trimming messages.
func WithTrimStrategy(strategy TrimStrategy) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.TrimStrategy = strategy
	}
}

// WithPreservePairs sets whether to preserve human-AI message pairs when trimming.
func WithPreservePairs(preserve bool) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.PreservePairs = preserve
	}
}

// WithMinMessages sets the minimum number of messages to preserve.
func WithMinMessages(min int) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.MinMessages = min
	}
}

// WithChatHistory sets the chat history implementation to use.
func WithChatHistory(history schema.ChatMessageHistory) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.ChatHistory = history
	}
}

// WithReturnMessages sets whether to return messages as a slice or as a string.
func WithReturnMessages(returnMessages bool) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.ReturnMessages = returnMessages
	}
}

// WithInputKey sets the key for input values in SaveContext.
func WithInputKey(key string) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.InputKey = key
	}
}

// WithOutputKey sets the key for output values in SaveContext.
func WithOutputKey(key string) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.OutputKey = key
	}
}

// WithHumanPrefix sets the prefix for human messages in string formatting.
func WithHumanPrefix(prefix string) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.HumanPrefix = prefix
	}
}

// WithAIPrefix sets the prefix for AI messages in string formatting.
func WithAIPrefix(prefix string) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.AIPrefix = prefix
	}
}

// WithMemoryKey sets the key used for storing memory in LoadMemoryVariables.
func WithMemoryKey(key string) EnhancedTokenBufferOption {
	return func(etb *EnhancedTokenBuffer) {
		etb.MemoryKey = key
	}
}

// applyEnhancedTokenBufferOptions applies the given options to create an EnhancedTokenBuffer.
func applyEnhancedTokenBufferOptions(options ...EnhancedTokenBufferOption) *EnhancedTokenBuffer {
	etb := &EnhancedTokenBuffer{
		// Default configuration
		ReturnMessages:  false,
		InputKey:        "input",
		OutputKey:       "output",
		HumanPrefix:     "Human",
		AIPrefix:        "AI",
		MemoryKey:       "history",
		TokenLimit:      2800, // Conservative default for GPT-3.5-turbo (4096 - ~1200 for response)
		EncodingModel:   "gpt-3.5-turbo",
		TrimStrategy:    TrimOldest,
		PreservePairs:   true,
		MinMessages:     2, // Keep at least one human-AI pair
		ChatHistory:     NewChatMessageHistory(),
	}

	// Apply all provided options
	for _, option := range options {
		option(etb)
	}

	return etb
}