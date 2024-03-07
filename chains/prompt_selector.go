package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

// PromptSelector is the interface for selecting a formatter depending on the
// LLM given.
type PromptSelector interface {
	GetPrompt(llm llms.Model) prompts.PromptTemplate
}

// ConditionalPromptSelector is a formatter selector that selects a prompt
// depending on conditionals.
type ConditionalPromptSelector struct {
	DefaultPrompt prompts.PromptTemplate
	Conditionals  []struct {
		Condition func(llms.Model) bool
		Prompt    prompts.PromptTemplate
	}
}

var _ PromptSelector = ConditionalPromptSelector{}

func (s ConditionalPromptSelector) GetPrompt(llm llms.Model) prompts.PromptTemplate {
	for _, conditional := range s.Conditionals {
		if conditional.Condition(llm) {
			return conditional.Prompt
		}
	}

	return s.DefaultPrompt
}
