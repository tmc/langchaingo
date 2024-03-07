package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"golang.org/x/exp/maps"
)

// MapReduceDocuments is a chain that combines documents by mapping a chain over them, then
// combining the results using another chain.
type MapReduceDocuments struct {
	// The chain to apply to each documents individually.
	LLMChain *LLMChain

	// The chain to combine the mapped results of the LLMChain.
	ReduceChain Chain

	// The memory of the chain.
	Memory schema.Memory

	// The variable name of where to put the results from the LLMChain into the collapse chain.
	// Only needed if the reduce chain has more then one expected input.
	ReduceDocumentVariableName string

	// The input variable where the documents should be placed int the LLMChain. Only needed if
	// the reduce chain has more then one expected input.
	LLMChainInputVariableName string

	// MaxNumberOfConcurrent represents the max number of concurrent calls done simultaneously to
	// the llm chain.
	MaxNumberOfConcurrent int

	// The input key where the documents to be combined should be.
	InputKey string

	// Whether to add the intermediate steps to the output.
	ReturnIntermediateSteps bool
}

var _ Chain = MapReduceDocuments{}

// NewMapReduceDocuments creates a new map reduce documents chain with some default values.
func NewMapReduceDocuments(llmChain *LLMChain, reduceChain Chain) MapReduceDocuments {
	return MapReduceDocuments{
		LLMChain:                   llmChain,
		ReduceChain:                reduceChain,
		Memory:                     memory.NewSimple(),
		ReduceDocumentVariableName: _combineDocumentsDefaultInputKey,
		LLMChainInputVariableName:  _combineDocumentsDefaultDocumentVariableName,
		MaxNumberOfConcurrent:      _defaultApplyMaxNumberWorkers,
		InputKey:                   _combineDocumentsDefaultInputKey,
	}
}

// Call handles the inner logic of the MapReduceDocuments documents chain.
func (c MapReduceDocuments) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint:lll
	// Get the documents from the input values.
	docs, ok := values[c.InputKey].([]schema.Document)
	if !ok {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
	}

	// Execute the chain with each of the documents asynchronously.
	mapResults, err := Apply(ctx, c.LLMChain, c.getApplyInputs(values, docs), c.MaxNumberOfConcurrent, options...)
	if err != nil {
		return nil, err
	}

	// Create a document for each of map results and create input values for each of the document.
	reduceInputs, err := c.mapResultsToReduceInputs(docs, mapResults, values)
	if err != nil {
		return nil, err
	}

	result, err := Call(ctx, c.ReduceChain, reduceInputs, options...)
	return c.maybeAddIntermediateSteps(result, mapResults), err
}

// If the LLMChain or the reduce chain only has one input variable, it will be used to place the
// input automatically.
func (c MapReduceDocuments) getInputVariable(givenInputName string, chainInputVariables []string) string {
	if len(chainInputVariables) == 1 {
		return chainInputVariables[0]
	}

	return givenInputName
}

func (c MapReduceDocuments) maybeAddIntermediateSteps(result map[string]any, intermediateSteps []map[string]any) map[string]any { //nolint:lll
	if !c.ReturnIntermediateSteps {
		return result
	}

	result[_intermediateStepsOutputKey] = intermediateSteps
	return result
}

func (c MapReduceDocuments) getApplyInputs(values map[string]any, docs []schema.Document) []map[string]any {
	llmChainInputVariable := c.getInputVariable(c.LLMChainInputVariableName, c.LLMChain.GetInputKeys())
	inputs := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		curInput := c.copyInputValuesWithoutInputKey(values)
		curInput[llmChainInputVariable] = d.PageContent
		inputs = append(inputs, curInput)
	}

	return inputs
}

func (c MapReduceDocuments) mapResultsToReduceInputs(
	docs []schema.Document,
	mapResults []map[string]any,
	inputValues map[string]any,
) (map[string]any, error) {
	resultDocs := make([]schema.Document, 0, len(docs))
	for i := 0; i < len(docs); i++ {
		curResult, ok := mapResults[i][c.LLMChain.OutputKey].(string)
		if !ok {
			return nil, ErrInvalidOutputValues
		}

		resultDocs = append(resultDocs, schema.Document{
			PageContent: curResult,
			Metadata:    docs[i].Metadata,
		})
	}

	documentInputVariable := c.getInputVariable(c.ReduceDocumentVariableName, c.ReduceChain.GetInputKeys())
	reduceInputs := c.copyInputValuesWithoutInputKey(inputValues)
	reduceInputs[documentInputVariable] = resultDocs

	return reduceInputs, nil
}

func (c MapReduceDocuments) copyInputValuesWithoutInputKey(inputValues map[string]any) map[string]any {
	inputValuesCopy := make(map[string]any)
	maps.Copy(inputValuesCopy, inputValues)
	delete(inputValuesCopy, c.InputKey)
	return inputValuesCopy
}

func (c MapReduceDocuments) GetInputKeys() []string {
	inputKeys := map[string]bool{c.InputKey: true}
	for _, key := range c.LLMChain.GetInputKeys() {
		if key == c.LLMChainInputVariableName {
			continue
		}
		inputKeys[key] = true
	}

	for _, key := range c.ReduceChain.GetInputKeys() {
		if key == c.ReduceDocumentVariableName {
			continue
		}
		inputKeys[key] = true
	}

	return maps.Keys(inputKeys)
}

func (c MapReduceDocuments) GetOutputKeys() []string {
	outputKeys := c.ReduceChain.GetOutputKeys()
	if c.ReturnIntermediateSteps {
		outputKeys = append(outputKeys, _intermediateStepsOutputKey)
	}

	return outputKeys
}

func (c MapReduceDocuments) GetMemory() schema.Memory { //nolint:ireturn
	return c.Memory
}
