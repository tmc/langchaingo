package localclient

import (
	"context"
	"errors"
)

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for a local LLM.
type Client struct {
	BinPath      string
	Args         []string
	GlobalAsArgs bool
}

// New returns a new local client.
func New(binPath string, globalAsArgs bool, args ...string) (*Client, error) {
	c := &Client{BinPath: binPath, GlobalAsArgs: globalAsArgs, Args: args}
	return c, nil
}

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Prompt string `json:"prompt"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	resp, err := c.createCompletion(ctx, &completionPayload{
		Prompt: r.Prompt,
	})
	if err != nil {
		return nil, err
	}
	return &Completion{
		Text: resp.Response,
	}, nil
}
