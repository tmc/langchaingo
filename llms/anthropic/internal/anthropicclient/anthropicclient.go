package anthropicclient

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

const (
	DefaultBaseURL = "https://api.anthropic.com/v1"
)

// ErrEmptyResponse is returned when the Anthropic API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the Anthropic API.
type Client struct {
	token   string
	Model   string
	baseURL string

	httpClient Doer
	// When true the client will use the legacy text completion API
	// when false it will use the new messages completion api
	UseLegacyTextCompletionApi bool
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
func New(token string, model string, baseURL string, httpClient Doer, useLegacyTextCompletionApi bool, opts ...Option) (*Client, error) {
	c := &Client{
		Model:                      model,
		token:                      token,
		baseURL:                    strings.TrimSuffix(baseURL, "/"),
		httpClient:                 httpClient,
		UseLegacyTextCompletionApi: useLegacyTextCompletionApi,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

type MessagesCompletionRequest struct {
	Model        string         `json:"model"`
	Messages     []*ChatMessage `json:"messages"`
	SystemPrompt string         `json:"system,omitempty"`
	Temperature  float64        `json:"temperature"`
	MaxTokens    int            `json:"max_tokens,omitempty"`
	StopWords    []string       `json:"stop_sequences,omitempty"`
	TopP         float64        `json:"top_p,omitempty"`
}

// CompletionRequest is a request to create a completion.
type LegacyTextCompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature"`
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

func (c *Client) CreateMessagesCompletion(ctx context.Context, r *MessagesCompletionRequest) (*MessagesCompletionResponsePayload, error) {
	resp, err := c.createMessagesCompletion(ctx, &messagesCompletionPayload{
		Model:        r.Model,
		Messages:     r.Messages,
		SystemPrompt: r.SystemPrompt,
		Temperature:  r.Temperature,
		MaxTokens:    r.MaxTokens,
		StopWords:    r.StopWords,
		TopP:         r.TopP,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateCompletion creates a completion.
func (c *Client) CreateLegacyTextCompletion(ctx context.Context, r *LegacyTextCompletionRequest) (*Completion, error) {
	resp, err := c.createLegacyTextCompletion(ctx, &legacyTextCompletionPayload{
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
