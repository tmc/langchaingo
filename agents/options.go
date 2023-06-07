package agents

import (
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

type creationOptions struct {
	maxIterations      int
	outputKey          string
	promptPrefix       string
	formatInstructions string
	promptSuffix       string
	prompt             prompts.PromptTemplate
}

// CreationOption is a function type that can be used to modify the creation of the agents
// and executors.
type CreationOption func(*creationOptions)

func (co creationOptions) getMrklPrompt(tools []tools.Tool) prompts.PromptTemplate {
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

func executorDefaultOptions() creationOptions {
	return creationOptions{
		maxIterations: _defaultMaxIterations,
		outputKey:     _defaultOutputKey,
	}
}

func mrklDefaultOptions() creationOptions {
	return creationOptions{
		promptPrefix:       _defaultMrklPrefix,
		formatInstructions: _defaultMrklFormatInstructions,
		promptSuffix:       _defaultMrklSuffix,
		outputKey:          _defaultOutputKey,
	}
}

// WithMaxIterations is an option for setting the max number of iterations the executor
// will complete.
func WithMaxIterations(iterations int) CreationOption {
	return func(co *creationOptions) {
		co.maxIterations = iterations
	}
}

// WithOutputKey is an option for setting the output key of the agent.
func WithOutputKey(outputKey string) CreationOption {
	return func(co *creationOptions) {
		co.outputKey = outputKey
	}
}

// WithPromptPrefix is an option for setting the prefix of the prompt used by the agent.
func WithPromptPrefix(prefix string) CreationOption {
	return func(co *creationOptions) {
		co.promptPrefix = prefix
	}
}

// WithPromptFormatInstructions is an option for setting the format instructions of the
// prompt used by the agent.
func WithPromptFormatInstructions(instructions string) CreationOption {
	return func(co *creationOptions) {
		co.formatInstructions = instructions
	}
}

// WithPromptFormatInstructions is an option for setting the suffix of the prompt used by the agent.
func WithPromptSuffix(suffix string) CreationOption {
	return func(co *creationOptions) {
		co.promptSuffix = suffix
	}
}

// WithPrompt is an option for setting the prompt the agent will use.
func WithPrompt(prompt prompts.PromptTemplate) CreationOption {
	return func(co *creationOptions) {
		co.prompt = prompt
	}
}
