package prompts

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrInputVariableReserved      = errors.New("cannot have input variable")
	ErrPartialVariableReserved    = errors.New("cannot have partial variable")
	ErrInvalidPartialVariableType = errors.New("invalid partial variable type")
)

// PromptTemplateBaseOptions are the options that are common to all prompt templates.
type PromptTemplateBaseOptions struct {
	// A list of variable names the prompt template expects
	InputVariables []string
	// How to parse the output of calling an LLM on this formatted prompt
	OutputParser schema.OutputParser[any]
	// Partial variables
	PartialVariables map[string]any
}

// checkInputVariables validates the input variable names do not include restricted names.
func checkInputVariables(inputVariables []string) error {
	for _, variable := range inputVariables {
		if variable == "stop" {
			return ErrInputVariableReserved
		}
	}

	return nil
}

// checkPartialVariables validates the partial variable names do not include restricted names.
func checkPartialVariables(partialValues map[string]any) error {
	for k := range partialValues {
		if k == "stop" {
			return ErrPartialVariableReserved
		}
	}

	return nil
}

// applyPromptTemplateBaseOptions applies the input options to the default options.
func applyPromptTemplateBaseOptions(input PromptTemplateBaseOptions) (PromptTemplateBaseOptions, error) {
	opts := PromptTemplateBaseOptions{
		InputVariables:   make([]string, 0),
		OutputParser:     nil,
		PartialVariables: make(map[string]any),
	}
	if len(input.InputVariables) != 0 {
		if err := checkInputVariables(input.InputVariables); err != nil {
			return opts, err
		}
		opts.InputVariables = append(opts.InputVariables, input.InputVariables...)
	}
	if input.OutputParser != nil {
		opts.OutputParser = input.OutputParser
	}
	if len(input.PartialVariables) != 0 {
		if err := checkPartialVariables(input.PartialVariables); err != nil {
			return opts, err
		}

		for k, v := range input.PartialVariables {
			opts.PartialVariables[k] = v
		}
	}

	return opts, nil
}

// TemplateType is the string type key uniquely identifying this class
// of prompt template.
type TemplateType string

const (
	// TemplateTypePrompt is the type for SerializedPromptTemplate.
	TemplateTypePrompt TemplateType = "prompt"
	// TemplateTypeFewShot is the type for SerializedFewShotTemplate.
	TemplateTypeFewShot TemplateType = "few_shot"
	// TemplateTypeMessage is the type for SerializedMessagePromptTemplate.
	TemplateTypeMessage TemplateType = "message"
	// TemplateTypeChatPrompt is the type for SerializedChatPromptTemplate.
	TemplateTypeChatPrompt TemplateType = "chat_prompt"
)

// PromptTemplater is an interface for prompt templates. It is also equivalent to
// BasePromptTemplate in langchain or BasePromptTemplate in langchainjs.
type PromptTemplater interface {
	MergePartialAndUserVariables(userVariables map[string]any) (map[string]any, error)
	// Format formats the prompt given the input values.
	Format(values map[string]any) (string, error)
	// FormatPromptValue formats the prompt given the input values and return a formatted prompt value.
	FormatPromptValue(values map[string]any) (schema.PromptValue, error)
	// GetPromptType returns the string type key uniquely identifying this class of prompt template.
	GetPromptType() TemplateType
}

// formatPromptValue formats the prompt given the input values and return a formatted prompt value.
func formatPromptValue(promptTemplater PromptTemplater, values map[string]any) (schema.PromptValue, error) {
	formattedPrompt, err := promptTemplater.Format(values)
	if err != nil {
		return nil, err
	}

	return NewStringPromptValue(formattedPrompt), nil
}

// mergePartialAndUserVariables merges the partial variables and user variables. This is the common
// and default implementation for PromptTemplater.
func mergePartialAndUserVariables(
	partialVariables map[string]any,
	userVariables map[string]any,
) (map[string]any, error) {
	if partialVariables == nil {
		partialVariables = make(map[string]any)
	}

	partialValues := make(map[string]any)
	for k, v := range partialVariables {
		switch val := v.(type) {
		case string:
			partialValues[k] = val
		case func() any:
			partialValues[k] = val()
		default:
			return make(map[string]any), fmt.Errorf("%w: %T of %v", ErrInvalidPartialVariableType, val, val)
		}
	}

	allKwargs := make(map[string]any, len(partialValues)+len(userVariables))
	for k, v := range partialValues {
		allKwargs[k] = v
	}
	for k, v := range userVariables {
		allKwargs[k] = v
	}

	return allKwargs, nil
}

// PromptTemplateOptions is the option to create a PromptTemplate. Is is also equivalent to
// PromptTemplate.
type PromptTemplateOptions struct {
	PromptTemplateBaseOptions

	// The propmt template
	Template string
	// The format of the prompt template. Options are 'go-template', 'f-string', 'jinja-2'
	TemplateFormat TemplateFormat
	// Whether or not to try validating the template on initialization
	ValidateTemplate bool
}

var _ PromptTemplater = PromptTemplate{}

// PromptTemplate is the default implementation of PromptTemplater.
type PromptTemplate struct {
	opts PromptTemplateOptions
}

// NewPromptTemplate creates a new PromptTemplate from the given option.
func NewPromptTemplate(inputOpts PromptTemplateOptions) (*PromptTemplate, error) {
	baseOpts, err := applyPromptTemplateBaseOptions(inputOpts.PromptTemplateBaseOptions)
	if err != nil {
		return nil, err
	}

	opts := PromptTemplateOptions{
		PromptTemplateBaseOptions: baseOpts,
		TemplateFormat:            TemplateFormatGoTemplate,
		ValidateTemplate:          true,
	}
	if inputOpts.Template != "" {
		opts.Template = inputOpts.Template
	}
	if inputOpts.TemplateFormat != "" && inputOpts.TemplateFormat != TemplateFormatGoTemplate {
		opts.TemplateFormat = inputOpts.TemplateFormat
	}
	if !inputOpts.ValidateTemplate {
		opts.ValidateTemplate = inputOpts.ValidateTemplate
	}

	p := &PromptTemplate{opts: opts}

	if p.opts.ValidateTemplate {
		totalInputVariables := make([]string, 0, len(p.opts.InputVariables)+len(p.opts.PartialVariables))
		totalInputVariables = append(totalInputVariables, p.opts.InputVariables...)
		totalInputVariables = append(totalInputVariables, lo.Keys(p.opts.PartialVariables)...)
		err = CheckValidTemplate(p.opts.Template, p.opts.TemplateFormat, totalInputVariables)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p PromptTemplate) MergePartialAndUserVariables(userVariables map[string]any) (map[string]any, error) {
	return mergePartialAndUserVariables(p.opts.PartialVariables, userVariables)
}

func (p PromptTemplate) GetPromptType() TemplateType {
	return TemplateTypePrompt
}

func (p PromptTemplate) Format(values map[string]any) (string, error) {
	allValues, err := p.MergePartialAndUserVariables(values)
	if err != nil {
		return "", err
	}

	return RenderTemplate(p.opts.Template, p.opts.TemplateFormat, allValues)
}

func (p PromptTemplate) FormatPromptValue(values map[string]any) (schema.PromptValue, error) {
	formattedPrompt, err := p.Format(values)
	if err != nil {
		return nil, err
	}

	return NewStringPromptValue(formattedPrompt), nil
}

// PromptTemplateFromExamplesOption is the option to create a PromptTemplate from examples.
type PromptTemplateFromExamplesOption struct {
	// List of examples to use in the prompt.
	Examples []string
	// String to go after the list of examples. Should generally set up the user's input.
	Suffix string
	// A list of variable names the final prompt template will expect
	InputVariables []string
	// The separator to use in between examples
	ExampleSeparator string
	// String that should go before any examples. Generally includes examples.
	Prefix string
}

// NewPromptTemplateFromExamples takes examples in list format with prefix and suffix
// to create a prompt. It is equivalent to PromptTemplate.from_examples in langchain and
// PromptTemplate.fromExamples in langchainjs.
//
// Intended to be used a way to dynamically create a prompt from examples.
func NewPromptTemplateFromExamples(inputOpts PromptTemplateFromExamplesOption) (*PromptTemplate, error) {
	opts := PromptTemplateFromExamplesOption{
		Examples:         make([]string, 0),
		InputVariables:   make([]string, 0),
		ExampleSeparator: "\n\n",
	}
	if inputOpts.Examples != nil {
		opts.Examples = inputOpts.Examples
	}
	if inputOpts.Suffix != "" {
		opts.Suffix = inputOpts.Suffix
	}
	if inputOpts.InputVariables != nil {
		opts.InputVariables = inputOpts.InputVariables
	}
	if inputOpts.ExampleSeparator != "" {
		opts.ExampleSeparator = inputOpts.ExampleSeparator
	}
	if inputOpts.Prefix != "" {
		opts.Prefix = inputOpts.Prefix
	}

	templateSlice := make([]string, 0, len(inputOpts.Examples)+1+1)
	templateSlice = append(templateSlice, inputOpts.Prefix)
	templateSlice = append(templateSlice, inputOpts.Examples...)
	templateSlice = append(templateSlice, inputOpts.Suffix)
	template := strings.Join(templateSlice, inputOpts.ExampleSeparator)
	return NewPromptTemplate(PromptTemplateOptions{
		PromptTemplateBaseOptions: PromptTemplateBaseOptions{
			InputVariables: inputOpts.InputVariables,
		},
		Template: template,
	})
}

// PromptTemplateFromFStringOption is the option to create a PromptTemplate from an f-string template.
type PromptTemplateFromFStringOption struct {
	// The f-string template
	Template string
	// Whether or not to try validating the template on initialization
	ValidateTemplate bool
	// How to parse the output of calling an LLM on this formatted prompt
	OutputParser schema.OutputParser[any]
	// Partial variables
	PartialVariables map[string]any
}

// NewPromptTemplateFromFStringTemplate loads prompt template from a template f-string. It is equivalent
// to PromptTemplate.from_template in langchain and PromptTemplate.fromTemplate in langchainjs.
func NewPromptTemplateFromFStringTemplate(PromptTemplateFromFStringOption) (*PromptTemplate, error) {
	// TODO: implement, after f-string template is implemented
	panic("not implemented")
}

// PromptTemplateFromFileOption is the option to create a PromptTemplate from a file.
type PromptTemplateFromFileOption struct {
	// The path to the file containing the template
	TemplateFile string
	// A list of variable names the prompt template expects
	InputVariables []string
}

// NewPromptTemplateFromFile loads prompt template from a file. It is equivalent to
// PromptTemplate.from_file in langchain.
func NewPromptTemplateFromFile(PromptTemplateFromFileOption) (*PromptTemplate, error) {
	// TODO: implement
	panic("not implemented")
}
