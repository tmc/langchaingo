package tools

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"go.starlark.net/lib/math"
	"go.starlark.net/starlark"
)

// Calculator is a tool that can do math.
type Calculator struct {
	CallbacksHandler callbacks.Handler
}

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

// Schema returns OpenAPI schema for the calculator tool.
func (c Calculator) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Calculator query to execute.",
			},
		},
		"required": []string{"query"},
	}
}

// Call validates the input returned by the LLM and calls the calculator tool.
func (c Calculator) Call(ctx context.Context, input any) (string, error) {
	query, ok := input.(map[string]any)["query"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: %v", input)
	}

	return c.call(ctx, query)
}

// call evaluates the input using a starlark evaluator and returns the result as a
// string. If the evaluator errors the error is given in the result to give the
// agent the ability to retry.
func (c Calculator) call(ctx context.Context, query string) (string, error) {
	if c.CallbacksHandler != nil {
		c.CallbacksHandler.HandleToolStart(ctx, query)
	}

	v, err := starlark.Eval(&starlark.Thread{Name: "main"}, "input", query, math.Module.Members)
	if err != nil {
		return fmt.Sprintf("error from evaluator: %s", err.Error()), nil //nolint:nilerr
	}
	result := v.String()

	if c.CallbacksHandler != nil {
		c.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}
