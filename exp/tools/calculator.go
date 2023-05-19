package tools

import (
	"go.starlark.net/lib/math"
	"go.starlark.net/starlark"
)

type Calculator struct{}

var _ Tool = Calculator{}

func (c Calculator) Description() string {
	return "Useful for getting the result of a math expression. The input to this tool should be a valid mathematical expression that could be executed by a starlark evaluator." //nolint:lll
}

func (c Calculator) Name() string {
	return "calculator"
}

func (c Calculator) Call(input string) (string, error) {
	v, err := starlark.Eval(&starlark.Thread{Name: "main"}, "input", input, math.Module.Members)
	if err != nil {
		return "I don't know how to do that.", nil
	}

	return v.String(), nil
}
