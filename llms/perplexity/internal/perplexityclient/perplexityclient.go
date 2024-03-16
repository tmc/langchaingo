package perplexityclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	defaultBaseURL              = "https://api.perplexity.com"
	defaultFunctionCallBehavior = "auto"
)

// ErrEmptyResponse is returned when the perplexity API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the perplexity API.
type Client struct {
	token      string
	Model      string
	baseURL    string
	httpClient Doer
}

// Option is an option for the perplexity client.
type Option func(*Client) error

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// New returns a new perplexity client.
func New(token string, model string, baseURL string, httpClient Doer,
	opts ...Option,
) (*Client, error) {
	c := &Client{
		token:      token,
		Model:      model,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: httpClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	resp, err := c.createCompletion(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return &Completion{
		Text: resp.Choices[0].Message.Content,
	}, nil
}

// EmbeddingRequest is a request to create an embedding.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// CreateChat creates chat request.
func (c *Client) CreateChat(ctx context.Context, r *ChatRequest) (*ChatResponse, error) {
	if r.Model == "" {
		if c.Model == "" {
			r.Model = defaultChatModel
		} else {
			r.Model = c.Model
		}
	}
	if r.FunctionCallBehavior == "" && len(r.Functions) > 0 {
		r.FunctionCallBehavior = defaultFunctionCallBehavior
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

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

}

func (c *Client) buildURL(suffix string, model string) string {
	// open ai implement:
	return fmt.Sprintf("%s%s", c.baseURL, suffix)
}
