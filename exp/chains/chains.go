package chains

import "github.com/tmc/langchaingo/schema"

// Interface all chains must implement
type Chain interface {
	Call(map[string]any) (map[string]any, error)
	GetMemory() schema.Memory
}

// Function handling the logic of calling chains
func Call(c Chain, inputValues map[string]any) (map[string]any, error) {
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

	outputValues, err := c.Call(fullValues)
	if err != nil {
		return nil, err
	}

	err = c.GetMemory().SaveContext(inputValues, outputValues)
	if err != nil {
		return nil, err
	}

	return outputValues, nil
}
