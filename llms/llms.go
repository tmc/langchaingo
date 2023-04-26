// Package llms defines the types for langchaingo LLMs.
package llms

import "context"

// LLM is a langchaingo Large Language Model.
type LLM interface {
	Call(ctx context.Context, prompt string, stopWords []string) (string, error)
	Generate(ctx context.Context, prompts []string, stopWords []string) ([]*Generation, error)
}

// Generation is a single generation from a langchaingo LLM.
type Generation struct {
	// Text is the generated text.
	Text string `json:"text"`
	// GenerationInfo is the generation info. This can contain vendor-specific information.
	GenerationInfo map[string]any `json:"generation_info"`
}
