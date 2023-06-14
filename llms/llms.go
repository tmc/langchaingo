// Package llms defines the types for langchaingo LLMs.
package llms

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// LanguageModel is the base class for all language models.
type LanguageModel interface {
	// Take in a list of prompt values and return an LLMResult.
	GeneratePrompt(ctx context.Context, prompts []schema.PromptValue, options ...CallOption) (LLMResult, error)
	// Get the number of tokens present in the text.
	GetNumTokens(text string) int
}

// LLM is a langchaingo Large Language Model.
type LLM interface {
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)
	Generate(ctx context.Context, prompts []string, options ...CallOption) ([]*Generation, error)
}

// ChatLLM is a langchaingo LLM that could be used for chatting.
type ChatLLM interface {
	Call(ctx context.Context, messages []schema.ChatMessage, options ...CallOption) (string, error)
	Generate(ctx context.Context, messages [][]schema.ChatMessage, options ...CallOption) ([]*Generation, error)
}

// Generation is a single generation from a langchaingo LLM.
type Generation struct {
	// Text is the generated text.
	Text string `json:"text"`
	// Message stores the potentially generated message.
	Message *schema.AIChatMessage `json:"message"`
	// GenerationInfo is the generation info. This can contain vendor-specific information.
	GenerationInfo map[string]any `json:"generation_info"`
}

// LLMResult is the class that contains all relevant information for an LLM Result.
type LLMResult struct {
	Generations [][]*Generation
	LLMOutput   map[string]any
}
