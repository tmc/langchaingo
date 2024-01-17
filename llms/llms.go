package llms

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/schema"
)

// LLM is an alias for model, for backwards compatibility.
//
// This alias may be removed in the future; please use Model instead.
type LLM = Model

// Model is an interface multi-modal models implement.
// Note: this is an experimental API.
type Model interface {
	// Call is a simplified interface for Model, generating a single string
	// response from a single string prompt.
	//
	// It is here for backwards compatibility only and may be removed in the
	// future; please use GenerateContent instead.
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)

	// GenerateContent asks the model to generate content from a sequence of
	// messages. It's the most general interface for LLMs that support chat-like
	// interactions.
	GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)
}

// Generation is a single generation from a langchaingo LLM.
// This type may be removed in the future; please don't use in new code.
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
// This type may be removed in the future; please don't use in new code.
type LLMResult struct {
	Generations [][]*Generation
	LLMOutput   map[string]any
}

// CallLLM is a helper function for implementing Call in terms of
// GenerateContent. It's aimed to be used by Model providers.
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
