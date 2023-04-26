package chains

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// Cain is the interface all chains must implement.
type Chain interface {
	Call(context.Context, map[string]any) (map[string]any, error)
	GetMemory() schema.Memory
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

	outputValues, err := c.Call(ctx, fullValues)
	if err != nil {
		return nil, err
	}

	err = c.GetMemory().SaveContext(inputValues, outputValues)
	if err != nil {
		return nil, err
	}

	return outputValues, nil
}
