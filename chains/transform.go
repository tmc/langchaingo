package chains

import (
	"context"

	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

// TransformFunc is the function type that the transform chain uses.
type TransformFunc func(context.Context, map[string]any, ...ChainCallOption) (map[string]any, error)

// Transform is a chain that runs an arbitrary function.
type Transform struct {
	Memory     schema.Memory
	Transform  TransformFunc
	InputKeys  []string
	OutputKeys []string
}

var _ Chain = Transform{}

// NewTransform creates a new transform chain with the function to use, the
// expected input and output variables.
func NewTransform(f TransformFunc, inputVariables []string, outputVariables []string) Transform {
	return Transform{
		Memory:     memory.NewSimple(),
		Transform:  f,
		InputKeys:  inputVariables,
		OutputKeys: outputVariables,
	}
}

// Call returns the output of the transform function.
func (c Transform) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) { //nolint:lll
	return c.Transform(ctx, inputs, options...)
}

// GetMemory gets the memory of the chain.
func (c Transform) GetMemory() schema.Memory {
	return c.Memory
}

// GetInputKeys returns the input keys the chain expects.
func (c Transform) GetInputKeys() []string {
	return c.InputKeys
}

// GetOutputKeys returns the output keys the chain returns.
func (c Transform) GetOutputKeys() []string {
	return c.OutputKeys
}
