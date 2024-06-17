package tools

import "context"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Schema() any
	Call(ctx context.Context, input any) (string, error)
}
