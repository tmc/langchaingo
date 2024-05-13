package tools

import "context"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input map[string]any) (string, error)
	Schema() Schema
}

type Schema struct {
	Type       string
	Properties map[string]any
	Required   []string
}
