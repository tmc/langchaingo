package huggingfaceclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrEmptyResponse = errors.New("empty response")
)

type Client struct {
	Token string
	Model string
	url   string
}

func New(token, model, url string) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	return &Client{
		Token: token,
		Model: model,
		url:   url,
	}, nil
}

type InferenceRequest struct {
	Model             string        `json:"repositoryId"`
	Prompt            string        `json:"prompt"`
	Task              InferenceTask `json:"task"`
	Temperature       float64       `json:"temperature"`
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
		Model:  request.Model,
		Inputs: request.Prompt,
		Parameters: parameters{
			Temperature:       request.Temperature,
			TopP:              request.TopP,
			TopK:              request.TopK,
			MinLength:         request.MinLength,
			MaxLength:         request.MaxLength,
			RepetitionPenalty: request.RepetitionPenalty,
			Seed:              request.Seed,
		},
	}
	resp, err := c.runInference(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}
	if len(resp) == 0 {
		return nil, ErrEmptyResponse
	}
	text := resp[0].Text
	// Clean response based on model type - remove prompt repetition for GPT-2 style models
	text = cleanModelResponse(text, request.Prompt, request.Model)
	return &InferenceResponse{
		Text: text,
	}, nil
}

// cleanModelResponse cleans the response text based on the model type.
func cleanModelResponse(text, prompt, model string) string {
	switch {
	case strings.Contains(strings.ToLower(model), "gpt2"):
		// For GPT-2 models, remove the prompt from the beginning if present
		if strings.HasPrefix(text, prompt) {
			return strings.TrimSpace(text[len(prompt):])
		}
	case strings.Contains(strings.ToLower(model), "t5"):
		// T5 models typically don't repeat the prompt, but may have extra tokens
		return strings.TrimSpace(text)
	default:
		// For other models, just trim whitespace
		return strings.TrimSpace(text)
	}
	return text
}

// EmbeddingRequest is a request to create an embedding.
type EmbeddingRequest struct {
	Options map[string]any `json:"options"`
	Inputs  []string       `json:"inputs"`
}

// CreateEmbedding creates embeddings.
func (c *Client) CreateEmbedding(
	ctx context.Context,
	model string,
	task string,
	r *EmbeddingRequest,
) ([][]float32, error) {
	resp, err := c.createEmbedding(ctx, model, task, &embeddingPayload{
		Inputs:  r.Inputs,
		Options: r.Options,
	})
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, ErrEmptyResponse
	}

	return resp, nil
}
