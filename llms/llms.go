package llms

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// LLM is a langchaingo Large Language Model.
type LLM interface {
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)
	Generate(ctx context.Context, prompts []string, options ...CallOption) ([]*Generation, error)
}

// ChatLLM is a langchaingo LLM that can be used for chatting.
type ChatLLM interface {
	Chat(ctx context.Context, messages []schema.ChatMessage, options ...CallOption) (*ChatGeneration, error)
}

// Generation is a single generation from a langchaingo LLM.
type Generation struct {
	// Text is the generated text.
	Text string `json:"text"`
	// GenerationInfo is the generation info. This can contain vendor-specific information.
	GenerationInfo map[string]any `json:"generation_info"`
}

// ChatGeneration is a single generation from a langchaingo ChatLLM.
type ChatGeneration struct {
	Message *schema.AIChatMessage `json:"message"`
	// GenerationInfo is the generation info. This can contain vendor-specific information.
	GenerationInfo map[string]any `json:"generation_info"`
}
