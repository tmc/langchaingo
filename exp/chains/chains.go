package chains

import "github.com/tmc/langchaingo/schema"

type Chain interface {
	Call(map[string]any) (map[string]any, error)
	GetMemory() schema.Memory
}

func Call(c Chain, inputValues map[string]any) (map[string]any, error) {
	fullValues := make(map[string]any, 0)

	for key, value := range inputValues {
		fullValues[key] = value
	}

	newValues := c.GetMemory().LoadMemoryVariables(inputValues)

	for key, value := range newValues {
		fullValues[key] = value
	}

	outputValues, err := c.Call(fullValues)
	if err != nil {
		return map[string]any{}, err
	}

	err = c.GetMemory().SaveContext(inputValues, outputValues)
	if err != nil {
		return map[string]any{}, err
	}

	return outputValues, nil
}
