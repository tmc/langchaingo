package anthropicclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const (
	DefaultBaseURL = "https://api.anthropic.com/v1"

	defaultModel = "claude-sonnet-4-5"
)

// ErrEmptyResponse is returned when the Anthropic API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the Anthropic API.
type Client struct {
	token   string
	Model   string
	baseURL string

	httpClient Doer

	// Changed from single string to slice for multiple beta headers
	anthropicBetaHeaders []string

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

// WithAnthropicBetaHeader adds an anthropic-beta header.
// Can be called multiple times to add multiple beta headers.
func WithAnthropicBetaHeader(val string) Option {
	return func(opts *Client) error {
		if opts.anthropicBetaHeaders == nil {
			opts.anthropicBetaHeaders = []string{}
		}
		opts.anthropicBetaHeaders = append(opts.anthropicBetaHeaders, val)
		return nil
	}
}

// New returns a new Anthropic client.
func New(token string, model string, baseURL string, opts ...Option) (*Client, error) {
	c := &Client{
		Model:      model,
		token:      token,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: httputil.DefaultClient,
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
	StreamingFunc streaming.Callback `json:"-"`
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
	Model       string           `json:"model"`
	Messages    []ChatMessage    `json:"messages"`
	System      any              `json:"system,omitempty"` // Can be string or []Content for caching
	Temperature *float64         `json:"temperature,omitempty"`
	MaxTokens   *int             `json:"max_tokens,omitempty"`
	TopP        *float64         `json:"top_p,omitempty"`
	Tools       []Tool           `json:"tools,omitempty"`
	ToolChoice  any              `json:"tool_choice,omitempty"`
	StopWords   []string         `json:"stop_sequences,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Thinking    *ThinkingPayload `json:"thinking,omitempty"`

	// BetaHeaders are additional beta feature headers to include
	BetaHeaders   []string           `json:"-"`
	StreamingFunc streaming.Callback `json:"-"`
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
		ToolChoice:    r.ToolChoice,
		Stream:        r.Stream,
		StreamingFunc: r.StreamingFunc,
		Thinking:      r.Thinking,
	}, r.BetaHeaders)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) setHeaders(req *http.Request, betaHeaders []string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.token) //nolint:canonicalheader

	// This is necessary as per https://docs.anthropic.com/en/api/versioning
	// If this changes frequently enough we should expose it as an option..
	req.Header.Set("anthropic-version", "2023-06-01") // nolint:canonicalheader

	// Set beta headers from request, falling back to client default
	// Multiple beta features are combined in a single header with comma separation
	var validHeaders []string
	if len(betaHeaders) > 0 {
		for _, header := range betaHeaders {
			if header != "" && !slices.Contains(validHeaders, header) {
				validHeaders = append(validHeaders, header)
			}
		}
	} else if len(c.anthropicBetaHeaders) > 0 {
		for _, header := range c.anthropicBetaHeaders {
			if header != "" && !slices.Contains(validHeaders, header) {
				validHeaders = append(validHeaders, header)
			}
		}
	}

	if len(validHeaders) > 0 {
		req.Header.Set("anthropic-beta", strings.Join(validHeaders, ",")) // nolint:canonicalheader
	}
}

func (c *Client) do(ctx context.Context, path string, payloadBytes []byte) (*http.Response, error) {
	return c.doWithHeaders(ctx, path, payloadBytes, nil)
}

func (c *Client) doWithHeaders(ctx context.Context, path string, payloadBytes []byte, betaHeaders []string) (*http.Response, error) {
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}

	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req, betaHeaders)

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
		return errors.New(msg)
	}
	return fmt.Errorf("%s: %s", msg, errResp.Error.Message)
}
