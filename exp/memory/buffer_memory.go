package memory

import (
	"fmt"

	"github.com/tmc/langchaingo/schema"
)

// Buffer is a simple form of memory that remembers previous conversational back and forths directly.
type Buffer struct {
	ChatHistory *ChatMessageHistory

	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AiPrefix       string
	MemoryKey      string
}

// Statically assert that Buffer implement the memory interface:
var _ schema.Memory = &Buffer{}

// NewBuffer is a function for crating a new buffer memory.
func NewBuffer() *Buffer {
	m := Buffer{
		ChatHistory:    NewChatMessageHistory(),
		ReturnMessages: false,
		InputKey:       "",
		OutputKey:      "",
		HumanPrefix:    "Human",
		AiPrefix:       "AI",
		MemoryKey:      "history",
	}

	return &m
}

// MemoryVariables gets the input key the buffer memory class will load dynamically.
func (m *Buffer) MemoryVariables() []string {
	return []string{m.InputKey}
}

// LoadMemoryVariables returns the previous chat messages stored. Previous chat messages are returned via the key specified in the MemoryKey field. Defaults to "history".
// If ReturnMessages is true the output is of the type []schema.ChatMessage.
// If ReturnMessages is false the output is a buffer string of the chat messages.
func (m *Buffer) LoadMemoryVariables(inputValuesGiven map[string]any) (map[string]any, error) {
	if m.ReturnMessages {
		return map[string]any{
			m.MemoryKey: m.ChatHistory.messages,
		}, nil
	}

	bufferString, err := schema.GetBufferString(m.ChatHistory.messages, m.HumanPrefix, m.AiPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		m.MemoryKey: bufferString,
	}, nil
}

// SaveContext saves input value as a user message and output value as ai message.
// By default the input and output key is an empty string. In cases where this is not changed the input and output values must only contain one entry. The two entries from the input and output values will be added as a user and ai message respectively.
// If a input key is set this key will be used to get the user message from the inputValues. The same is true for the output values map.
// The input and output values used has to be a string.
func (m *Buffer) SaveContext(inputValues map[string]any, outputValues map[string]any) error {
	userInputValue, err := getInputValue(inputValues, m.InputKey)
	if err != nil {
		return err
	}

	m.ChatHistory.AddUserMessage(userInputValue)

	aiOutputValue, err := getInputValue(outputValues, m.OutputKey)
	if err != nil {
		return nil
	}

	m.ChatHistory.AddAIMessage(aiOutputValue)

	return nil
}

// Clear sets the chat messages to a new and empty chat message history.
func (m *Buffer) Clear() error {
	m.ChatHistory = NewChatMessageHistory()
	return nil
}

func getInputValue(inputValues map[string]any, inputKey string) (string, error) {
	if inputKey != "" {
		inputValue, ok := inputValues[inputKey]
		if !ok {
			return "", fmt.Errorf("input values %v do not contain inputKey %s", inputValues, inputKey)
		}

		return getInputValueReturnToString(inputValue, inputValues, inputKey)
	}

	if len(inputValues) == 1 {
		for _, inputValue := range inputValues {
			return getInputValueReturnToString(inputValue, inputValues, inputKey)
		}
	}

	if len(inputValues) == 0 {
		return "", fmt.Errorf(`input values %v have 0 keys `, inputValues)
	}

	return "", fmt.Errorf(`input values %v have multiple keys. Specify input key when creating the buffer memory or remove keys`, inputValues)
}

func getInputValueReturnToString(inputValue interface{}, inputValues map[string]any, inputKey string) (string, error) {
	switch value := inputValue.(type) {
	case string:
		return value, nil
	default:
		return "", fmt.Errorf("input values to buffer memory must be string. Got type %T. Input values: %v. Memory input key: %s", inputValue, inputValues, inputKey)
	}
}
