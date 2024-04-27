package memory

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	// defaultConversationWindowSize is the default number of previous conversation.
	defaultConversationWindowSize = 5
	// defaultMessageSize indicates the length of a complete message, currently consisting of 2 parts: ai and human.
	defaultMessageSize = 2
)

// ConversationWindowBuffer for storing conversation memory.
type ConversationWindowBuffer struct {
	ConversationBuffer
	ConversationWindowSize int
}

// Statically assert that ConversationWindowBuffer implement the memory interface.
var _ schema.Memory = &ConversationWindowBuffer{}

// NewConversationWindowBuffer is a function for crating a new window buffer memory.
func NewConversationWindowBuffer(
	conversationWindowSize int,
	options ...ConversationBufferOption,
) *ConversationWindowBuffer {
	if conversationWindowSize <= 0 {
		conversationWindowSize = defaultConversationWindowSize
	}
	tb := &ConversationWindowBuffer{
		ConversationWindowSize: conversationWindowSize,
		ConversationBuffer:     *applyBufferOptions(options...),
	}

	return tb
}

// MemoryVariables uses ConversationBuffer method for memory variables.
func (wb *ConversationWindowBuffer) MemoryVariables(ctx context.Context) []string {
	return wb.ConversationBuffer.MemoryVariables(ctx)
}

// LoadMemoryVariables uses ConversationBuffer method for loading memory variables.
func (wb *ConversationWindowBuffer) LoadMemoryVariables(ctx context.Context, _ map[string]any) (map[string]any, error) {
	messages, err := wb.ChatHistory.Messages(ctx)
	if err != nil {
		return nil, err
	}
	messages, _ = wb.cutMessages(messages)

	if wb.ReturnMessages {
		return map[string]any{
			wb.MemoryKey: messages,
		}, nil
	}

	bufferString, err := llms.GetBufferString(messages, wb.HumanPrefix, wb.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		wb.MemoryKey: bufferString,
	}, nil
}

// SaveContext uses ConversationBuffer method for saving context and prunes memory buffer if needed.
func (wb *ConversationWindowBuffer) SaveContext(
	ctx context.Context, inputValues map[string]any, outputValues map[string]any,
) error {
	err := wb.ConversationBuffer.SaveContext(ctx, inputValues, outputValues)
	if err != nil {
		return err
	}
	messages, err := wb.ConversationBuffer.ChatHistory.Messages(ctx)
	if err != nil {
		return err
	}
	if messages, ok := wb.cutMessages(messages); ok {
		err := wb.ConversationBuffer.ChatHistory.SetMessages(ctx, messages)
		if err != nil {
			return err
		}
	}
	return nil
}

func (wb *ConversationWindowBuffer) cutMessages(message []llms.ChatMessage) ([]llms.ChatMessage, bool) {
	if len(message) > wb.ConversationWindowSize*defaultMessageSize {
		return message[len(message)-wb.ConversationWindowSize*defaultMessageSize:], true
	}
	return message, false
}

// Clear uses ConversationBuffer method for clearing buffer memory.
func (wb *ConversationWindowBuffer) Clear(ctx context.Context) error {
	return wb.ConversationBuffer.Clear(ctx)
}
