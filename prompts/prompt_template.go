package prompts

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"golang.org/x/exp/maps"
)

var (
	// ErrInputVariableReserved is returned when there is a conflict with a reserved variable name.
	ErrInputVariableReserved = errors.New("conflict with reserved variable name")
	// ErrInvalidPartialVariableType is returned when the partial variable is not a string or a function.
	ErrInvalidPartialVariableType = errors.New("invalid partial variable type")
)

// Formatter is an interface for formatting a map of values into a string.
type Formatter interface {
	Format(values map[string]any) (string, error)
}

// FormatPrompter is an interface for formatting a map of values into a prompt value.
type FormatPrompter interface {
	FormatPrompt(values map[string]any) (schema.PromptValue, error)
}

// PromptTemplate contains common fields for all prompt templates.
type PromptTemplate struct {
	// A list of variable names the prompt template expects.
	InputVariables []string

	// OutputParser is a function that parses the output of the prompt template.
	OutputParser schema.OutputParser[any]

	// PartialVariables represents a map of variable names to values or functions that return values.
	// If the value is a function, it will be called when the prompt template is rendered.
	PartialVariables map[string]any
}

// checkInputVariables validates the input variable names do not include restricted names.
func checkInputVariables(inputVariables []string) error {
	for _, variable := range inputVariables {
		if variable == "stop" {
			return fmt.Errorf("%w: %v", ErrInputVariableReserved, variable)
		}
	}
	return nil
}

// checkPartialVariables validates the partial variable names do not include restricted names.
func checkPartialVariables(partialValues map[string]any) error {
	return checkInputVariables(maps.Keys(partialValues))
}
