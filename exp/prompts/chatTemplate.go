package prompts

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/exp/schema"
)

/*
The Message struct has the responsibilities of BaseMessageStringPromptTemplate and all sub classes in the typeScript version.
*/

type Message struct {
	inputVariables []string
	prompt         Template
	toChatMessage  func(message string) schema.ChatMessage
}

func (m Message) format(values map[string]any) (schema.ChatMessage, error) {
	formatted, err := m.prompt.Format(values)
	if err != nil {
		return schema.AiChatMessage{}, err
	}

	return m.toChatMessage(formatted), nil
}

func NewSystemMessage(systemPrompt Template) Message {
	return Message{
		prompt:         systemPrompt,
		toChatMessage:  func(message string) schema.ChatMessage { return schema.SystemChatMessage{Text: message} },
		inputVariables: systemPrompt.GetInputVariables(),
	}
}

func NewHumanMessage(systemPrompt Template) Message {
	return Message{
		prompt:         systemPrompt,
		toChatMessage:  func(message string) schema.ChatMessage { return schema.HumanChatMessage{Text: message} },
		inputVariables: systemPrompt.GetInputVariables(),
	}
}

func NewAiMessage(systemPrompt Template) Message {
	return Message{
		prompt:         systemPrompt,
		toChatMessage:  func(message string) schema.ChatMessage { return schema.AiChatMessage{Text: message} },
		inputVariables: systemPrompt.GetInputVariables(),
	}
}

type ChatTemplate struct {
	promptMessages   []Message
	inputVariables   []string
	partialVariables map[string]any
	validateTemplate bool
}

func (t ChatTemplate) GetInputVariables() []string { return t.inputVariables }

func (t ChatTemplate) Format(inputValues map[string]any) (string, error) {
	promptValue, err := t.FormatPromptValue(inputValues)
	return promptValue.String(), err
}

func (t ChatTemplate) FormatPromptValue(inputValues map[string]any) (PromptValue, error) {
	resultMessages, err := t.formatMessages(inputValues)
	return ChatPromptValue{resultMessages}, err
}

func (t ChatTemplate) formatMessages(userValues map[string]any) ([]schema.ChatMessage, error) {
	allValues := mergePartialAndUserVariables(t.partialVariables, userValues)

	resultMessages := make([]schema.ChatMessage, 0)
	for _, promptMessage := range t.promptMessages {
		curMessageInputValues := make(map[string]any, 0)

		for _, inputVariable := range promptMessage.inputVariables {
			value, ok := allValues[inputVariable]
			if !ok {
				return []schema.ChatMessage{}, fmt.Errorf("Missing value for input variable %s ", inputVariable)
			}

			curMessageInputValues[inputVariable] = value
		}

		message, err := promptMessage.format(curMessageInputValues)
		if err != nil {
			return []schema.ChatMessage{}, err
		}

		resultMessages = append(resultMessages, message)
	}

	return resultMessages, nil
}

func (t ChatTemplate) validate() error {
	inputVariablesToMessages := make(map[string]bool, 0)
	for _, promptMessage := range t.promptMessages {
		for _, inputVariable := range promptMessage.inputVariables {
			inputVariablesToMessages[inputVariable] = true
		}
	}

	inputVariablesGiven := make(map[string]bool, 0)
	for _, inputVariable := range t.inputVariables {
		inputVariablesGiven[inputVariable] = true
	}

	for partialVariable := range t.partialVariables {
		inputVariablesGiven[partialVariable] = true
	}

	difference := make([]string, 0) // Slice of input variables given, that are not used in any of the messages
	for inputVariable := range inputVariablesGiven {
		if !inputVariablesToMessages[inputVariable] {
			difference = append(difference, inputVariable)
		}
	}

	if len(difference) > 0 {
		return fmt.Errorf("Input variables %v are not used in any of the prompt messages", difference)
	}

	otherDifference := make([]string, 0)
	for inputVariablesToMessage := range inputVariablesToMessages {
		if !inputVariablesGiven[inputVariablesToMessage] {
			otherDifference = append(otherDifference, inputVariablesToMessage)
		}
	}

	if len(otherDifference) > 0 {
		return fmt.Errorf("Input variables %v are used in prompt messages but not in the prompt template", otherDifference)
	}

	return nil
}

func NewChatTemplate(promptMessages []Message, inputVariables []string) (ChatTemplate, error) {
	t := ChatTemplate{
		promptMessages: promptMessages,
		inputVariables: inputVariables,
	}

	return t, t.validate()
}

type ChatPromptValue struct {
	messages []schema.ChatMessage
}

// Formats the ChatPromptValue as a JSON string
func (v ChatPromptValue) String() string {

	keyValueJSON := make([]string, 0)
	for i := 0; i < len(v.messages); i++ {
		keyValueJSON = append(keyValueJSON, fmt.Sprintf("{\"text\":\"%s\"}", v.messages[i].GetText()))
	}

	return fmt.Sprintf("[%s]", strings.Join(keyValueJSON, ",")) //Joins the string with [] around creating an array of objects
}

func (v ChatPromptValue) ToChatMessages() []schema.ChatMessage {
	return v.messages
}

type ChatTemplateOption func(p *ChatTemplate)

func WithPartialVariablesChat(partialVariables map[string]any) ChatTemplateOption {
	return func(t *ChatTemplate) {
		t.partialVariables = partialVariables
	}
}

func WithValidateTemplateChat(validateTemplate bool) ChatTemplateOption {
	return func(t *ChatTemplate) {
		t.validateTemplate = validateTemplate
	}
}
