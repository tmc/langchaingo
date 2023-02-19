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
		Model:  request.RepoID,
		Inputs: request.Prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}
	if len(resp) == 0 {
		return nil, ErrEmptyResponse
	}
	text := resp[0].Text
	// TODO: Add response cleaning based on Model.
	// e.g., for gpt2, text = text[len(request.Prompt)+1:]
	return &InferenceResponse{
		Text: text,
	}, nil
}
