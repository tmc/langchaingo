package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

const _stuffSummarizationTemplate = `Write a concise summary of the following:


"{{.context}}"


CONCISE SUMMARY:`

const _refineSummarizationTemplate = `Your job is to produce a final concise summary
We have provided an existing summary up to a certain point: "{{.existing_answer}}"
We have the opportunity to refine the existing summary
(only if needed) with some more context below.
------------
"{{.context}}"
------------

Given the new context, refine the original summary
If the context isn't useful, return the original summary.

REFINED SUMMARY:`

// LoadStuffSummarization loads a summarization chain that stuffs all documents
// given into the prompt.
func LoadStuffSummarization(llm llms.Model) StuffDocuments {
	llmChain := NewLLMChain(llm, prompts.NewPromptTemplate(
		_stuffSummarizationTemplate, []string{"context"},
	))

	return NewStuffDocuments(llmChain)
}

// LoadRefineSummarization loads a refine documents chain for summarization of
// documents.
func LoadRefineSummarization(llm llms.Model) RefineDocuments {
	llmChain := NewLLMChain(llm, prompts.NewPromptTemplate(
		_stuffSummarizationTemplate, []string{"context"},
	))
	refineLLMChain := NewLLMChain(llm, prompts.NewPromptTemplate(
		_refineSummarizationTemplate, []string{"existing_answer", "context"},
	))

	return NewRefineDocuments(llmChain, refineLLMChain)
}

// LoadMapReduceSummarization loads a map reduce documents chain for
// summarization of documents.
func LoadMapReduceSummarization(llm llms.Model) MapReduceDocuments {
	mapChain := NewLLMChain(llm, prompts.NewPromptTemplate(
		_stuffSummarizationTemplate, []string{"context"},
	))
	combineChain := LoadStuffSummarization(llm)

	return NewMapReduceDocuments(mapChain, combineChain)
}
