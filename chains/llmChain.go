package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/outputParsers"
	"github.com/tmc/langchaingo/prompts"
)

type LLMChain struct {
	prompt       prompts.PromptTemplate
	llm          llms.LLM
	OutputKey    string
	Memory       memory.Memory
	OutputParser outputParsers.OutputParser
}

func NewLLMChain(llm llms.LLM, prompt prompts.PromptTemplate) LLMChain {
	chain := LLMChain{
		prompt:       prompt,
		llm:          llm,
		OutputKey:    "text",
		OutputParser: outputParsers.NewEmptyOutputParser(),
		Memory:       memory.NewBufferMemory(),
	}

	return chain
}

func (c LLMChain) GetMemory() memory.Memory {
	return c.Memory
}

func (c LLMChain) GetChainType() string {
	return "llm_chain"
}

func (c LLMChain) Call(values ChainValues) (ChainValues, error) {
	//TODO: stop

	promptValue, err := c.prompt.FormatPromptValue(values)
	if err != nil {
		return ChainValues{}, err
	}

	generations, err := c.llm.Generate([]string{promptValue.String()})
	if err != nil {
		return ChainValues{}, err
	}

	finalOutput, err := c.OutputParser.ParseWithPrompt(generations[0].Text, promptValue)
	if err != nil {
		return ChainValues{}, err
	}

	return ChainValues{c.OutputKey: finalOutput}, nil
}
