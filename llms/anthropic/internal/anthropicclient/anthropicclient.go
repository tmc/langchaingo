package anthropicclient

import (
	"context"
	"errors"
	"net/http"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1"
)

// ErrEmptyResponse is returned when the Anthropic API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the Anthropic API.
type Client struct {
	token   string
	Model   string
	baseURL string

	httpClient Doer
}

// Option is an option for the Anthropic client.
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

// New returns a new Anthropic client.
func New(token string, model string, opts ...Option) (*Client, error) {
	c := &Client{
		Model:      model,
		token:      token,
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
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
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens_to_sample,omitempty"`
	StopWords   []string `json:"stop_sequences,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	Stream      bool     `json:"stream,omitempty"`

	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	resp, err := c.createCompletion(ctx, &completionPayload{
		Model:         r.Model,
		Prompt:        r.Prompt,
		Temperature:   r.Temperature,
		MaxTokens:     r.MaxTokens,
		StopWords:     r.StopWords,
		TopP:          r.TopP,
		Stream:        r.Stream,
		StreamingFunc: r.StreamingFunc,
	})
	if err != nil {
		return nil, err
	}
	return &Completion{
		Text: resp.Completion,
	}, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.token)
	// TODO: expose version as a option/parameter
	req.Header.Set("anthropic-version", "2023-06-01")
}
