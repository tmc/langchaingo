package tools

import "context"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Call(context.Context, string) (string, error)
}
