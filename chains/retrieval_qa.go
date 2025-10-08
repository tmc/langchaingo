package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

const (
	_retrievalQADefaultInputKey          = "query"
	_retrievalQADefaultSourceDocumentKey = "source_documents"
)

// RetrievalQA is a chain used for question-answering against a retriever.
// First the chain gets documents from the retriever, then the documents
// and the query is used as input to another chain. Typically, that chain
// combines the documents into a prompt that is sent to an LLM.
type RetrievalQA struct {
	// Retriever used to retrieve the relevant documents.
	Retriever schema.Retriever

	// The chain the documents and query is given to.
	CombineDocumentsChain Chain

	// The input key to get the query from, by default "query".
	InputKey string

	// If the chain should return the documents used by the combine
	// documents chain in the "source_documents" key.
	ReturnSourceDocuments bool
}

var _ Chain = RetrievalQA{}

// NewRetrievalQA creates a new RetrievalQA from a retriever and a chain for
// combining documents. The chain for combining documents is expected to
// have the expected input values for the "question" and "input_documents"
// key.
func NewRetrievalQA(combineDocumentsChain Chain, retriever schema.Retriever) RetrievalQA {
	return RetrievalQA{
		Retriever:             retriever,
		CombineDocumentsChain: combineDocumentsChain,
		InputKey:              _retrievalQADefaultInputKey,
		ReturnSourceDocuments: false,
	}
}

// NewRetrievalQAFromLLM loads a question answering combine documents chain
// from the llm and creates a new retrievalQA chain.
func NewRetrievalQAFromLLM(llm llms.Model, retriever schema.Retriever) RetrievalQA {
	return NewRetrievalQA(
		LoadStuffQA(llm),
		retriever,
	)
}

// Call gets relevant documents from the retriever and gives them to the combine
// documents chain.
func (c RetrievalQA) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) { // nolint: lll
	query, ok := values[c.InputKey].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	docs, err := c.Retriever.GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, err
	}

	result, err := Call(ctx, c.CombineDocumentsChain, map[string]any{
		"question":        query,
		"input_documents": docs,
	}, options...)
	if err != nil {
		return nil, err
	}

	if c.ReturnSourceDocuments {
		result[_retrievalQADefaultSourceDocumentKey] = docs
	}

	return result, nil
}

func (c RetrievalQA) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
}

func (c RetrievalQA) GetInputKeys() []string {
	return []string{c.InputKey}
}

func (c RetrievalQA) GetOutputKeys() []string {
	outputKeys := append([]string{}, c.CombineDocumentsChain.GetOutputKeys()...)
	if c.ReturnSourceDocuments {
		outputKeys = append(outputKeys, _retrievalQADefaultSourceDocumentKey)
	}

	return outputKeys
}
