package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

const stuffSummarizationTemplate = `Write a concise summary of the following:


"{{.context}}"


CONCISE SUMMARY:`

func LoadStuffSummarization(llm llms.LLM) StuffDocuments {
	llmChain := NewLLMChain(llm, prompts.NewPromptTemplate(
		stuffSummarizationTemplate, []string{"context"},
	))

	return NewStuffDocuments(llmChain)
}
