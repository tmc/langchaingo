package postgre

import (
	"errors"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/postgre/internal"
	"github.com/tmc/langchaingo/schema"
)

// ErrInvalidInputValues is returned when input values given to a memory in save context are invalid.
var ErrInvalidInputValues = errors.New("invalid input values")

type PostgreBuffer struct {
	ChatHistory *memory.ChatMessageHistory
	DB          *internal.Database

	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AIPrefix       string
	MemoryKey      string
}

var _ schema.Memory = &PostgreBuffer{}

func NewPostgreBuffer(dsn string) *PostgreBuffer {
	buffer := PostgreBuffer{
		ChatHistory: memory.NewChatMessageHistory(),

		ReturnMessages: false,
		InputKey:       "",
		OutputKey:      "",
		HumanPrefix:    "Human",
		AIPrefix:       "AI",
		MemoryKey:      "history",
	}

	db, err := internal.NewDatabase(dsn)
	if err != nil {
		log.Fatal(err)
	}

	buffer.DB = db

	return &buffer
}

func (buffer *PostgreBuffer) SetSession(id string) {
	buffer.DB.SetSession(id)
}

func (buffer *PostgreBuffer) SessionID() string {
	return buffer.DB.SessionID()
}

func (buffer *PostgreBuffer) MemoryVariables() []string {
	return []string{buffer.MemoryKey}
}

func (buffer *PostgreBuffer) LoadMemoryVariables(inputs map[string]any) (map[string]any, error) {
	msgs, err := buffer.DB.GetHistroy()
	if err != nil {
		return nil, err
	}

	buffer.ChatHistory = memory.NewChatMessageHistory(
		memory.WithPreviousMessages(msgs),
	)

	if buffer.ReturnMessages {
		return map[string]any{
			buffer.MemoryKey: buffer.ChatHistory.Messages(),
		}, nil
	}

	bufferString, err := schema.GetBufferString(buffer.ChatHistory.Messages(), buffer.HumanPrefix, buffer.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		buffer.MemoryKey: bufferString,
	}, nil
}

// SaveContext saves the context of the PostgreBuffer.
//
// It takes in two maps, inputs and outputs, which contain key-value pairs of any type.
// The function retrieves the value associated with buffer.InputKey from the inputs map
// and adds it as a user message to the ChatHistory. Then, it retrieves the value
// associated with buffer.OutputKey from the outputs map and adds it as an AI message
// to the ChatHistory. The function then uses the ChatHistory, HumanPrefix, and AIPrefix
// properties of the buffer to generate a bufferString using the GetBufferString function
// from the schema package. Finally, it saves the ChatHistory messages and bufferString
// to the DB using the SaveHistory function, and returns any error encountered.
//
// Return type: error.
func (buffer *PostgreBuffer) SaveContext(inputs map[string]any, outputs map[string]any) error {
	userInputValue, err := getInputValue(inputs, buffer.InputKey)
	if err != nil {
		return err
	}

	buffer.ChatHistory.AddUserMessage(userInputValue)

	aiOutPutValue, err := getInputValue(outputs, buffer.OutputKey)
	if err != nil {
		return err
	}

	buffer.ChatHistory.AddAIMessage(aiOutPutValue)

	bufferString, err := schema.GetBufferString(buffer.ChatHistory.Messages(), buffer.HumanPrefix, buffer.AIPrefix)
	if err != nil {
		return err
	}

	err = buffer.DB.SaveHistory(buffer.ChatHistory.Messages(), bufferString)
	if err != nil {
		return err
	}

	return nil
}

func (buffer *PostgreBuffer) Clear() error {
	buffer.ChatHistory.Clear()
	err := buffer.DB.ClearHistroy()
	if err != nil {
		return err
	}
	return nil
}

func getInputValue(inputValues map[string]any, inputKey string) (string, error) {
	// If the input key is set, return the value in the inputValues with the input key.
	if inputKey != "" {
		inputValue, ok := inputValues[inputKey]
		if !ok {
			return "", fmt.Errorf(
				"%w: %v do not contain inputKey %s",
				ErrInvalidInputValues,
				inputValues,
				inputKey,
			)
		}

		return getInputValueReturnToString(inputValue)
	}

	// Otherwise error if length of map isn't one, or return the only entry in the map.
	if len(inputValues) > 1 {
		return "", fmt.Errorf(
			"%w: multiple keys and no input key set",
			ErrInvalidInputValues,
		)
	}

	for _, inputValue := range inputValues {
		return getInputValueReturnToString(inputValue)
	}

	return "", fmt.Errorf("%w: 0 keys", ErrInvalidInputValues)
}

func getInputValueReturnToString(
	inputValue interface{},
) (string, error) {
	switch value := inputValue.(type) {
	case string:
		return value, nil
	default:
		return "", fmt.Errorf(
			"%w: input value %v not string",
			ErrInvalidInputValues,
			inputValue,
		)
	}
}
