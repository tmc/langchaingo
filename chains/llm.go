package chains

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

const _llmChainDefaultOutputKey = "text"

type LLMChain struct {
	prompt       prompts.PromptTemplate
	llm          llms.LLM
	Memory       schema.Memory
	OutputParser schema.OutputParser[any]

	OutputKey string
}

var _ Chain = &LLMChain{}

// NewLLMChain creates a new LLMChain with an llm and a prompt.
func NewLLMChain(llm llms.LLM, prompt prompts.PromptTemplate) *LLMChain {
	chain := &LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputParser: outputparser.NewSimple(),
		Memory:       memory.NewSimple(),

		OutputKey: _llmChainDefaultOutputKey,
	}

	return chain
}

// Call formats the prompts with the input values, generates using the llm, and parses
// the output from the llm with the output parser. This function should not be called
// directly, use rather the Call or Run function if the prompt only requires one input
// value.
func (c LLMChain) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) {
	promptValue, err := c.prompt.FormatPrompt(values)
	if err != nil {
		return nil, err
	}

	generations, err := c.llm.Generate(ctx, []string{promptValue.String()}, getLLMCallOptions(options...)...)
	if err != nil {
		return nil, err
	}

	finalOutput, err := c.OutputParser.ParseWithPrompt(generations[0].Text, promptValue)
	if err != nil {
		return nil, err
	}

	return map[string]any{c.OutputKey: finalOutput}, nil
}

// Predict runs the chain and returns the output as a string. Returns an error
// if the output parser in the llm chain does not return a string.
func (c LLMChain) Predict(ctx context.Context, values map[string]any, options ...ChainCallOption) (string, error) {
	result, err := Call(ctx, c, values, options...)
	if err != nil {
		return "", err
	}

	output, ok := result[c.OutputKey].(string)
	if !ok {
		return "", ErrOutputNotStringInPredict
	}
	return output, nil
}

// GetMemory returns the memory.
func (c LLMChain) GetMemory() schema.Memory {
	return c.Memory //nolint:ireturn
}

// GetInputKeys returns the expected input keys.
func (c LLMChain) GetInputKeys() []string {
	return append([]string{}, c.prompt.InputVariables...)
}

// GetOutputKeys returns the output keys the chain will return.
func (c LLMChain) GetOutputKeys() []string {
	return []string{c.OutputKey}
}
