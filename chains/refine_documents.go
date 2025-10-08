package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

const (
	_refineDocumentsDefaultDocumentTemplate    = "{{.page_content}}"
	_refineDocumentsDefaultInitialResponseName = "existing_answer"
)

// RefineDocuments is a chain used for combining and processing unstructured
// text data. The chain iterates over the documents one by one to update a
// running answer, at each turn using the previous version of the answer and
// the next document as context.
type RefineDocuments struct {
	// Chain used to construct the first text using the first document.
	LLMChain *LLMChain

	// Chain used to refine the first text using the additional documents.
	RefineLLMChain *LLMChain

	// Prompt to format the documents. Documents are given in the variable
	// with the name "page_content". All metadata from the documents are
	// also given to the prompt template.
	DocumentPrompt prompts.PromptTemplate

	InputKey             string
	OutputKey            string
	DocumentVariableName string
	InitialResponseName  string
}

var _ Chain = RefineDocuments{}

// NewRefineDocuments creates a new refine documents chain from the llm
// chain used to construct the initial text and the llm used to refine
// the text.
func NewRefineDocuments(initialLLMChain, refineLLMChain *LLMChain) RefineDocuments {
	return RefineDocuments{
		LLMChain:       initialLLMChain,
		RefineLLMChain: refineLLMChain,
		DocumentPrompt: prompts.NewPromptTemplate(
			_refineDocumentsDefaultDocumentTemplate,
			[]string{"page_content"},
		),
		InputKey:             _combineDocumentsDefaultInputKey,
		OutputKey:            _combineDocumentsDefaultOutputKey,
		DocumentVariableName: _combineDocumentsDefaultDocumentVariableName,
		InitialResponseName:  _refineDocumentsDefaultInitialResponseName,
	}
}

// Call handles the inner logic of the refine documents chain.
func (c RefineDocuments) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint:lll
	// Get the documents to be combined.
	docs, ok := values[c.InputKey].([]schema.Document)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("%w: documents slice has no elements", ErrInvalidInputValues)
	}

	// Get the rest of the input variables.
	rest := make(map[string]any, len(values))
	for key, value := range values {
		if key == c.InputKey {
			continue
		}
		rest[key] = value
	}

	// Create a text using the first document.
	initialInputs, err := c.constructInitialInputs(docs[0], rest)
	if err != nil {
		return nil, err
	}
	response, err := Predict(ctx, c.LLMChain, initialInputs, options...)
	if err != nil {
		return nil, err
	}

	// Refine the text using the rest of the documents.
	for i := 1; i < len(docs); i++ {
		refineInputs, err := c.constructRefineInputs(docs[i], response, rest)
		if err != nil {
			return nil, err
		}
		response, err = Predict(ctx, c.RefineLLMChain, refineInputs, options...)
		if err != nil {
			return nil, err
		}
	}

	return map[string]any{
		c.OutputKey: response,
	}, nil
}

func (c RefineDocuments) constructInitialInputs(doc schema.Document, rest map[string]any) (map[string]any, error) {
	return c.getBaseInputs(doc, rest)
}

func (c RefineDocuments) constructRefineInputs(doc schema.Document, lastResponse string, rest map[string]any) (map[string]any, error) { //nolint:lll
	inputs, err := c.getBaseInputs(doc, rest)
	if err != nil {
		return nil, err
	}

	inputs[c.InitialResponseName] = lastResponse
	return inputs, nil
}

// getBaseInputs formats the document given and adds the formatted document
// and the rest of the input variables to the inputs.
func (c RefineDocuments) getBaseInputs(doc schema.Document, rest map[string]any) (map[string]any, error) {
	var err error
	baseInfo := make(map[string]any, len(doc.Metadata)+1)
	baseInfo["page_content"] = doc.PageContent
	for key, value := range doc.Metadata {
		baseInfo[key] = value
	}

	documentInfo := make(map[string]any, 0)
	for _, promptVariable := range c.DocumentPrompt.InputVariables {
		if _, ok := baseInfo[promptVariable]; !ok {
			return nil, fmt.Errorf(
				"%w: document is missing metadata for %s used in the document prompt",
				ErrInvalidInputValues, promptVariable,
			)
		}
		documentInfo[promptVariable] = baseInfo[promptVariable]
	}

	inputs := make(map[string]any, len(rest))
	inputs[c.DocumentVariableName], err = c.DocumentPrompt.Format(documentInfo)
	if err != nil {
		return nil, err
	}

	for key, value := range rest {
		inputs[key] = value
	}
	return inputs, nil
}

func (c RefineDocuments) GetInputKeys() []string {
	inputKeys := []string{c.InputKey}
	for _, key := range c.LLMChain.GetInputKeys() {
		if key == c.DocumentVariableName {
			continue
		}
		inputKeys = append(inputKeys, key)
	}

	return inputKeys
}

func (c RefineDocuments) GetOutputKeys() []string {
	return []string{c.OutputKey}
}

func (c RefineDocuments) GetMemory() schema.Memory { //nolint:ireturn
	return memory.NewSimple()
}
