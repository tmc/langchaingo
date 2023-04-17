package memory

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type Memory interface {
	LoadMemoryVariables(InputValues) (InputValues, error)
	SaveContext(InputValues, InputValues) error
}

type InputValues map[string]any

func getInputValue(inputValues InputValues, inputKey string) (string, error) {
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

func getInputValueReturnToString(inputValue interface{}, inputValues InputValues, inputKey string) (string, error) {
	switch value := inputValue.(type) {
	case string:
		return value, nil
	default:
		return "", fmt.Errorf("Input values to buffer memory must be string. Got type %T. Input values: %v. Memory input key: %s", inputValue, inputValues, inputKey)
	}
}

func getBufferString(messages []schema.ChatMessage, humanPrefix, aiPrefix string) string {
	stringMessages := make([]string, 0)

	for i := 0; i < len(messages); i++ {
		role := ""

		switch message := messages[i].(type) {
		case schema.AiChatMessage:
			role = aiPrefix
			break
		case schema.HumanChatMessage:
			role = humanPrefix
			break
		default:
			role = message.GetType()

		}

		stringMessages = append(stringMessages, fmt.Sprintf("%s: %s", role, messages[i].GetText()))
	}

	return strings.Join(stringMessages[:], "\n")
}

type EmptyMemory struct{}

func (m EmptyMemory) LoadMemoryVariables(InputValues) (InputValues, error) { return InputValues{}, nil }
func (m EmptyMemory) SaveContext(InputValues, InputValues) error           { return nil }

func NewEmptyMemory() EmptyMemory {
	return EmptyMemory{}
}
