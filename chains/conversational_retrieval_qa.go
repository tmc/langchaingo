package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	_conversationalRetrievalQADefaultInputKey             = "question"
	_conversationalRetrievalQADefaultSourceDocumentKey    = "source_documents"
	_conversationalRetrievalQADefaultGeneratedQuestionKey = "generated_question"
)

// ConversationalRetrievalQA chain builds on RetrievalQA to provide a chat history component.
type ConversationalRetrievalQA struct {
	// Retriever used to retrieve the relevant documents.
	Retriever schema.Retriever

	// Memory that remembers previous conversational back and forths directly.
	Memory schema.Memory

	// CombineDocumentsChain The chain used to combine any retrieved documents.
	CombineDocumentsChain Chain

	// CondenseQuestionChain The chain the documents and query is given to.
	// The chain used to generate a new question for the sake of retrieval.
	// This chain will take in the current question (with variable `question`)
	// and any chat history (with variable `chat_history`) and will produce
	// a new standalone question to be used later on.
	CondenseQuestionChain Chain

	// OutputKey The output key to return the final answer of this chain in.
	OutputKey string

	// RephraseQuestion Whether to pass the new generated question to the CombineDocumentsChain.
	// If true, will pass the new generated question along.
	// If false, will only use the new generated question for retrieval and pass the
	// original question along to the CombineDocumentsChain.
	RephraseQuestion bool

	// ReturnGeneratedQuestion Return the generated question as part of the final result.
	ReturnGeneratedQuestion bool

	// InputKey The input key to get the query from, by default "query".
	InputKey string

	// ReturnSourceDocuments Return the retrieved source documents as part of the final result.
	ReturnSourceDocuments bool
}

var _ Chain = ConversationalRetrievalQA{}

// NewConversationalRetrievalQA creates a new NewConversationalRetrievalQA.
func NewConversationalRetrievalQA(
	combineDocumentsChain Chain,
	condenseQuestionChain Chain,
	retriever schema.Retriever,
	memory schema.Memory,
) ConversationalRetrievalQA {
	return ConversationalRetrievalQA{
		Memory:                  memory,
		Retriever:               retriever,
		CombineDocumentsChain:   combineDocumentsChain,
		CondenseQuestionChain:   condenseQuestionChain,
		InputKey:                _conversationalRetrievalQADefaultInputKey,
		OutputKey:               _llmChainDefaultOutputKey,
		RephraseQuestion:        true,
		ReturnGeneratedQuestion: false,
		ReturnSourceDocuments:   false,
	}
}

func NewConversationalRetrievalQAFromLLM(
	llm llms.Model,
	retriever schema.Retriever,
	memory schema.Memory,
) ConversationalRetrievalQA {
	return NewConversationalRetrievalQA(
		LoadStuffQA(llm),
		LoadCondenseQuestionGenerator(llm),
		retriever,
		memory,
	)
}

// Call gets question, and relevant documents by question from the retriever and gives them to the combine
// documents chain.
func (c ConversationalRetrievalQA) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) { // nolint: lll
	query, ok := values[c.InputKey].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}
	chatHistoryStr, ok := values[c.Memory.GetMemoryKey(ctx)].(string)
	if !ok {
		chatHistory, ok := values[c.Memory.GetMemoryKey(ctx)].([]llms.ChatMessage)
		if !ok {
			return nil, fmt.Errorf("%w: %w", ErrMissingMemoryKeyValues, ErrMemoryValuesWrongType)
		}

		bufferStr, err := llms.GetBufferString(chatHistory, "Human", "AI")
		if err != nil {
			return nil, err
		}

		chatHistoryStr = bufferStr
	}

	question, err := c.getQuestion(ctx, query, chatHistoryStr)
	if err != nil {
		return nil, err
	}

	docs, err := c.Retriever.GetRelevantDocuments(ctx, question)
	if err != nil {
		return nil, err
	}

	result, err := Predict(ctx, c.CombineDocumentsChain, map[string]any{
		"question":        c.rephraseQuestion(query, question),
		"input_documents": docs,
	}, options...)
	if err != nil {
		return nil, err
	}

	output := make(map[string]any)

	output[_llmChainDefaultOutputKey] = result
	if c.ReturnSourceDocuments {
		output[_conversationalRetrievalQADefaultSourceDocumentKey] = docs
	}
	if c.ReturnGeneratedQuestion {
		output[_conversationalRetrievalQADefaultGeneratedQuestionKey] = question
	}

	return output, nil
}

func (c ConversationalRetrievalQA) GetMemory() schema.Memory {
	return c.Memory
}

func (c ConversationalRetrievalQA) GetInputKeys() []string {
	return []string{c.InputKey}
}

func (c ConversationalRetrievalQA) GetOutputKeys() []string {
	outputKeys := append([]string{}, c.CombineDocumentsChain.GetOutputKeys()...)
	if c.ReturnSourceDocuments {
		outputKeys = append(outputKeys, _conversationalRetrievalQADefaultSourceDocumentKey)
	}

	return outputKeys
}

func (c ConversationalRetrievalQA) getQuestion(
	ctx context.Context,
	question string,
	chatHistoryStr string,
) (string, error) {
	if len(chatHistoryStr) == 0 {
		return question, nil
	}

	results, err := Call(
		ctx,
		c.CondenseQuestionChain,
		map[string]any{
			"chat_history": chatHistoryStr,
			"question":     question,
		},
	)
	if err != nil {
		return "", err
	}

	newQuestion, ok := results[c.OutputKey].(string)
	if !ok {
		return "", ErrInvalidOutputValues
	}

	return newQuestion, nil
}

func (c ConversationalRetrievalQA) rephraseQuestion(question string, newQuestion string) string {
	if c.RephraseQuestion {
		return newQuestion
	}

	return question
}
