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
type RetrievalQA struct {
	Retriever             schema.Retriever
	CombineDocumentsChain Chain

	InputKey              string
	ReturnSourceDocuments bool
}

var _ Chain = RetrievalQA{}

// NewRetrievalQA creates a new RetrievalQA from a retriever and a chain for
// combining documents. The chain for combining documents is expected to only
// have the expected input values "question" and "input_documents".
func NewRetrievalQA(combineDocumentsChain Chain, retriever schema.Retriever) RetrievalQA {
	return RetrievalQA{
		Retriever:             retriever,
		CombineDocumentsChain: combineDocumentsChain,
		InputKey:              _retrievalQADefaultInputKey,
		ReturnSourceDocuments: false,
	}
}

// NewRetrievalQAFromLLM loads a combine documents chain from the llm and
// creates a new retrievalQA chain.
func NewRetrievalQAFromLLM(llm llms.LLM, retriever schema.Retriever) RetrievalQA {
	return NewRetrievalQA(
		LoadQAStuffChain(llm),
		retriever,
	)
}

// Call gets relevant documents from the retriever and gives them to the combine
// documents chain.
func (c RetrievalQA) Call(ctx context.Context, values map[string]any) (map[string]any, error) {
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
	})
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
