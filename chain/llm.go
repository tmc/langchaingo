package chain

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

type LLMChain struct {
	prompt       *prompts.PromptTemplate
	llm          llms.LLM
	Memory       schema.Memory
	OutputParser schema.OutputParser[any]

	OutputKey string
}

var _ Chain = LLMChain{}

// NewLLMChain creates a new LLMChain with an llm and a prompt.
func NewLLMChain(llm llms.LLM, prompt *prompts.PromptTemplate) LLMChain {
	chain := LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputParser: outputparser.NewSimple(),
		Memory:       memory.NewSimple(),

		OutputKey: "text",
	}

	return chain
}

// Call_ formats the prompts with the input values, generates using the llm and, parses
// the output from the llm with the output parser. This function should not be called
// directly, use rather the Call or Run function.
func (c LLMChain) Call_(ctx context.Context, values map[string]any) (map[string]any, error) {
	promptValue, err := c.prompt.FormatPrompt(values)
	if err != nil {
		return nil, err
	}

	var stop []string
	if stopVal, ok := values["stop"].([]string); ok {
		stop = stopVal
	}

	generations, err := c.llm.Generate(ctx, []string{promptValue.String()}, stop)
	if err != nil {
		return nil, err
	}

	finalOutput, err := c.OutputParser.ParseWithPrompt(generations[0].Text, promptValue)
	if err != nil {
		return nil, err
	}

	return map[string]any{c.OutputKey: finalOutput}, nil
}

// GetMemory returns the memory.
func (c LLMChain) GetMemory() schema.Memory {
	return c.Memory
}

// GetInputKeys returns the input keys.
func (c LLMChain) GetInputKeys() []string {
	return c.prompt.InputVariables
}

// GetOutputKeys returns the output keys the chain will return.
func (c LLMChain) GetOutputKeys() []string {
	return []string{c.OutputKey}
}
