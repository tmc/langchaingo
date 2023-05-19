package tools

// Tool is a tool for the llm agent to interact with different application.
type Tool interface {
	Name() string
	Description() string
	Call(string) (string, error)
}
