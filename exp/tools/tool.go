// Package tools defines the types for tools to be used by the llms agents.
package tools

// Tool is a tool for the llm agent to interact with diferent application.
type Tool struct {
	Name        string
	Description string
	Run         func(string) string
}

func NewTool(name, description string, run func(string) string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		Run:         run,
	}
}
