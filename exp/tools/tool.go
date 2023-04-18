// Package tools defines the types for tools to be used by the llms agents.
package tools

// Tool is a tool for the llm agent to interact with diferent application.
type Tool struct {
	Name string
	run  func(string) string
}
