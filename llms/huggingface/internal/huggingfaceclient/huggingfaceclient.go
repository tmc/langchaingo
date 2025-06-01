package huggingfaceclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tmc/langchaingo/httputil"
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
	provider   string // Inference provider for router-based requests
}

func New(token, model, url string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	client := &Client{
		Token:      token,
		Model:      model,
		url:        url,
		httpClient: httputil.DefaultClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Option configures a HuggingFace client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client for the HuggingFace client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithProvider sets the inference provider for router-based requests.
func WithProvider(provider string) Option {
	return func(c *Client) {
		c.provider = provider
	}
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
	// TODO: Add response cleaning based on Model.
	// e.g., for gpt2, text = text[len(request.Prompt)+1:]
	return &InferenceResponse{
		Text: text,
	}, nil
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
