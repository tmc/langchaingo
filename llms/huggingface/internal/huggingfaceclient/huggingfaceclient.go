package huggingfaceclient

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrEmptyResponse = errors.New("empty response")
)

type Client struct {
	token string
}

func New(token string) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	return &Client{
		token: token,
	}, nil
}

type InferenceRequest struct {
	RepoID string        `json:"repositoryId"`
	Prompt string        `json:"prompt"`
	Task   InferenceTask `json:"task"`
}

type InferenceResponse struct {
	Text string `json:"generated_text"`
}

func (c *Client) RunInference(ctx context.Context, request *InferenceRequest) (*InferenceResponse, error) {
	resp, err := c.runInference(ctx, &inferencePayload{
		Model:  "gpt2",
		Inputs: request.Prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}
	if len(resp) == 0 {
		return nil, ErrEmptyResponse
	}
	text := resp[0].Text
	// Strip the prompt from the response:
	text = text[len(request.Prompt)+1:]
	return &InferenceResponse{
		Text: text,
	}, nil
}
