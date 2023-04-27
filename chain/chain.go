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

	// ErrMultipleInputsInRun is returned in the run function if the chain expects
	// more then one input values.
	ErrMultipleInputsInRun = errors.New("run not supported in chain with more then one expected input")
	// ErrMultipleOutputsInRun is returned in the run function if the chain expects
	// more then one output values.
	ErrMultipleOutputsInRun = errors.New("run not supported in chain with more then one expected output")
	// ErrMultipleOutputsInRun is returned in the run function if the chain returns
	// a value that is not a string.
	ErrWrongOutputTypeInRun = errors.New("run not supported in chain that returns value that is not string")
)

// Cain is the interface all chains must implement.
type Chain interface {
	// Call runs the logic of the chain and returns the output. This method should
	// not be called directly. Use rather the Call function that handles the memory
	// of the chain.
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
	if err := validateInputs(c, inputValues); err != nil {
		return nil, err
	}

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
	if err := validateOutputs(c, outputValues); err != nil {
		return nil, err
	}

	err = c.GetMemory().SaveContext(inputValues, outputValues)
	if err != nil {
		return nil, err
	}

	return outputValues, nil
}

// Run can be used to call a chain if the chain only expects one string input
// and one string output.
func Run(ctx context.Context, c Chain, input string) (string, error) {
	inputKeys := c.GetInputKeys()
	if len(inputKeys) != 1 {
		return "", ErrMultipleInputsInRun
	}

	outputKeys := c.GetOutputKeys()
	if len(outputKeys) != 1 {
		return "", ErrMultipleOutputsInRun
	}

	inputValues := map[string]any{inputKeys[0]: input}
	outputValues, err := Call(ctx, c, inputValues)
	if err != nil {
		return "", err
	}

	outputValue, ok := outputValues[outputKeys[0]].(string)
	if !ok {
		return "", ErrWrongOutputTypeInRun
	}

	return outputValue, nil
}

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
