package prompts

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	// ErrInvalidPartialVariableType is returned when a partial variable is not a supported type.
	// Valid types are string, int, float64, bool, and func() string, func() int, func() float64, func() bool.
	ErrInvalidPartialVariableType = errors.New("invalid partial variable type")
	// ErrNeedChatMessageList is returned when the variable is not a list of chat messages.
	ErrNeedChatMessageList = errors.New("variable should be a list of chat messages")
)

// PromptTemplate is a template that can be rendered with dynamic variables.
// It implements [Formatter] and [FormatPrompter] interfaces.
//
// [NewPromptTemplate] creates a template using Go template syntax by default.
// Example:
//
//	template := prompts.NewPromptTemplate(
//		"Summarize this {{.content}} in {{.style}} style",
//		[]string{"content", "style"},
//	)
//	prompt, err := template.FormatPrompt(data)
type PromptTemplate struct {
	// Template is the prompt template.
	Template string

	// A list of variable names the prompt template expects.
	InputVariables []string

	// TemplateFormat specifies the template syntax. Defaults to [TemplateFormatGoTemplate].
	// See [TemplateFormat] constants.
	TemplateFormat TemplateFormat

	// OutputParser is a function that parses the output of the prompt template.
	OutputParser schema.OutputParser[any]

	// PartialVariables pre-populates common values. Functions are called at render time.
	// Valid types: string, int, float64, bool or func() string, func() int,
	// func() float64, func() bool.
	PartialVariables map[string]any
}

// NewPromptTemplate creates a new [PromptTemplate] using [TemplateFormatGoTemplate] syntax.
// This is the recommended constructor for Go applications. The template format defaults
// to Go template syntax which is the idiomatic choice for Go applications.
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
		return "", fmt.Errorf("resolving partial values: %w", err)
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

// resolvePartialValues merges partial variables with provided values.
// Partial variable functions are called to get their current values.
// Supports string, int, float64, bool values and their corresponding function types.
func resolvePartialValues(partialValues map[string]any, values map[string]any) (map[string]any, error) {
	resolvedValues := make(map[string]any)

	// Track errors for potential multi-error reporting
	var errs []error

	for variable, value := range partialValues {
		switch value := value.(type) {
		case string:
			resolvedValues[variable] = value
		case int:
			resolvedValues[variable] = value
		case float64:
			resolvedValues[variable] = value
		case bool:
			resolvedValues[variable] = value
		case func() string:
			resolvedValues[variable] = value()
		case func() int:
			resolvedValues[variable] = value()
		case func() float64:
			resolvedValues[variable] = value()
		case func() bool:
			resolvedValues[variable] = value()
		default:
			errs = append(errs, fmt.Errorf("%w: variable %q has type %T", ErrInvalidPartialVariableType, variable, value))
		}
	}

	// If we have errors, join them and return
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Override with provided values
	for variable, value := range values {
		resolvedValues[variable] = value
	}

	return resolvedValues, nil
}
