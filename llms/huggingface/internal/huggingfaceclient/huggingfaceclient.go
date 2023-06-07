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
	Token string
	Model string
}

func New(token string, model string) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	return &Client{
		Token: token,
		Model: model,
	}, nil
}

type InferenceRequest struct {
	Model             string        `json:"repositoryId"`
	Prompt            string        `json:"prompt"`
	Task              InferenceTask `json:"task"`
	Temperature       float64       `json:"temperature,omitempty"`
	TopP              float64       `json:"top_p,omitempty"`
	TopK              int           `json:"top_k,omitempty"`
	MinLength         int           `json:"min_length,omitempty"`
	MaxLength         int           `json:"max_length,omitempty"`
	RepetitionPenalty float64       `json:"repetition_penalty,omitempty"`
	Seed              int           `json:"seed,omitempty"`
}

type InferenceResponse struct {
	Text string `json:"generated_text"`
}

func (c *Client) RunInference(ctx context.Context, request *InferenceRequest) (*InferenceResponse, error) {
	payload := &inferencePayload{
		Model:             request.Model,
		Inputs:            request.Prompt,
		Temperature:       request.Temperature,
		TopP:              request.TopP,
		TopK:              request.TopK,
		MinLength:         request.MinLength,
		MaxLength:         request.MaxLength,
		RepetitionPenalty: request.RepetitionPenalty,
		Seed:              request.Seed,
	}
	resp, err := c.runInference(ctx, payload)
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
