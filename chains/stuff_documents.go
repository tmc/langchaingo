package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

const (
	_combineDocumentsDefaultInputKey             = "input_documents"
	_combineDocumentsDefaultOutputKey            = "text"
	_combineDocumentsDefaultDocumentVariableName = "context"
	_stuffDocumentsDefaultSeparator              = "\n\n"
)

// StuffDocuments is a chain that combines documents with a separator and uses
// the stuffed documents in an LLMChain. The input values to the llm chain
// contains all input values given to this chain, and the stuffed document as
// a string in the key specified by the "DocumentVariableName" field that is
// by default set to "context".
type StuffDocuments struct {
	// LLMChain is the LLMChain called after formatting the documents.
	LLMChain *LLMChain

	// Input key is the input key the StuffDocuments chain expects the
	//  documents to be in.
	InputKey string

	// DocumentVariableName is the variable name used in the llm_chain to put
	// the documents in.
	DocumentVariableName string

	// Separator is the string used to join the documents.
	Separator string
}

var _ Chain = StuffDocuments{}

// NewStuffDocuments creates a new stuff documents chain with an LLM chain used
// after formatting the documents.
func NewStuffDocuments(llmChain *LLMChain) StuffDocuments {
	return StuffDocuments{
		LLMChain: llmChain,

		InputKey:             _combineDocumentsDefaultInputKey,
		DocumentVariableName: _combineDocumentsDefaultDocumentVariableName,
		Separator:            _stuffDocumentsDefaultSeparator,
	}
}

// Call handles the inner logic of the StuffDocuments chain.
func (c StuffDocuments) Call(
	ctx context.Context, values map[string]any, options ...ChainCallOption,
) (map[string]any, error) {
	docs, ok := values[c.InputKey].([]schema.Document)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	inputValues := make(map[string]any)
	for key, value := range values {
		inputValues[key] = value
	}

	inputValues[c.DocumentVariableName] = c.joinDocuments(docs)
	return Call(ctx, c.LLMChain, inputValues, options...)
}

// GetMemory returns a simple memory.
func (c StuffDocuments) GetMemory() schema.Memory {
	return memory.NewSimple()
}

// GetInputKeys returns the expected input keys, by default "input_documents".
func (c StuffDocuments) GetInputKeys() []string {
	return []string{c.InputKey}
}

// GetOutputKeys returns the output keys the chain will return.
func (c StuffDocuments) GetOutputKeys() []string {
	return append([]string{}, c.LLMChain.GetOutputKeys()...)
}

// joinDocuments joins the documents with the separator.
func (c StuffDocuments) joinDocuments(docs []schema.Document) string {
	var text string
	docLen := len(docs)
	for k, doc := range docs {
		text += doc.PageContent
		if k != docLen-1 {
			text += c.Separator
		}
	}
	return text
}
