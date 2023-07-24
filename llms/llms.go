package llms

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// LanguageModel is the interface all language models must implement.
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

// ChatLLM is a langchaingo LLM that can be used for chatting.
type ChatLLM interface {
	Call(ctx context.Context, messages []schema.ChatMessage, options ...CallOption) (*schema.AIChatMessage, error)
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

func GeneratePrompt(ctx context.Context, l LLM, promptValues []schema.PromptValue, options ...CallOption) (LLMResult, error) { //nolint:lll
	prompts := make([]string, 0, len(promptValues))
	for _, promptValue := range promptValues {
		prompts = append(prompts, promptValue.String())
	}
	generations, err := l.Generate(ctx, prompts, options...)
	return LLMResult{
		Generations: [][]*Generation{generations},
	}, err
}

func GenerateChatPrompt(ctx context.Context, l ChatLLM, promptValues []schema.PromptValue, options ...CallOption) (LLMResult, error) { //nolint:lll
	messages := make([][]schema.ChatMessage, 0, len(promptValues))
	for _, promptValue := range promptValues {
		messages = append(messages, promptValue.Messages())
	}
	generations, err := l.Generate(ctx, messages, options...)
	return LLMResult{
		Generations: [][]*Generation{generations},
	}, err
}
