package chains

import (
	"github.com/tmc/langchaingo/exp/memory"
	"github.com/tmc/langchaingo/exp/output_parsers"
	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type LLMChain struct {
	prompt       prompts.Template
	llm          llms.LLM
	OutputKey    string
	Memory       schema.Memory
	OutputParser output_parsers.OutputParser
}

func NewLLMChain(llm llms.LLM, prompt prompts.Template) LLMChain {
	chain := LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputKey:    "text",
		OutputParser: output_parsers.NewEmptyOutputParser(),
		Memory:       memory.NewEmptyMemory(),
	}

	return chain
}

func (c LLMChain) GetMemory() schema.Memory {
	return c.Memory
}

func (c LLMChain) Call(values map[string]any) (map[string]any, error) {
	var stop []string
	promptValue, err := c.prompt.FormatPromptValue(values)
	if err != nil {
		return map[string]any{}, err
	}

	if stopVal, ok := values["stop"].([]string); ok {
		stop = stopVal
	}

	generations, err := c.llm.Generate([]string{promptValue.String()}, stop)
	if err != nil {
		return map[string]any{}, err
	}

	finalOutput, err := c.OutputParser.ParseWithPrompt(generations[0].Text, promptValue)
	if err != nil {
		return map[string]any{}, err
	}

	return map[string]any{c.OutputKey: finalOutput}, nil
}
