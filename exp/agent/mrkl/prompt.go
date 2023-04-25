package mrkl

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/prompts"
)

const (
	Prefix             = `Answer the following questions as best you can. You have access to the following Tools:`
	FormatInstructions = `Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [ {tool_names} ]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question`
)
const Suffix = `Begin!

Question: {input}
Thought:{agent_scratchpad}`

// CreatePromptOptions is a struct that holds options for creating a prompt template.
type CreatePromptOptions struct {
	Prefix             string
	Suffix             string
	FormatInstructions string
	InputVariables     []string
}

// PromptTemplateOption is a function type that can be used to modify the CreatePromptOptions.
type PromptTemplateOption func(p *CreatePromptOptions)

// WithPrefix is a function that sets a custom prefix for the prompt template.
func WithPrefix(prefix string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.Prefix = prefix
	}
}

// WithSuffix is a function that sets a custom suffix for the prompt template.
func WithSuffix(suffix string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.Suffix = suffix
	}
}

// WithFormatInstructions is a function that sets custom format instructions for the prompt template.
func WithFormatInstructions(formatInstructions string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.FormatInstructions = formatInstructions
	}
}

// WithInputVariables is a function that sets custom input variables for the prompt template.
func WithInputVariables(inputVariables []string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.InputVariables = inputVariables
	}
}

func (options *CreatePromptOptions) handleDefaultValues() {
	if options.Prefix == "" {
		options.Prefix = Prefix
	}

	if options.Suffix == "" {
		options.Suffix = Suffix
	}

	if options.FormatInstructions == "" {
		options.FormatInstructions = FormatInstructions
	}
}

// createPrompt is a function that takes a slice of tools and a variadic list of prompt
// template options, and returns a prompt template with the specified options.
// It returns an error if there is any issue during the creation process.
func createPrompt(
	tools []tools.Tool,
	options ...PromptTemplateOption,
) (prompts.PromptTemplate, error) {
	opts := &CreatePromptOptions{}
	for _, option := range options {
		option(opts)
	}
	opts.handleDefaultValues()

	var toolsStrings strings.Builder
	for _, tool := range tools {
		toolsStrings.WriteString(fmt.Sprintf("%s: %s\n", tool.Name, tool.Description))
	}

	var toolsNames strings.Builder
	for i, tool := range tools {
		if i > 0 {
			toolsNames.WriteString(", ")
		}
		toolsNames.WriteString(tool.Name)
	}
	formatInstructions := strings.Replace(FormatInstructions, "{tool_names}", toolsNames.String(), -1)

	template := strings.Join([]string{opts.Prefix, toolsStrings.String(), formatInstructions, opts.Suffix}, "\n\n")

	return prompts.PromptTemplate{
		Template:       template,
		InputVariables: []string{"input", "tool_names", "agent_scratchpad"},
	}, nil
}
