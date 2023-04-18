package prompts

import "github.com/tmc/langchaingo/llms"

type PromptSelector interface {
	GetPrompt(llm llms.LLM) PromptTemplate
}

type Conditional struct {
	Condition func(llms.LLM) bool
	Prompt    PromptTemplate
}

type ConditionalPromptSelector struct {
	defaultPrompt PromptTemplate
	conditionals  []Conditional
}

func NewConditionalPromptSelector(defaultPrompt PromptTemplate, conditionals []Conditional) ConditionalPromptSelector {
	return ConditionalPromptSelector{
		defaultPrompt: defaultPrompt,
		conditionals:  conditionals,
	}
}

func (s ConditionalPromptSelector) GetPrompt(llm llms.LLM) PromptTemplate {
	for _, condition := range s.conditionals {
		if condition.Condition(llm) {
			return condition.Prompt
		}
	}

	return s.defaultPrompt
}
