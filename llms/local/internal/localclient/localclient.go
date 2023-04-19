package localclient

import (
	"context"
	"errors"
)

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for a local LLM.
type Client struct {
	binPath string
	args    []string
}

// New returns a new local client.
func New(binPath string, args string) (*Client, error) {
	var parts []string
	var currentPart string
	var inDoubleQuote, inSingleQuote bool

	for _, char := range args {
		switch {
		case char == ' ' && !inDoubleQuote && !inSingleQuote:
			if currentPart != "" {
				parts = append(parts, currentPart)
				currentPart = ""
			}
		case char == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
		case char == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
		default:
			currentPart += string(char)
		}
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	c := &Client{binPath: binPath, args: parts}
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
