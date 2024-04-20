package memory

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// ConversationTokenBuffer for storing conversation memory.
type ConversationTokenBuffer struct {
	ConversationBuffer
	LLM           llms.Model
	MaxTokenLimit int
}

// Statically assert that ConversationTokenBuffer implement the memory interface.
var _ schema.Memory = &ConversationTokenBuffer{}

// NewConversationTokenBuffer is a function for crating a new token buffer memory.
func NewConversationTokenBuffer(
	llm llms.Model,
	maxTokenLimit int,
	options ...ConversationBufferOption,
) *ConversationTokenBuffer {
	tb := &ConversationTokenBuffer{
		LLM:                llm,
		MaxTokenLimit:      maxTokenLimit,
		ConversationBuffer: *applyBufferOptions(options...),
	}

	return tb
}

// MemoryVariables uses ConversationBuffer method for memory variables.
func (tb *ConversationTokenBuffer) MemoryVariables(ctx context.Context) []string {
	return tb.ConversationBuffer.MemoryVariables(ctx)
}

// LoadMemoryVariables uses ConversationBuffer method for loading memory variables.
func (tb *ConversationTokenBuffer) LoadMemoryVariables(
	ctx context.Context, inputs map[string]any,
) (map[string]any, error) {
	return tb.ConversationBuffer.LoadMemoryVariables(ctx, inputs)
}

// SaveContext uses ConversationBuffer method for saving context and prunes memory buffer if needed.
func (tb *ConversationTokenBuffer) SaveContext(
	ctx context.Context, inputValues map[string]any, outputValues map[string]any,
) error {
	err := tb.ConversationBuffer.SaveContext(ctx, inputValues, outputValues)
	if err != nil {
		return err
	}
	currBufferLength, err := tb.getNumTokensFromMessages(ctx)
	if err != nil {
		return err
	}

	if currBufferLength > tb.MaxTokenLimit {
		// while currBufferLength is greater than MaxTokenLimit we keep removing messages from the memory
		// from the oldest
		for currBufferLength > tb.MaxTokenLimit {
			messages, err := tb.ChatHistory.Messages(ctx)
			if err != nil {
				return err
			}

			if len(messages) == 0 {
				break
			}

			err = tb.ChatHistory.SetMessages(ctx, append(messages[:0], messages[1:]...))
			if err != nil {
				return err
			}

			currBufferLength, err = tb.getNumTokensFromMessages(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Clear uses ConversationBuffer method for clearing buffer memory.
func (tb *ConversationTokenBuffer) Clear(ctx context.Context) error {
	return tb.ConversationBuffer.Clear(ctx)
}

func (tb *ConversationTokenBuffer) getNumTokensFromMessages(ctx context.Context) (int, error) {
	messages, err := tb.ChatHistory.Messages(ctx)
	if err != nil {
		return 0, err
	}

	bufferString, err := llms.GetBufferString(
		messages,
		tb.ConversationBuffer.HumanPrefix,
		tb.ConversationBuffer.AIPrefix,
	)
	if err != nil {
		return 0, err
	}

	return llms.CountTokens("", bufferString), nil
}
