package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
)

// Chain is the interface all chains must implement.
type Chain interface {
	// Call runs the logic of the chain and returns the output. This method should
	// not be called directly. Use rather the Call, Run or Predict functions that
	// handles the memory and other aspects of the chain.
	Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error)
	// GetMemory gets the memory of the chain.
	GetMemory() schema.Memory
	// InputKeys returns the input keys the chain expects.
	GetInputKeys() []string
	// OutputKeys returns the output keys the chain returns.
	GetOutputKeys() []string
}

// Call is the standard function used for executing chains.
func Call(ctx context.Context, c Chain, inputValues map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint: lll
	fullValues := make(map[string]any, 0)
	for key, value := range inputValues {
		fullValues[key] = value
	}

	newValues, err := c.GetMemory().LoadMemoryVariables(inputValues)
	if err != nil {
		return nil, err
	}

	for key, value := range newValues {
		fullValues[key] = value
	}

	if err := validateInputs(c, fullValues); err != nil {
		return nil, err
	}

	outputValues, err := c.Call(ctx, fullValues, options...)
	if err != nil {
		return nil, err
	}
	if err := validateOutputs(c, outputValues); err != nil {
		return nil, err
	}

	err = c.GetMemory().SaveContext(inputValues, outputValues)
	if err != nil {
		return nil, err
	}

	return outputValues, nil
}

// Run can be used to execute a chain if the chain only expects one input and one
// string output.
func Run(ctx context.Context, c Chain, input any, options ...ChainCallOption) (string, error) {
	inputKeys := c.GetInputKeys()
	memoryKeys := c.GetMemory().MemoryVariables()
	neededKeys := make([]string, 0, len(inputKeys))

	// Remove keys gotten from the memory.
	for _, inputKey := range inputKeys {
		isInMemory := false
		for _, memoryKey := range memoryKeys {
			if inputKey == memoryKey {
				isInMemory = true
				continue
			}
		}
		if isInMemory {
			continue
		}
		neededKeys = append(neededKeys, inputKey)
	}
	if len(neededKeys) != 1 {
		return "", ErrMultipleInputsInRun
	}

	outputKeys := c.GetOutputKeys()
	if len(outputKeys) != 1 {
		return "", ErrMultipleOutputsInRun
	}

	inputValues := map[string]any{neededKeys[0]: input}
	outputValues, err := Call(ctx, c, inputValues, options...)
	if err != nil {
		return "", err
	}

	outputValue, ok := outputValues[outputKeys[0]].(string)
	if !ok {
		return "", ErrWrongOutputTypeInRun
	}

	return outputValue, nil
}

// Predict can be used to execute a chain if the chain only expects one string output.
func Predict(ctx context.Context, c Chain, inputValues map[string]any, options ...ChainCallOption) (string, error) {
	outputValues, err := Call(ctx, c, inputValues, options...)
	if err != nil {
		return "", err
	}

	outputKeys := c.GetOutputKeys()
	if len(outputKeys) != 1 {
		return "", ErrMultipleOutputsInPredict
	}

	outputValue, ok := outputValues[outputKeys[0]].(string)
	if !ok {
		return "", ErrOutputNotStringInPredict
	}

	return outputValue, nil
}

const _defaultApplyMaxNumberWorkers = 5

// Apply executes the chain for each of the inputs asynchronously.
func Apply(ctx context.Context, c Chain, inputValues []map[string]any, maxWorkers int, options ...ChainCallOption) ([]map[string]any, error) { // nolint:lll
	if maxWorkers <= 0 {
		maxWorkers = _defaultApplyMaxNumberWorkers
	}

	inputJobs := make(chan map[string]any, len(inputValues))
	resultsChan := make(chan struct {
		result map[string]any
		err    error
	}, len(inputValues))
	defer close(inputJobs)
	defer close(resultsChan)

	for w := 0; w < maxWorkers; w++ {
		go func() {
			for input := range inputJobs {
				res, err := Call(ctx, c, input, options...)
				resultsChan <- struct {
					result map[string]any
					err    error
				}{
					result: res,
					err:    err,
				}
			}
		}()
	}

	for _, input := range inputValues {
		inputJobs <- input
	}

	results := make([]map[string]any, len(inputValues))
	for i := range inputValues {
		r := <-resultsChan
		if r.err != nil {
			return nil, r.err
		}
		results[i] = r.result
	}

	return results, nil
}

func validateInputs(c Chain, inputValues map[string]any) error {
	for _, k := range c.GetInputKeys() {
		if _, ok := inputValues[k]; !ok {
			return fmt.Errorf("%w: %w: %v", ErrInvalidInputValues, ErrMissingInputValues, k)
		}
	}
	return nil
}

func validateOutputs(c Chain, outputValues map[string]any) error {
	for _, k := range c.GetOutputKeys() {
		if _, ok := outputValues[k]; !ok {
			return fmt.Errorf("%w: %v", ErrInvalidOutputValues, k)
		}
	}
	return nil
}
