package agents

import (
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

type CreationOptions struct {
	maxIterations      int
	outputKey          string
	promptPrefix       string
	formatInstructions string
	promptSuffix       string
	prompt             prompts.PromptTemplate
}

// CreationOption is a function type that can be used to modify the creation of the agents
// and executors.
type CreationOption func(*CreationOptions)

func (co CreationOptions) getMrklPrompt(tools []tools.Tool) prompts.PromptTemplate {
	if co.prompt.Template != "" {
		return co.prompt
	}

	return createMRKLPrompt(
		tools,
		co.promptPrefix,
		co.formatInstructions,
		co.promptSuffix,
	)
}

func executorDefaultOptions() CreationOptions {
	return CreationOptions{
		maxIterations: _defaultMaxIterations,
		outputKey:     _defaultOutputKey,
	}
}

func mrklDefaultOptions() CreationOptions {
	return CreationOptions{
		promptPrefix:       _defaultMrklPrefix,
		formatInstructions: _defaultMrklFormatInstructions,
		promptSuffix:       _defaultMrklSuffix,
		outputKey:          _defaultOutputKey,
	}
}

func conversationalDefaultOptions() CreationOptions {
	return CreationOptions{
		promptPrefix:       _defaultConverationalPreffix,
		formatInstructions: _defaultConverationalFormatInstructions,
		promptSuffix:       _defaultConverationalSuffix,
		outputKey:          _defaultOutputKey,
	}
}

func (co CreationOptions) getConversationalPrompt(tools []tools.Tool) prompts.PromptTemplate {
	if co.prompt.Template != "" {
		return co.prompt
	}

	return createConversationalPrompt(
		tools,
		co.promptPrefix,
		co.formatInstructions,
		co.promptSuffix,
	)
}

// WithMaxIterations is an option for setting the max number of iterations the executor
// will complete.
func WithMaxIterations(iterations int) CreationOption {
	return func(co *CreationOptions) {
		co.maxIterations = iterations
	}
}

// WithOutputKey is an option for setting the output key of the agent.
func WithOutputKey(outputKey string) CreationOption {
	return func(co *CreationOptions) {
		co.outputKey = outputKey
	}
}

// WithPromptPrefix is an option for setting the prefix of the prompt used by the agent.
func WithPromptPrefix(prefix string) CreationOption {
	return func(co *CreationOptions) {
		co.promptPrefix = prefix
	}
}

// WithPromptFormatInstructions is an option for setting the format instructions of the
// prompt used by the agent.
func WithPromptFormatInstructions(instructions string) CreationOption {
	return func(co *CreationOptions) {
		co.formatInstructions = instructions
	}
}

// WithPromptFormatInstructions is an option for setting the suffix of the prompt used by the agent.
func WithPromptSuffix(suffix string) CreationOption {
	return func(co *CreationOptions) {
		co.promptSuffix = suffix
	}
}

// WithPrompt is an option for setting the prompt the agent will use.
func WithPrompt(prompt prompts.PromptTemplate) CreationOption {
	return func(co *CreationOptions) {
		co.prompt = prompt
	}
}
