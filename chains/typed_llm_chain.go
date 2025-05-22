package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// TypedLLMChain is a generic version of LLMChain that produces typed outputs.
type TypedLLMChain[T any] struct {
	Prompt           prompts.FormatPrompter
	LLM              llms.Model
	Memory           schema.Memory
	CallbacksHandler callbacks.Handler
	OutputParser     schema.OutputParser[T]
	OutputKey        string
}

var _ TypedChain[string] = (*TypedLLMChain[string])(nil)

// NewTypedLLMChain creates a new typed LLM chain.
func NewTypedLLMChain[T any](
	llm llms.Model,
	prompt prompts.FormatPrompter,
	outputParser schema.OutputParser[T],
	opts ...ChainCallOption,
) *TypedLLMChain[T] {
	opt := &chainCallOption{}
	for _, o := range opts {
		o(opt)
	}

	return &TypedLLMChain[T]{
		Prompt:           prompt,
		LLM:              llm,
		OutputParser:     outputParser,
		Memory:           memory.NewSimple(),
		CallbacksHandler: opt.CallbackHandler,
		OutputKey:        _llmChainDefaultOutputKey,
	}
}

// NewStringChain creates a new LLM chain that produces string outputs.
func NewStringChain(
	llm llms.Model,
	prompt prompts.FormatPrompter,
	opts ...ChainCallOption,
) *TypedLLMChain[string] {
	return NewTypedLLMChain(llm, prompt, outputparser.NewSimple(), opts...)
}

// Call executes the chain and returns a typed result plus metadata.
func (c *TypedLLMChain[T]) Call(
	ctx context.Context,
	values map[string]any,
	options ...ChainCallOption,
) (T, map[string]any, error) {
	var zero T

	promptValue, err := c.Prompt.FormatPrompt(values)
	if err != nil {
		return zero, nil, err
	}

	result, err := llms.GenerateFromSinglePrompt(
		ctx,
		c.LLM,
		promptValue.String(),
		GetLLMCallOptions(options...)...,
	)
	if err != nil {
		return zero, nil, err
	}

	// Parse to typed output
	finalOutput, err := c.OutputParser.ParseWithPrompt(result, promptValue)
	if err != nil {
		return zero, nil, err
	}

	// Return typed output + metadata
	metadata := map[string]any{
		"prompt":     promptValue.String(),
		"raw_output": result,
		"llm_type":   fmt.Sprintf("%T", c.LLM),
	}

	return finalOutput, metadata, nil
}

// GetMemory returns the memory.
func (c *TypedLLMChain[T]) GetMemory() schema.Memory {
	return c.Memory
}

// GetInputKeys returns the input keys required by the prompt.
func (c *TypedLLMChain[T]) GetInputKeys() []string {
	return c.Prompt.InputVariables()
}

// OutputType returns a description of the output type.
func (c *TypedLLMChain[T]) OutputType() string {
	return fmt.Sprintf("TypedLLMChain[%T]", *new(T))
}

// AsLegacyChain converts this typed chain to work with the legacy Chain interface.
func (c *TypedLLMChain[T]) AsLegacyChain() Chain {
	return NewChainAdapter(c)
}

// GetCallbackHandler returns the callback handler (implements callbacks.HandlerHaver).
func (c *TypedLLMChain[T]) GetCallbackHandler() callbacks.Handler {
	return c.CallbacksHandler
}