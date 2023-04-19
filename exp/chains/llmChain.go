package chains

import (
	"github.com/tmc/langchaingo/exp/memory"
	"github.com/tmc/langchaingo/exp/outputParsers"
	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/llms"
)

type LLMChain struct {
	prompt       prompts.Template
	llm          llms.LLM
	OutputKey    string
	Memory       memory.Memory
	OutputParser outputParsers.OutputParser
}

func NewLLMChain(llm llms.LLM, prompt prompts.Template) LLMChain {
	chain := LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputKey:    "text",
		OutputParser: outputParsers.NewEmptyOutputParser(),
		Memory:       memory.NewEmptyMemory(),
	}

	return chain
}

func (c LLMChain) GetMemory() memory.Memory {
	return c.Memory
}

func (c LLMChain) Call(values ChainValues) (ChainValues, error) {
	var stop []string
	promptValue, err := c.prompt.FormatPromptValue(values)
	if err != nil {
		return ChainValues{}, err
	}

	if stopVal, ok := values["stop"].([]string); ok {
		stop = stopVal
	}

	generations, err := c.llm.Generate([]string{promptValue.String()}, stop)
	if err != nil {
		return ChainValues{}, err
	}

	finalOutput, err := c.OutputParser.ParseWithPrompt(generations[0].Text, promptValue)
	if err != nil {
		return ChainValues{}, err
	}

	return ChainValues{c.OutputKey: finalOutput}, nil
}
