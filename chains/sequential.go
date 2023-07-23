package chains

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/internal/util"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

// SequentialChain is a chain that runs multiple chains in sequence,
// where the output of one chain is the input of the next.
type SequentialChain struct {
	chains     []Chain
	inputKeys  []string
	outputKeys []string
	memory     schema.Memory
}

func NewSequentialChain(chains []Chain, inputKeys []string, outputKeys []string) (*SequentialChain, error) {
	if err := validateSeqChain(chains, inputKeys, outputKeys); err != nil {
		return nil, err
	}

	return &SequentialChain{
		chains:     chains,
		inputKeys:  inputKeys,
		outputKeys: outputKeys,
		memory:     memory.NewSimple(),
	}, nil
}

func validateSeqChain(chain []Chain, inputKeys []string, outputKeys []string) error {
	knownKeys := util.ToSet(inputKeys)

	for i, c := range chain {
		// Check that chain has input keys that are in knownKeys
		missingKeys := util.Difference(c.GetInputKeys(), knownKeys)
		if len(missingKeys) > 0 {
			return fmt.Errorf(
				"%w: chain at index %d is missing required input keys: %v",
				ErrChainInitialization, i, missingKeys,
			)
		}

		// Check that chain does not have output keys that are already in knownKeys
		overlappingKeys := util.Intersection(c.GetOutputKeys(), knownKeys)
		if len(overlappingKeys) > 0 {
			return fmt.Errorf(
				"%w: chain at index %d has output keys that already exist: %v",
				ErrChainInitialization, i, overlappingKeys,
			)
		}

		// Add the chain's output keys to knownKeys
		for _, key := range c.GetOutputKeys() {
			knownKeys[key] = struct{}{}
		}
	}

	// Check that outputKeys are in knownKeys
	for _, key := range outputKeys {
		if _, ok := knownKeys[key]; !ok {
			return fmt.Errorf("%w: output key %s is not in the known keys", ErrChainInitialization, key)
		}
	}

	return nil
}

// Call runs the logic of the chains and returns the outputs. This method should
// not be called directly. Use rather the Call, Run or Predict functions that
// handles the memory and other aspects of the chain.
func (c *SequentialChain) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint:lll
	var outputs map[string]any
	var err error
	for _, chain := range c.chains {
		outputs, err = Call(ctx, chain, inputs, options...)
		if err != nil {
			return nil, err
		}
		// Set the input for the next chain to the output of the current chain
		inputs = outputs
	}
	return outputs, nil
}

// GetMemory gets the memory of the chain.
func (c *SequentialChain) GetMemory() schema.Memory {
	return c.memory
}

// InputKeys returns the input keys the chain expects.
func (c *SequentialChain) GetInputKeys() []string {
	return c.inputKeys
}

// OutputKeys returns the output keys the chain returns.
func (c *SequentialChain) GetOutputKeys() []string {
	return c.outputKeys
}

const (
	input  = "input"
	output = "output"
)

var (
	ErrInvalidInputNumberInSimpleSeq  = errors.New("single input expected for chains supplied to SimpleSequentialChain")
	ErrInvalidOutputNumberInSimpleSeq = errors.New("single output expected for chains supplied to SimpleSequentialChain")
)

// SimpleSequentialChain is a chain that runs multiple chains in sequence,
// where the output of one chain is the input of the next.
// All the chains must have a single input and a single output.
type SimpleSequentialChain struct {
	chains []Chain
	memory schema.Memory
}

func NewSimpleSequentialChain(chains []Chain) (*SimpleSequentialChain, error) {
	if err := validateSimpleSeq(chains); err != nil {
		return nil, err
	}

	return &SimpleSequentialChain{chains: chains, memory: memory.NewSimple()}, nil
}

func validateSimpleSeq(chains []Chain) error {
	for i, chain := range chains {
		if len(chain.GetInputKeys()) != 1 {
			return fmt.Errorf(
				"%w: chain at index [%d] has input keys: %v",
				ErrInvalidInputNumberInSimpleSeq, i, chain.GetInputKeys(),
			)
		}

		if len(chain.GetOutputKeys()) != 1 {
			return fmt.Errorf(
				"%w: chain at index [%d] has output keys: %v",
				ErrInvalidOutputNumberInSimpleSeq, i, chain.GetOutputKeys(),
			)
		}
	}
	return nil
}

// Call runs the logic of the chains and returns the output.
// This method should not be called directly.
// Use the Run function that handles the memory and other aspects of the chain.
func (c *SimpleSequentialChain) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint:lll
	input := inputs[input]
	for _, chain := range c.chains {
		var err error
		input, err = Run(ctx, chain, input, options...)
		if err != nil {
			return nil, err
		}
	}
	return map[string]any{output: input}, nil
}

// GetMemory gets the memory of the chain.
func (c *SimpleSequentialChain) GetMemory() schema.Memory {
	return c.memory
}

// InputKeys returns the input keys of the first chain.
func (c *SimpleSequentialChain) GetInputKeys() []string {
	return []string{input}
}

// OutputKeys returns the output keys of the last chain.
func (c *SimpleSequentialChain) GetOutputKeys() []string {
	return []string{output}
}
