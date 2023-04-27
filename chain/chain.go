package chain

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/schema"
)

var (
	// ErrInvalidInputValues is returned when some expected input values keys to
	// a chain is missing.
	ErrInvalidInputValues = errors.New("missing keys in input values")
	// ErrInvalidOutputValues is returned when expected output keys to a chain does
	// not match the actual keys in the return output values map.
	ErrInvalidOutputValues = errors.New("missing keys in output values")
)

// Cain is the interface all chains must implement.
type Chain interface {
	// Call runs the logic of the chain and returns the output.
	Call_(context.Context, map[string]any) (map[string]any, error)
	// GetMemory gets the memory of the chain.
	GetMemory() schema.Memory
	// InputKeys returns the input keys the chain expects.
	GetInputKeys() []string
	// OutputKeys returns the output keys the chain expects.
	GetOutputKeys() []string
}

// Call is the function used for calling chains.
func Call(ctx context.Context, c Chain, inputValues map[string]any) (map[string]any, error) {
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

	outputValues, err := c.Call_(ctx, fullValues)
	if err != nil {
		return nil, err
	}

	err = c.GetMemory().SaveContext(inputValues, outputValues)
	if err != nil {
		return nil, err
	}

	return outputValues, nil
}

// Run can be used if the chain only expects one string input and  one output
// string.
func Run(ctx context.Context, c Chain, input string) string

func validateInputs(c Chain, inputValues map[string]any) error {
	for _, k := range c.GetInputKeys() {
		if _, ok := inputValues[k]; !ok {
			return ErrInvalidInputValues
		}
	}
	return nil
}

func validateOutputs(c Chain, outputValues map[string]any) error {
	for _, k := range c.GetInputKeys() {
		if _, ok := outputValues[k]; !ok {
			return ErrInvalidOutputValues
		}
	}
	return nil
}
