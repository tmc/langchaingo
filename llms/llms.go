package llms

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/schema"
)

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

// Model is an interface multi-modal models implement.
// Note: this is an experimental API.
type Model interface {
	// GenerateContent asks the model to generate content from a sequence of
	// messages. It's the most general interface for LLMs that support chat-like
	// interactions.
	GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)
}

// Generation is a single generation from a langchaingo LLM.
type Generation struct {
	// Text is the generated text.
	Text string `json:"text"`
	// Message stores the potentially generated message.
	Message *schema.AIChatMessage `json:"message"`
	// GenerationInfo is the generation info. This can contain vendor-specific information.
	GenerationInfo map[string]any `json:"generation_info"`
	// StopReason is the reason the generation stopped.
	StopReason string `json:"stop_reason"`
}

// LLMResult is the class that contains all relevant information for an LLM Result.
type LLMResult struct {
	Generations [][]*Generation
	LLMOutput   map[string]any
}

func CallLLM(ctx context.Context, llm Model, prompt string, options ...CallOption) (string, error) {
	msg := MessageContent{
		Role:  schema.ChatMessageTypeHuman,
		Parts: []ContentPart{TextContent{prompt}},
	}

	resp, err := llm.GenerateContent(ctx, []MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return "", errors.New("empty response from model") //nolint:goerr113
	}
	c1 := choices[0]
	return c1.Content, nil
}
