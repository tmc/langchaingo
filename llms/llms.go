// Package llms defines the types for langchaingo LLMs.
package llms

// LLM is a langchaingo Large Language Model.
type LLM interface {
	Call(prompt string) (string, error)
	Generate(prompts []string) ([]*Generation, error)
}

// Generation is a single generation from a langchaingo LLM.
type Generation struct {
	// Text is the generated text.
	Text string `json:"text"`
	// GenerationInfo is the generation info.
	GenerationInfo map[string]any `json:"generation_info"`
}
