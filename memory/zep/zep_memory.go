package zep

import (
	"context"

	"github.com/getzep/zep-go"
	zepClient "github.com/getzep/zep-go/client"
	"github.com/starmvp/langchaingo/llms"
	"github.com/starmvp/langchaingo/memory"
	"github.com/starmvp/langchaingo/schema"
)

// Memory is a simple form of memory that remembers previous conversational back and forth directly.
type Memory struct {
	ChatHistory    schema.ChatMessageHistory
	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AIPrefix       string
	MemoryKey      string
	MemoryType     zep.MemoryGetRequestMemoryType
	ZepClient      *zepClient.Client
	SessionID      string
}

// Statically assert that ZepMemory implement the memory interface.
var _ schema.Memory = &Memory{}

// NewMemory is a function for crating a new buffer memory.
func NewMemory(client *zepClient.Client, sessionID string, options ...MemoryOption) *Memory {
	m := applyZepMemoryOptions(options...)
	m.ZepClient = client
	m.SessionID = sessionID
	m.ChatHistory = NewZepChatMessageHistory(
		m.ZepClient,
		m.SessionID,
		WithChatHistoryMemoryType(m.MemoryType),
		WithChatHistoryHumanPrefix(m.HumanPrefix),
		WithChatHistoryAIPrefix(m.AIPrefix),
	)
	return m
}

// MemoryVariables gets the input key the buffer memory class will load dynamically.
func (m *Memory) MemoryVariables(context.Context) []string {
	return []string{m.MemoryKey}
}

// LoadMemoryVariables returns the previous chat messages stored in memory
// as well as a system message with conversation facts and most relevant summary.
// Previous chat messages are returned in a map with the key specified in the MemoryKey field. This key defaults to
// "history". If ReturnMessages is set to true the output is a slice of schema.ChatMessage. Otherwise,
// the output is a buffer string of the chat messages.
func (m *Memory) LoadMemoryVariables(
	ctx context.Context, _ map[string]any,
) (map[string]any, error) {
	messages, err := m.ChatHistory.Messages(ctx)
	if err != nil {
		return nil, err
	}

	if m.ReturnMessages {
		return map[string]any{
			m.MemoryKey: messages,
		}, nil
	}

	bufferString, err := llms.GetBufferString(messages, m.HumanPrefix, m.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		m.MemoryKey: bufferString,
	}, nil
}

// SaveContext uses the input values to the llm to save a user message, and the output values
// of the llm to save an AI message. If the input or output key is not set, the input values or
// output values must contain only one key such that the function can know what string to
// add as a user and AI message. On the other hand, if the output key or input key is set, the
// input key must be a key in the input values and the output key must be a key in the output
// values. The values in the input and output values used to save a user and AI message must
// be strings.
func (m *Memory) SaveContext(
	ctx context.Context,
	inputValues map[string]any,
	outputValues map[string]any,
) error {
	userInputValue, err := memory.GetInputValue(inputValues, m.InputKey)
	if err != nil {
		return err
	}
	err = m.ChatHistory.AddUserMessage(ctx, userInputValue)
	if err != nil {
		return err
	}

	aiOutputValue, err := memory.GetInputValue(outputValues, m.OutputKey)
	if err != nil {
		return err
	}
	err = m.ChatHistory.AddAIMessage(ctx, aiOutputValue)
	if err != nil {
		return err
	}

	return nil
}

// Clear sets the chat messages to a new and empty chat message history.
func (m *Memory) Clear(ctx context.Context) error {
	return m.ChatHistory.Clear(ctx)
}

func (m *Memory) GetMemoryKey(context.Context) string {
	return m.MemoryKey
}
