package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

const _defaultStuffQATemplate = `Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.

{{.context}}

Question: {{.question}}
Helpful Answer:`

const _defaultRefineTemplate = `The original question is as follows: {{.question}}
We have provided an existing answer: {{.existing_answer}}
We have the opportunity to refine the existing answer
(only if needed) with some more context below.
------------
{{.context}}
------------
Given the new context, refine the original answer to better answer the question. 
If the context isn't useful, return the original answer.`

// LoadStuffQA loads a StuffDocuments chain with default prompts for the llm chain.
func LoadStuffQA(llm llms.LLM) StuffDocuments {
	defaultQAPromptTemplate := prompts.NewPromptTemplate(
		_defaultStuffQATemplate,
		[]string{"context", "question"},
	)

	qaPromptSelector := ConditionalPromptSelector{
		DefaultPrompt: defaultQAPromptTemplate,
	}

	prompt := qaPromptSelector.GetPrompt(llm)
	llmChain := NewLLMChain(llm, prompt)
	return NewStuffDocuments(llmChain)
}

// LoadRefineQA loads a refine documents chain for question answering. Inputs are
// "question" and "input_documents".
func LoadRefineQA(llm llms.LLM) RefineDocuments {
	questionPrompt := prompts.NewPromptTemplate(
		_defaultStuffQATemplate,
		[]string{"context", "question"},
	)
	refinePrompt := prompts.NewPromptTemplate(
		_defaultRefineTemplate,
		[]string{"question", "existing_answer", "context"},
	)

	return NewRefineDocuments(
		NewLLMChain(llm, questionPrompt),
		NewLLMChain(llm, refinePrompt),
	)
}
