package prompts

import "github.com/tmc/langchaingo/llms"

// ChatPromptTemplate is a prompt template for chat messages.
type ChatPromptTemplate struct {
	// Messages is the list of the messages to be formatted.
	Messages []MessageFormatter

	// PartialVariables represents a map of variable names to values or functions
	// that return values. If the value is a function, it will be called when the
	// prompt template is rendered.
	PartialVariables map[string]any
}

var (
	_ Formatter        = ChatPromptTemplate{}
	_ MessageFormatter = ChatPromptTemplate{}
	_ FormatPrompter   = ChatPromptTemplate{}
)

// FormatPrompt formats the messages into a chat prompt value.
func (p ChatPromptTemplate) FormatPrompt(values map[string]any) (llms.PromptValue, error) { //nolint:ireturn
	resolvedValues, err := resolvePartialValues(p.PartialVariables, values)
	if err != nil {
		return nil, err
	}

	formattedMessages := make([]llms.ChatMessage, 0, len(p.Messages))
	for _, m := range p.Messages {
		curFormattedMessages, err := m.FormatMessages(resolvedValues)
		if err != nil {
			return nil, err
		}

		formattedMessages = append(formattedMessages, curFormattedMessages...)
	}

	return ChatPromptValue(formattedMessages), nil
}

// Format formats the messages with values given and returns the messages as a string.
func (p ChatPromptTemplate) Format(values map[string]any) (string, error) {
	promptValue, err := p.FormatPrompt(values)
	return promptValue.String(), err
}

// FormatMessages formats the messages with the values and returns the formatted messages.
func (p ChatPromptTemplate) FormatMessages(values map[string]any) ([]llms.ChatMessage, error) {
	promptValue, err := p.FormatPrompt(values)
	if promptValue == nil {
		return nil, err
	}
	return promptValue.Messages(), err
}

// GetInputVariables returns the input variables the prompt expect.
func (p ChatPromptTemplate) GetInputVariables() []string {
	inputVariablesMap := make(map[string]bool, 0)
	for _, msg := range p.Messages {
		for _, variable := range msg.GetInputVariables() {
			inputVariablesMap[variable] = true
		}
	}

	inputVariables := make([]string, 0, len(inputVariablesMap))
	for variable := range inputVariablesMap {
		inputVariables = append(inputVariables, variable)
	}
	return inputVariables
}

// NewChatPromptTemplate creates a new chat prompt template from a list of message formatters.
func NewChatPromptTemplate(messages []MessageFormatter) ChatPromptTemplate {
	return ChatPromptTemplate{
		Messages: messages,
	}
}
