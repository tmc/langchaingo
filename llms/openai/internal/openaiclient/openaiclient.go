package openaiclient

import (
	"context"
	"errors"
)

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the OpenAI API.
type Client struct {
	token string
	model string
}

// New returns a new OpenAI client.
func New(token string, model string) (*Client, error) {
	c := &Client{token: token, model: model}
	return c, nil
}

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Model     string   `json:"model"`
	Prompt    string   `json:"prompt"`
	MaxTokens int      `json:"max_tokens"`
	StopWords []string `json:"stop,omitempty"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	r.Model = c.model
	if r.Model == "" {
		r.Model = defaultCompletionModel
	}

	resp, err := c.createCompletion(ctx, &completionPayload{
		Model:     r.Model,
		Prompt:    r.Prompt,
		MaxTokens: r.MaxTokens,
		StopWords: r.StopWords,
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

// CreateCompletion creates embeddings.
func (c *Client) CreateEmbedding(ctx context.Context, r *EmbeddingRequest) ([][]float64, error) {
	r.Model = c.model
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
	r.Model = c.model
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
