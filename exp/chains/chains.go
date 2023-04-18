package chains

import "github.com/tmc/langchaingo/exp/memory"

type ChainValues map[string]any

type Chain interface {
	Call(cv ChainValues, stop []string) (ChainValues, error)
	GetMemory() memory.Memory
}

func Call(c Chain, inputValues map[string]any, stop []string) (ChainValues, error) {
	fullValues := make(ChainValues, 0)

	for key, value := range inputValues {
		fullValues[key] = value
	}

	newValues, err := c.GetMemory().LoadMemoryVariables(inputValues)
	if err != nil {
		return ChainValues{}, err
	}

	for key, value := range newValues {
		fullValues[key] = value
	}

	outputValues, err := c.Call(fullValues, stop)
	if err != nil {
		return ChainValues{}, err
	}

	err = c.GetMemory().SaveContext(inputValues, memory.InputValues(outputValues))
	if err != nil {
		return ChainValues{}, err
	}

	return outputValues, nil
}
