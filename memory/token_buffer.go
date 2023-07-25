package memory

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// ConversationTokenBuffer for storing conversation memory.
type ConversationTokenBuffer struct {
	ConversationBuffer
	LLM           llms.LanguageModel
	MaxTokenLimit int
}

// Statically assert that ConversationTokenBuffer implement the memory interface.
var _ schema.Memory = &ConversationTokenBuffer{}

// NewConversationTokenBuffer is a function for crating a new token buffer memory.
func NewConversationTokenBuffer(
	llm llms.LanguageModel,
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
func (tb *ConversationTokenBuffer) MemoryVariables() []string {
	return tb.ConversationBuffer.MemoryVariables()
}

// LoadMemoryVariables uses ConversationBuffer method for loading memory variables.
func (tb *ConversationTokenBuffer) LoadMemoryVariables(inputs map[string]any) (map[string]any, error) {
	return tb.ConversationBuffer.LoadMemoryVariables(inputs)
}

// SaveContext uses ConversationBuffer method for saving context and prunes memory buffer if needed.
func (tb *ConversationTokenBuffer) SaveContext(inputValues map[string]any, outputValues map[string]any) error {
	err := tb.ConversationBuffer.SaveContext(inputValues, outputValues)
	if err != nil {
		return err
	}
	currBufferLength, err := tb.getNumTokensFromMessages()
	if err != nil {
		return err
	}

	if currBufferLength > tb.MaxTokenLimit {
		// while currBufferLength is greater than MaxTokenLimit we keep removing messages from the memory
		// from the oldest
		for currBufferLength > tb.MaxTokenLimit {
			if len(tb.ChatHistory.Messages()) == 0 {
				break
			}

			tb.ChatHistory.SetMessages(append(tb.ChatHistory.Messages()[:0], tb.ChatHistory.Messages()[1:]...))
			currBufferLength, err = tb.getNumTokensFromMessages()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Clear uses ConversationBuffer method for clearing buffer memory.
func (tb *ConversationTokenBuffer) Clear() error {
	return tb.ConversationBuffer.Clear()
}

func (tb *ConversationTokenBuffer) getNumTokensFromMessages() (int, error) {
	bufferString, err := schema.GetBufferString(
		tb.ChatHistory.Messages(),
		tb.ConversationBuffer.HumanPrefix,
		tb.ConversationBuffer.AIPrefix,
	)
	if err != nil {
		return 0, err
	}

	return tb.LLM.GetNumTokens(bufferString), nil
}
