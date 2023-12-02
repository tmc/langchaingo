package memory

import (
	"context" // nolint:goimports,gofumpt,gci
	"github.com/tmc/langchaingo/schema"
)

const DefaultMultiple = 2

// ConversationBufferWindow is a simple form of memory that remembers previous conversational back and forths directly.
type ConversationBufferWindow struct {
	ChatHistory schema.ChatMessageHistory

	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AIPrefix       string
	MemoryKey      string
	K              int
}

// Statically assert that ConversationWindowBuffer implement the memory interface.
var _ schema.Memory = &ConversationBufferWindow{}

// NewConversationWindowBuffer is a function for crating a new buffer memory.
func NewConversationWindowBuffer(options ...ConversationBufferWindowOption) *ConversationBufferWindow {
	return applyBufferWindowOptions(options...)
}

func (c ConversationBufferWindow) GetMemoryKey(_ context.Context) string {
	return c.MemoryKey
}

func (c ConversationBufferWindow) MemoryVariables(_ context.Context) []string {
	return []string{c.MemoryKey}
}

func (c ConversationBufferWindow) LoadMemoryVariables(
	ctx context.Context,
	_ map[string]any) (map[string]any, error) { // nolint:gofumpt
	messages, err := c.ChatHistory.Messages(ctx)
	if err != nil {
		return nil, err
	}
	messageLength := len(messages)
	start := 0
	if messageLength > c.K*DefaultMultiple {
		start = messageLength - c.K*DefaultMultiple
	}
	newMessages := messages[start:]

	if c.ReturnMessages {
		return map[string]any{
			c.MemoryKey: newMessages,
		}, nil
	}

	bufferString, err := schema.GetBufferString(newMessages, c.HumanPrefix, c.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		c.MemoryKey: bufferString,
	}, nil
}

func (c ConversationBufferWindow) SaveContext(
	ctx context.Context,
	inputs map[string]any,
	outputs map[string]any) error { // nolint:gofumpt
	userInputValue, err := getInputValue(inputs, c.InputKey)
	if err != nil {
		return err
	}
	err = c.ChatHistory.AddUserMessage(ctx, userInputValue)
	if err != nil {
		return err
	}

	aiOutputValue, err := getInputValue(outputs, c.OutputKey)
	if err != nil {
		return err
	}
	err = c.ChatHistory.AddAIMessage(ctx, aiOutputValue)
	if err != nil {
		return err
	}

	return nil
}

func (c ConversationBufferWindow) Clear(ctx context.Context) error {
	return c.ChatHistory.Clear(ctx)
}
