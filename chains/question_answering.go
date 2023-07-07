package chains

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

//nolint:lll
const _defaultStuffQATemplate = `Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.

{{.context}}

Question: {{.question}}
Helpful Answer:`

const _defaultRefineQATemplate = `The original question is as follows: {{.question}}
We have provided an existing answer: {{.existing_answer}}
We have the opportunity to refine the existing answer
(only if needed) with some more context below.
------------
{{.context}}
------------
Given the new context, refine the original answer to better answer the question. 
If the context isn't useful, return the original answer.`

//nolint:lll
const _defaultMapReduceGetInformationQATemplate = `Use the following portion of a long document to see if any of the text is relevant to answer the question. 
Return any relevant text verbatim.
{{.context}}
Question: {{.question}}
Relevant text, if any:`

//nolint:lll
const _defaultMapReduceCombineQATemplate = `Given the following extracted parts of a long document and a question, create a final answer. 
If you don't know the answer, just say that you don't know. Don't try to make up an answer.

QUESTION: {{.question}}
=========
{{.context}}
=========
FINAL ANSWER:`

// LoadStuffQA loads a StuffDocuments chain with default prompts for the llm chain.
func LoadStuffQA(llm llms.LanguageModel) StuffDocuments {
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
func LoadRefineQA(llm llms.LanguageModel) RefineDocuments {
	questionPrompt := prompts.NewPromptTemplate(
		_defaultStuffQATemplate,
		[]string{"context", "question"},
	)
	refinePrompt := prompts.NewPromptTemplate(
		_defaultRefineQATemplate,
		[]string{"question", "existing_answer", "context"},
	)

	return NewRefineDocuments(
		NewLLMChain(llm, questionPrompt),
		NewLLMChain(llm, refinePrompt),
	)
}

// LoadRefineQA loads a refine documents chain for question answering. Inputs are
// "question" and "input_documents".
func LoadMapReduceQA(llm llms.LanguageModel) MapReduceDocuments {
	getInfoPrompt := prompts.NewPromptTemplate(
		_defaultMapReduceGetInformationQATemplate,
		[]string{"question", "context"},
	)
	combinePrompt := prompts.NewPromptTemplate(
		_defaultMapReduceCombineQATemplate,
		[]string{"question", "context"},
	)

	mapChain := NewLLMChain(llm, getInfoPrompt)
	reduceChain := NewStuffDocuments(
		NewLLMChain(llm, combinePrompt),
	)

	return NewMapReduceDocuments(mapChain, reduceChain)
}
