package prompts

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	// ErrInputVariableReserved is returned when there is a conflict with a reserved variable name.
	ErrInputVariableReserved = errors.New("conflict with reserved variable name")
	// ErrInvalidPartialVariableType is returned when the partial variable is not a string or a function.
	ErrInvalidPartialVariableType = errors.New("invalid partial variable type")
	// ErrNeedChatMessageList is returned when the variable is not a list of chat messages.
	ErrNeedChatMessageList = errors.New("variable should be a list of chat messages")
)

// PromptTemplate contains common fields for all prompt templates.
type PromptTemplate struct {
	// Template is the prompt template.
	Template string

	// A list of variable names the prompt template expects.
	InputVariables []string

	// TemplateFormat is the format of the prompt template.
	TemplateFormat TemplateFormat

	// OutputParser is a function that parses the output of the prompt template.
	OutputParser schema.OutputParser[any]

	// PartialVariables represents a map of variable names to values or functions
	// that return values. If the value is a function, it will be called when the
	// prompt template is rendered.
	PartialVariables map[string]any
}

// NewPromptTemplate returns a new prompt template.
func NewPromptTemplate(template string, inputVars []string) PromptTemplate {
	return PromptTemplate{
		Template:       template,
		InputVariables: inputVars,
		TemplateFormat: TemplateFormatGoTemplate,
	}
}

var (
	_ Formatter      = PromptTemplate{}
	_ FormatPrompter = PromptTemplate{}
)

// Format formats the prompt template and returns a string value.
func (p PromptTemplate) Format(values map[string]any) (string, error) {
	resolvedValues, err := resolvePartialValues(p.PartialVariables, values)
	if err != nil {
		return "", err
	}

	return RenderTemplate(p.Template, p.TemplateFormat, resolvedValues)
}

// FormatPrompt formats the prompt template and returns a string prompt value.
func (p PromptTemplate) FormatPrompt(values map[string]any) (llms.PromptValue, error) { //nolint:ireturn
	f, err := p.Format(values)
	if err != nil {
		return nil, err
	}

	return StringPromptValue(f), nil //nolint:ireturn
}

// GetInputVariables returns the input variables the prompt expect.
func (p PromptTemplate) GetInputVariables() []string {
	return p.InputVariables
}

func resolvePartialValues(partialValues map[string]any, values map[string]any) (map[string]any, error) {
	resolvedValues := make(map[string]any)
	for variable, value := range partialValues {
		switch value := value.(type) {
		case string:
			resolvedValues[variable] = value
		case func() string:
			resolvedValues[variable] = value()
		default:
			return nil, fmt.Errorf("%w: %v", ErrInvalidPartialVariableType, variable)
		}
	}
	for variable, value := range values {
		resolvedValues[variable] = value
	}
	return resolvedValues, nil
}
