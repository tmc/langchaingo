package tools

import "context"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input map[string]any) (string, error)

	// Schema returns the schema for the tool.
	Schema() map[string]any
}
