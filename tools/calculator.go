package tools

import (
	"context"

	"go.starlark.net/lib/math"
	"go.starlark.net/starlark"
)

// Calculator is a tool that can do math.
type Calculator struct{}

var _ Tool = Calculator{}

// Description returns a string describing the calculator tool.
func (c Calculator) Description() string {
	return `Useful for getting the result of a math expression. 
	The input to this tool should be a valid mathematical expression that could be executed by a starlark evaluator.`
}

// Name returns the name of the tool.
func (c Calculator) Name() string {
	return "calculator"
}

// Call evaluates the input using a starlak evaluator and returns the
// result as a string.
func (c Calculator) Call(_ context.Context, input string) (string, error) {
	v, err := starlark.Eval(&starlark.Thread{Name: "main"}, "input", input, math.Module.Members)
	if err != nil {
		return "I don't know how to do that.", nil
	}

	return v.String(), nil
}
