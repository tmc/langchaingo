package openaiclient

import (
	"context"
	"errors"
)

var ErrEmptyResponse = errors.New("empty response")

type Client struct {
	token string
}

func New(token string) (*Client, error) {
	c := &Client{token: token}
	return c, nil
}

type CompletionRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type Completion struct {
	Text string `json:"text"`
}

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
