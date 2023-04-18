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
	StopWords    []string
}

func NewLLMChain(llm llms.LLM, prompt prompts.Template, stopWords []string) LLMChain {
	chain := LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputKey:    "text",
		OutputParser: outputParsers.NewEmptyOutputParser(),
		Memory:       memory.NewEmptyMemory(),
		StopWords:    stopWords,
	}

	return chain
}

func (c LLMChain) GetMemory() memory.Memory {
	return c.Memory
}

func (c LLMChain) Call(values ChainValues) (ChainValues, error) {

	promptValue, err := c.prompt.FormatPromptValue(values)
	if err != nil {
		return ChainValues{}, err
	}

	generations, err := c.llm.Generate([]string{promptValue.String()}, c.StopWords)
	if err != nil {
		return ChainValues{}, err
	}

	finalOutput, err := c.OutputParser.ParseWithPrompt(generations[0].Text, promptValue)
	if err != nil {
		return ChainValues{}, err
	}

	return ChainValues{c.OutputKey: finalOutput}, nil
}
