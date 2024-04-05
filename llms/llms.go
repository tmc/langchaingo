package llms

import (
	"context"
	"errors"
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
	// Deprecated: this method is retained for backwards compatibility. Use the
	// more general [GenerateContent] instead. You can also use
	// the [GenerateFromSinglePrompt] function which provides a similar capability
	// to Call and is built on top of the new interface.
	Call(ctx context.Context, prompt string, options ...CallOption) (string, error)
}

// GenerateFromSinglePrompt is a convenience function for calling an LLM with
// a single string prompt, expecting a single string response. It's useful for
// simple, string-only interactions and provides a slightly more ergonomic API
// than the more general [llms.Model.GenerateContent].
func GenerateFromSinglePrompt(ctx context.Context, llm Model, prompt string, options ...CallOption) (string, error) {
	msg := MessageContent{
		Role:  ChatMessageTypeHuman,
		Parts: []ContentPart{TextContent{prompt}},
	}

	resp, err := llm.GenerateContent(ctx, []MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return "", errors.New("empty response from model")
	}
	c1 := choices[0]
	return c1.Content, nil
}
