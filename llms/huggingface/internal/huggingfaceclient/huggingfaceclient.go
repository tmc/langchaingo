package huggingfaceclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrEmptyResponse = errors.New("empty response")
)

type Client struct {
	Token      string
	Model      string
	url        string
	httpClient *http.Client
}

func New(token, model, url string) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	return &Client{
		Token:      token,
		Model:      model,
		url:        url,
		httpClient: http.DefaultClient,
	}, nil
}

// SetHTTPClient sets the HTTP client for the Hugging Face client.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
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
	
	// Clean response based on model type
	text = c.cleanResponse(request.Prompt, text, request.Model)
	
	return &InferenceResponse{
		Text: text,
	}, nil
}

// cleanResponse cleans the model response based on the model type and prompt.
func (c *Client) cleanResponse(prompt, text, model string) string {
	// Handle GPT-2 style models that include the prompt in the response
	if strings.Contains(strings.ToLower(model), "gpt2") || 
	   strings.Contains(strings.ToLower(model), "gpt-2") {
		if strings.HasPrefix(text, prompt) {
			return strings.TrimSpace(text[len(prompt):])
		}
	}
	
	// Handle DialoGPT models that might repeat the prompt
	if strings.Contains(strings.ToLower(model), "dialogpt") {
		if strings.HasPrefix(text, prompt) {
			return strings.TrimSpace(text[len(prompt):])
		}
	}
	
	// For other models, return as-is but trimmed
	return strings.TrimSpace(text)
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
