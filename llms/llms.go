// Package llms defines the types for langchaingo LLMs.
package llms

// LLM is a langchaingo Large Language Model.
type LLM interface {
	Call(prompt string) (string, error)
	Generate(prompts []string) ([]string, error)
}
