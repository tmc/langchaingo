package prompts

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNoExample is returned when none of the Examples and ExampleSelector are provided.
	ErrNoExample = errors.New("no example is provided")
	// ErrExamplesAndExampleSelectorProvided is returned when there are no Examples and ExampleSelector.
	ErrExamplesAndExampleSelectorProvided = errors.New("only one of 'Examples' and 'example_selector' should be" +
		" provided")
)

// FewShotPrompt contains fields for a few-shot prompt.
type FewShotPrompt struct {
	// Examples to format into the prompt. Either this or ExamplePrompt should be provided.
	Examples []map[string]string
	// ExampleSelector to choose the examples to format into the prompt. Either this or Examples should be provided.
	ExampleSelector ExampleSelector
	// ExamplePrompt is used to format an individual example.
	ExamplePrompt PromptTemplate
	// A prompt template string to put before the examples.
	Prefix string
	// A prompt template string to put after the examples.
	Suffix string
	// A list of the names of the variables the prompt template expects.
	InputVariables map[string]any
	// Represents a map of variable names to values or functions that return values. If the value is a function, it will
	// be called when the prompt template is rendered.
	PartialVariables map[string]any
	// String separator used to join the prefix, the examples, and suffix.
	ExampleSeparator string
	// The format of the prompt template. Options are: 'f-string', 'jinja2'.
	TemplateFormat TemplateFormat
	// Whether to try validating the template.
	ValidateTemplate bool
}

// NewFewShotPrompt creates a new few-shot prompt with the given input. It returns error if there is no example, both
// examples and exampleSelector are provided, or CheckValidTemplate returns err when ValidateTemplate is true.
func NewFewShotPrompt(examplePrompt PromptTemplate, examples []map[string]string, exampleSelector ExampleSelector,
	prefix string, suffix string, input map[string]interface{}, partialInput map[string]interface{},
	exampleSeparator string, templateFormat TemplateFormat, validateTemplate bool,
) (*FewShotPrompt, error) {
	err := validateExamples(examples, exampleSelector)
	if err != nil {
		return nil, err
	}
	prompt := &FewShotPrompt{
		ExamplePrompt:    examplePrompt,
		Prefix:           prefix,
		Suffix:           suffix,
		InputVariables:   input,
		PartialVariables: partialInput,
		Examples:         examples,
		ExampleSelector:  exampleSelector,
		ExampleSeparator: "\n\n",
		TemplateFormat:   templateFormat,
		ValidateTemplate: validateTemplate,
	}
	if exampleSeparator != "" {
		prompt.ExampleSeparator = exampleSeparator
	}

	if prompt.ValidateTemplate {
		err := CheckValidTemplate(prompt.Prefix+prompt.Suffix, prompt.TemplateFormat, append(getMapKeys(input),
			getMapKeys(partialInput)...))
		if err != nil {
			return nil, err
		}
	}
	return prompt, nil
}

// validateExamples validates the provided example and exampleSelector. One of them must be provided only.
func validateExamples(examples []map[string]string, exampleSelector ExampleSelector) error {
	if examples != nil && exampleSelector != nil {
		return ErrExamplesAndExampleSelectorProvided
	} else if examples == nil && exampleSelector == nil {
		return ErrNoExample
	}
	return nil
}

// getExamples returns the provided examples or returns error when there is no example.
func (p *FewShotPrompt) getExamples(input map[string]string) ([]map[string]string, error) {
	switch {
	case p.Examples != nil:
		return p.Examples, nil
	case p.ExampleSelector != nil:
		return p.ExampleSelector.SelectExamples(input), nil
	default:
		return nil, ErrNoExample
	}
}

// Format assembles and formats the pieces of the prompt with the given input values and partial values.
func (p *FewShotPrompt) Format(values map[string]interface{}) (string, error) {
	resolvedValues, err := resolvePartialValues(p.PartialVariables, values)
	if err != nil {
		return "", err
	}
	stringResolvedValues := map[string]string{}
	for k, v := range resolvedValues {
		strVal, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("%w: %T", ErrInvalidPartialVariableType, v)
		}
		stringResolvedValues[k] = strVal
	}
	examples, err := p.getExamples(stringResolvedValues)
	if err != nil {
		return "", err
	}
	exampleStrings := make([]string, len(examples))

	for i, example := range examples {
		exampleMap := make(map[string]interface{})
		for k, v := range example {
			exampleMap[k] = v
		}

		res, err := p.ExamplePrompt.Format(exampleMap)
		if err != nil {
			return "", err
		}
		exampleStrings[i] = res
	}

	template := assemblePieces(p.Prefix, p.Suffix, exampleStrings, p.ExampleSeparator)
	return defaultformatterMapping[p.TemplateFormat](template, resolvedValues)
}

// assemblePieces assembles the pieces of the few-shot prompt.
func assemblePieces(prefix, suffix string, exampleStrings []string, separator string) string {
	const additionalCapacity = 2
	pieces := make([]string, 0, len(exampleStrings)+additionalCapacity)
	if prefix != "" {
		pieces = append(pieces, prefix)
	}

	for _, elem := range exampleStrings {
		if elem != "" {
			pieces = append(pieces, elem)
		}
	}

	if suffix != "" {
		pieces = append(pieces, suffix)
	}

	return strings.Join(pieces, separator)
}

// getMapKeys returns the keys of the provided map.
func getMapKeys(inputMap map[string]any) []string {
	keys := make([]string, 0, len(inputMap))
	for k := range inputMap {
		keys = append(keys, k)
	}
	return keys
}
