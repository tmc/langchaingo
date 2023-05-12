package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

// LoadQAStuffChain loads a StuffDocuments chain with default prompts for the llm
// chain.
func LoadQAStuffChain(llm llms.LLM) StuffDocuments {
	defaultQAPromptTemplate, _ := prompts.NewPromptTemplate(
		"Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.\n\n{{.context}}\n\nQuestion: {{.question}}\nHelpful Answer:", //nolint
		[]string{"context", "question"},
	)

	qaPromptSelector := ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := NewLLMChain(llm, prompt)
	return NewStuffDocuments(llmChain)
}
