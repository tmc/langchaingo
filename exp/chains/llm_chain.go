package chains

import (
	"context"

	"github.com/tmc/langchaingo/exp/output_parsers"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

type LLMChain struct {
	prompt       prompts.FormatPrompter
	llm          llms.LLM
	OutputKey    string
	Memory       schema.Memory
	OutputParser output_parsers.OutputParser
}

func NewLLMChain(llm llms.LLM, prompt prompts.FormatPrompter) LLMChain {
	chain := LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputKey:    "text",
		OutputParser: output_parsers.NewEmptyOutputParser(),
		Memory:       memory.NewSimple(),
	}

	return chain
}

func (c LLMChain) GetMemory() schema.Memory {
	return c.Memory
}

func (c LLMChain) Call(ctx context.Context, values map[string]any) (map[string]any, error) {
	var stop []string
	promptValue, err := c.prompt.FormatPrompt(values)
	if err != nil {
		return nil, err
	}

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
