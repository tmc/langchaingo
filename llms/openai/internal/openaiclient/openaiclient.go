package openaiclient

import (
	"context"
	"errors"
	"net/http"
)

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the OpenAI API.
type Client struct {
	token        string
	Model        string
	baseURL      string
	organization string
	httpClient   Doer
}

// Option is an option for the OpenAI client.
type Option func(*Client) error

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// WithHTTPClient allows setting a custom HTTP client.
func WithHTTPClient(client Doer) Option {
	return func(c *Client) error {
		c.httpClient = client

		return nil
	}
}

// New returns a new OpenAI client.
func New(token string, model string, baseURL string, organization string, opts ...Option) (*Client, error) {
	c := &Client{
		token:        token,
		Model:        model,
		baseURL:      baseURL,
		organization: organization,
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Temperature      float64  `json:"temperature,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	N                int      `json:"n,omitempty"`
	StopWords        []string `json:"stop,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	resp, err := c.createCompletion(ctx, &completionPayload{
		Model:            r.Model,
		Prompt:           r.Prompt,
		Temperature:      r.Temperature,
		MaxTokens:        r.MaxTokens,
		StopWords:        r.StopWords,
		N:                r.N,
		FrequencyPenalty: r.FrequencyPenalty,
		PresencePenalty:  r.PresencePenalty,
		TopP:             r.TopP,
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

// EmbeddingRequest is a request to create an embedding.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// CreateEmbedding creates embeddings.
func (c *Client) CreateEmbedding(ctx context.Context, r *EmbeddingRequest) ([][]float64, error) {
	r.Model = c.Model
	if r.Model == "" {
		r.Model = defaultEmbeddingModel
	}

	resp, err := c.createEmbedding(ctx, &embeddingPayload{
		Model: r.Model,
		Input: r.Input,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, ErrEmptyResponse
	}

	embeddings := make([][]float64, 0)
	for i := 0; i < len(resp.Data); i++ {
		embeddings = append(embeddings, resp.Data[i].Embedding)
	}

	return embeddings, nil
}

// CreateChat creates chat request.
func (c *Client) CreateChat(ctx context.Context, r *ChatRequest) (*ChatResponse, error) {
	r.Model = c.Model
	if r.Model == "" {
		r.Model = defaultChatModel
	}
	resp, err := c.createChat(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp, nil
}
