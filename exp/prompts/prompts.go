package prompts

import (
	"github.com/tmc/langchaingo/schema"
)

type Template interface {
	Format(inputValues map[string]any) (string, error)
	FormatPromptValue(inputValues map[string]any) (PromptValue, error)
	GetInputVariables() []string
}

type PromptValue interface {
	String() string
	ToChatMessages() []schema.ChatMessage
}

type StringPromptValue struct {
	value string
}

func (v StringPromptValue) String() string { return v.value }
func (v StringPromptValue) ToChatMessages() []schema.ChatMessage {
	return []schema.ChatMessage{schema.HumanChatMessage{Text: v.value}}
}

func mergePartialAndUserVariables(partialVariables, userVariables map[string]any) map[string]any {
	allValues := make(map[string]any)
	for variable, value := range partialVariables {
		allValues[variable] = value
	}

	for variable, value := range userVariables {
		allValues[variable] = value
	}

	return allValues
}
