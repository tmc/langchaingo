package openaiclient

import (
	"context"
	"errors"
)

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the OpenAI API.
type Client struct {
	token string
}

// New returns a new OpenAI client.
func New(token string) (*Client, error) {
	c := &Client{token: token}
	return c, nil
}

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	resp, err := c.createCompletion(ctx, &completionPayload{
		Model:     defaultModel,
		Prompt:    r.Prompt,
		MaxTokens: r.MaxTokens,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return &Completion{
		Text: resp.Choices[0].Text,
	}, nil
}
