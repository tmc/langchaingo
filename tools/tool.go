package tools

import "context"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
}

// ToolWithParameters is an optional interface for tools that declare their own
// parameter schema for OpenAI function calling. Tools that do not implement it
// fall back to the legacy single-string "__arg1" schema.
type ToolWithParameters interface {
	Tool
	// Parameters returns the JSON Schema object describing the tool's
	// function-calling parameters, e.g.:
	//
	//	{
	//	  "type": "object",
	//	  "properties": {"city": {"type": "string", "description": "..."}},
	//	  "required": ["city"]
	//	}
	//
	// The schema is passed through verbatim to the LLM provider (e.g. OpenAI
	// function calling), so it must be a valid JSON Schema object with
	// "type": "object".
	Parameters() map[string]any
}
