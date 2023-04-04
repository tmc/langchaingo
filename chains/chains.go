package chains

import "github.com/tmc/langchaingo/memory"

type ChainValues map[string]any

type Chain interface {
	Call(ChainValues) (ChainValues, error) //
	GetChainType() string
	GetMemory() memory.Memory
}

func Call(c Chain, values map[string]any) (ChainValues, error) {
	fullValues := make(ChainValues, 0)

	for key, value := range values {
		fullValues[key] = value
	}

	newValues, err := c.GetMemory().LoadMemoryVariables(values)
	if err != nil {
		return ChainValues{}, err
	}

	for key, value := range newValues {
		fullValues[key] = value
	}

	outputValues, err := c.Call(fullValues)
	if err != nil {
		return ChainValues{}, err
	}

	err = c.GetMemory().SaveContext(values, memory.InputValues(outputValues))
	if err != nil {
		return ChainValues{}, err
	}

	return outputValues, nil
}
