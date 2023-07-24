package memory

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// TokenBuffer for storing conversation memory.
type TokenBuffer struct {
	Buffer
	LLM           llms.LanguageModel
	MaxTokenLimit int
}

// Statically assert that TokenBuffer implement the memory interface.
var _ schema.Memory = &TokenBuffer{}

// NewTokenBuffer is a function for crating a new token buffer memory.
func NewTokenBuffer(llm llms.LanguageModel, maxTokenLimit int, options ...BufferOption) *TokenBuffer {
	tb := &TokenBuffer{
		LLM:           llm,
		MaxTokenLimit: maxTokenLimit,
		Buffer:        *applyBufferOptions(options...),
	}

	return tb
}

// MemoryVariables uses Buffer method for memory variables.
func (tb *TokenBuffer) MemoryVariables() []string {
	return tb.Buffer.MemoryVariables()
}

// LoadMemoryVariables uses Buffer method for loading memory variables.
func (tb *TokenBuffer) LoadMemoryVariables(inputs map[string]any) (map[string]any, error) {
	return tb.Buffer.LoadMemoryVariables(inputs)
}

// SaveContext uses Buffer method for saving context and prunes memory buffer if needed.
func (tb *TokenBuffer) SaveContext(inputValues map[string]any, outputValues map[string]any) error {
	err := tb.Buffer.SaveContext(inputValues, outputValues)
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
			tb.ChatHistory.SetMessages(append(tb.ChatHistory.Messages()[:0], tb.ChatHistory.Messages()[1:]...))
			currBufferLength, err = tb.getNumTokensFromMessages()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Clear uses Buffer method for clearing buffer memory.
func (tb *TokenBuffer) Clear() error {
	return tb.Buffer.Clear()
}

func (tb *TokenBuffer) getNumTokensFromMessages() (int, error) {
	sum := 0
	for _, message := range tb.ChatHistory.Messages() {
		bufferString, err := schema.GetBufferString([]schema.ChatMessage{message}, tb.Buffer.HumanPrefix, tb.Buffer.AIPrefix)
		if err != nil {
			return 0, err
		}

		sum += tb.LLM.GetNumTokens(bufferString)
	}

	return sum, nil
}
