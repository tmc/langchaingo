// Package fake provides a canned-response implementation of llms.Model
// for tests, in the spirit of testing/fstest's MapFS.
package fake

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/llms"
)

type LLM struct {
	responses []string
	index     int
}

func NewFakeLLM(responses []string) *LLM {
	return &LLM{
		responses: responses,
		index:     0,
	}
}

// GenerateContent returns the next configured response. It honors
// context cancellation and, when llms.WithStreamingFunc is set, streams
// the response to the callback before returning it; a callback error
// aborts the request.
func (f *LLM) GenerateContent(ctx context.Context, _ []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(f.responses) == 0 {
		return nil, errors.New("no responses configured")
	}
	if f.index >= len(f.responses) {
		f.index = 0 // reset index
	}
	response := f.responses[f.index]
	f.index++

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}
	if opts.StreamingFunc != nil {
		if err := opts.StreamingFunc(ctx, []byte(response)); err != nil {
			return nil, err
		}
	}
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: response}},
	}, nil
}

// Call  the model with a prompt.
func (f *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	resp, err := f.GenerateContent(ctx, []llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: prompt}}}}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) < 1 {
		return "", errors.New("empty response from model")
	}
	return resp.Choices[0].Content, nil
}

// Reset the index to 0.
func (f *LLM) Reset() {
	f.index = 0
}

// AddResponse adds a response to the list of responses.
func (f *LLM) AddResponse(response string) {
	f.responses = append(f.responses, response)
}
