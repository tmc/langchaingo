package anthropicclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	DefaultBaseURL = "https://api.anthropic.com/v1"

	defaultModel = "claude-3-5-sonnet-20240620"
)

// ErrEmptyResponse is returned when the Anthropic API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the Anthropic API.
type Client struct {
	token   string
	Model   string
	baseURL string

	httpClient Doer

	anthropicBetaHeader string

	// UseLegacyTextCompletionsAPI is a flag to use the legacy text completions API.
	UseLegacyTextCompletionsAPI bool
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

// WithLegacyTextCompletionsAPI enables the use of the legacy text completions API.
func WithLegacyTextCompletionsAPI(val bool) Option {
	return func(opts *Client) error {
		opts.UseLegacyTextCompletionsAPI = val
		return nil
	}
}

// WithAnthropicBetaHeader sets the anthropic-beta header.
func WithAnthropicBetaHeader(val string) Option {
	return func(opts *Client) error {
		opts.anthropicBetaHeader = val
		return nil
	}
}

// New returns a new Anthropic client.
func New(token string, model string, baseURL string, opts ...Option) (*Client, error) {
	c := &Client{
		Model:   model,
		token:   token,
		baseURL: strings.TrimSuffix(baseURL, "/"),
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

type MessageRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	System      string        `json:"system,omitempty"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	StopWords   []string      `json:"stop_sequences,omitempty"`
	Stream      bool          `json:"stream,omitempty"`

	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

// CreateMessage creates message for the messages api.
func (c *Client) CreateMessage(ctx context.Context, r *MessageRequest) (*MessageResponsePayload, error) {
	resp, err := c.createMessage(ctx, &messagePayload{
		Model:         r.Model,
		Messages:      r.Messages,
		System:        r.System,
		Temperature:   r.Temperature,
		MaxTokens:     r.MaxTokens,
		StopWords:     r.StopWords,
		TopP:          r.TopP,
		Tools:         r.Tools,
		Stream:        r.Stream,
		StreamingFunc: r.StreamingFunc,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.token) //nolint:canonicalheader

	// This is necessary as per https://docs.anthropic.com/en/api/versioning
	// If this changes frequently enough we should expose it as an option..
	req.Header.Set("anthropic-version", "2023-06-01") // nolint:canonicalheader
	if c.anthropicBetaHeader != "" {
		req.Header.Set("anthropic-beta", c.anthropicBetaHeader) // nolint:canonicalheader
	}
}

func (c *Client) do(ctx context.Context, path string, payloadBytes []byte) (*http.Response, error) {
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}

	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp, nil
}

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (c *Client) decodeError(resp *http.Response) error {
	msg := fmt.Sprintf("API returned unexpected status code: %d", resp.StatusCode)

	var errResp errorMessage
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return errors.New(msg) // nolint:goerr113
	}
	return fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
}
