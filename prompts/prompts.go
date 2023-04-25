package prompts

import "github.com/tmc/langchaingo/schema"

// Formatter is an interface for formatting a map of values into a string.
type Formatter interface {
	Format(values map[string]any) (string, error)
}

// FormatPrompter is an interface for formatting a map of values into a prompt.
type FormatPrompter interface {
	FormatPrompt(values map[string]any) (schema.PromptValue, error)
}
