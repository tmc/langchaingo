package chains

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
)

// TypedChain is a generic interface for chains that produce typed outputs.
// This is a proof-of-concept for the typed chain proposal.
type TypedChain[T any] interface {
	// Call runs the logic of the chain and returns typed output + metadata
	Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (T, map[string]any, error)

	// GetMemory gets the memory of the chain
	GetMemory() schema.Memory

	// GetInputKeys returns the input keys the chain expects
	GetInputKeys() []string

	// OutputType returns a description of the output type
	OutputType() string
}

// ChainAdapter adapts a TypedChain to the legacy Chain interface for backward compatibility.
type ChainAdapter[T any] struct {
	chain TypedChain[T]
}

// NewChainAdapter creates an adapter to make a TypedChain compatible with the legacy Chain interface.
func NewChainAdapter[T any](chain TypedChain[T]) *ChainAdapter[T] {
	return &ChainAdapter[T]{chain: chain}
}

func (a *ChainAdapter[T]) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) {
	result, metadata, err := a.chain.Call(ctx, inputs, options...)
	if err != nil {
		return nil, err
	}

	// Combine typed result with metadata
	output := make(map[string]any)
	for k, v := range metadata {
		output[k] = v
	}
	output["result"] = result

	return output, nil
}

func (a *ChainAdapter[T]) GetMemory() schema.Memory {
	return a.chain.GetMemory()
}

func (a *ChainAdapter[T]) GetInputKeys() []string {
	return a.chain.GetInputKeys()
}

func (a *ChainAdapter[T]) GetOutputKeys() []string {
	return []string{"result"} // Standard key for adapted chains
}

// TypedCall is a generic version of the Call function for TypedChains.
func TypedCall[T any](
	ctx context.Context,
	c TypedChain[T],
	inputValues map[string]any,
	options ...ChainCallOption,
) (T, map[string]any, error) {
	fullValues := make(map[string]any, 0)
	for key, value := range inputValues {
		fullValues[key] = value
	}

	newValues, err := c.GetMemory().LoadMemoryVariables(ctx, inputValues)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	for key, value := range newValues {
		fullValues[key] = value
	}

	// TODO: Add callback handling similar to existing Call function

	if err := validateTypedInputs(c, fullValues); err != nil {
		var zero T
		return zero, nil, err
	}

	result, metadata, err := c.Call(ctx, fullValues, options...)
	if err != nil {
		return result, metadata, err
	}

	// TODO: Add memory saving logic

	return result, metadata, nil
}

// TypedRun is a generic version of Run for TypedChains.
func TypedRun[T any](
	ctx context.Context,
	c TypedChain[T],
	input any,
	options ...ChainCallOption,
) (T, error) {
	inputKeys := c.GetInputKeys()
	if len(inputKeys) != 1 {
		var zero T
		return zero, fmt.Errorf("chain must have exactly one input key, got %d", len(inputKeys))
	}

	result, _, err := TypedCall(ctx, c, map[string]any{inputKeys[0]: input}, options...)
	return result, err
}

// TypedPredict is a convenience function for simple predictions.
func TypedPredict[T any](
	ctx context.Context,
	c TypedChain[T],
	values map[string]any,
	options ...ChainCallOption,
) (T, error) {
	result, _, err := TypedCall(ctx, c, values, options...)
	return result, err
}

// Helper function for input validation (similar to existing validateInputs)
func validateTypedInputs[T any](c TypedChain[T], values map[string]any) error {
	inputKeys := c.GetInputKeys()
	for _, key := range inputKeys {
		if _, ok := values[key]; !ok {
			return fmt.Errorf("missing required input key: %s", key)
		}
	}
	return nil
}

// Example typed chain implementations

// StringChain is a convenience type for chains that produce strings.
type StringChain = TypedChain[string]

// DocumentChain is a convenience type for chains that produce documents.
type DocumentChain = TypedChain[[]schema.Document]