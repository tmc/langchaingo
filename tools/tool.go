package tools

import (
	"context"
	"errors"
)

const ErrInvalidTool = "invalid_tool"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
}

type Kit []Tool

func (tb *Kit) UseTool(ctx context.Context, toolName string, toolArgs string) (string, error) {
	for _, tool := range *tb {
		if tool.Name() == toolName {
			return tool.Call(ctx, toolArgs)
		}
	}
	return "", errors.New(ErrInvalidTool)
}
