package llms

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/schema"
)

// LLM is an alias for model, for backwards compatibility.
//
// Deprecated: This alias may be removed in the future; please use Model
// instead.
type LLM = Model

// Model is an interface multi-modal models implement.
type Model interface {
	// GenerateContent asks the model to generate content from a sequence of
	// messages. It's the most general interface for multi-modal LLMs that support
	// chat-like interactions.
	GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)

	// Call is a simplified interface for a text-only Model, generating a single
	// string response from a single string prompt.
	//
	// It is here for backwards compatibility only and may be removed
	// in the future; please use GenerateContent instead.
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)
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
