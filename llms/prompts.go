package llms

// PromptValue is the interface that all prompt values must implement.
type PromptValue interface {
	String() string
	Messages() []ChatMessage
}
