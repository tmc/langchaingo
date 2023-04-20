package mrkl

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/exp/tools"
)

const PREFIX = `Answer the following questions as best you can. You have access to the following Tools:`
const FORMAT_INSTRUCTIONS = `Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [ {tool_names} ]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question`
const SUFFIX = `Begin!

Question: {input}
Thought:{agent_scratchpad}`

type CreatePromptOptions struct {
	Prefix             string
	Suffix             string
	FormatInstructions string
	InputVariables     []string
}

type PromptTemplateOption func(p *CreatePromptOptions)

func WithPrefix(prefix string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.Prefix = prefix
	}
}

func WithSuffix(suffix string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.Suffix = suffix
	}
}

func WithFormatInstructions(formatInstructions string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.FormatInstructions = formatInstructions
	}
}

func WithInputVariables(inputVariables []string) PromptTemplateOption {
	return func(p *CreatePromptOptions) {
		p.InputVariables = inputVariables
	}
}

func (options *CreatePromptOptions) handleDefaultValues() {
	if options.Prefix == "" {
		options.Prefix = PREFIX
	}

	if options.Suffix == "" {
		options.Suffix = SUFFIX
	}

	if options.FormatInstructions == "" {
		options.FormatInstructions = FORMAT_INSTRUCTIONS
	}
}

func createPrompt(
	tools []tools.Tool,
	options ...PromptTemplateOption,
) (prompts.PromptTemplate, error) {
	opts := &CreatePromptOptions{}
	opts.handleDefaultValues()
	for _, option := range options {
		option(opts)
	}

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
	formatInstructions := strings.Replace(FORMAT_INSTRUCTIONS, "{tool_names}", toolsNames.String(), -1)

	template := strings.Join([]string{opts.Prefix, toolsStrings.String(), formatInstructions, opts.Suffix}, "\n\n")

	return prompts.NewPromptTemplate(template, []string{"input", "tool_names", "agent_scratchpad"})
}
