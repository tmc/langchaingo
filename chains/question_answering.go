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

//nolint:lll
const _defaultMapRerankTemplate = `Use the following pieces of context to answer the question at the end. If you don't know the answer, just say that you don't know, don't try to make up an answer.
In addition to giving an answer, also return a score of how fully it answered the user's question. This should be in the following format:
Question: [question here]
Helpful Answer: [answer here]
Score: [score between 0 and 100]
How to determine the score:
- Higher is a better answer
- Better responds fully to the asked question, with sufficient level of detail
- If you do not know the answer based on the context, that should be a score of 0
- Don't be overconfident!
Example #1
Context:
---------
Apples are red
---------
Question: what color are apples?
Helpful Answer: red
Score: 100
Example #2
Context:
---------
it was night and the witness forgot his glasses. he was not sure if it was a sports car or an suv
---------
Question: what type was the car?
Helpful Answer: a sports car or an suv
Score: 60
Example #3
Context:
---------
Pears are either red or orange
---------
Question: what color are apples?
Helpful Answer: This document does not answer the question
Score: 0
Begin!
Context:
---------
{{.context}}
---------
Question: {{.question}}
Helpful Answer:`

// nolint: lll
const _defaultCondenseQuestionTemplate = `Given the following conversation and a follow up question, rephrase the follow up question to be a standalone question, in its original language.

Chat History:
{{.chat_history}}
Follow Up Input: {{.question}}
Standalone question:`

// LoadCondenseQuestionGenerator chain is used to generate a new question for the sake of retrieval.
func LoadCondenseQuestionGenerator(llm llms.Model) *LLMChain {
	condenseQuestionPromptTemplate := prompts.NewPromptTemplate(
		_defaultCondenseQuestionTemplate,
		[]string{"chat_history", "question"},
	)
	return NewLLMChain(llm, condenseQuestionPromptTemplate)
}

// LoadStuffQA loads a StuffDocuments chain with default prompts for the llm chain.
func LoadStuffQA(llm llms.Model) StuffDocuments {
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
func LoadRefineQA(llm llms.Model) RefineDocuments {
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

// LoadMapReduceQA loads a refine documents chain for question answering. Inputs are
// "question" and "input_documents".
func LoadMapReduceQA(llm llms.Model) MapReduceDocuments {
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

// LoadMapRerankQA loads a map rerank documents chain for question answering. Inputs are
// "question" and "input_documents".
func LoadMapRerankQA(llm llms.Model) MapRerankDocuments {
	mapRerankPrompt := prompts.NewPromptTemplate(
		_defaultMapRerankTemplate,
		[]string{"context", "question"},
	)

	mapRerankLLMChain := NewLLMChain(llm, mapRerankPrompt)

	mapRerank := NewMapRerankDocuments(mapRerankLLMChain)

	return *mapRerank
}
