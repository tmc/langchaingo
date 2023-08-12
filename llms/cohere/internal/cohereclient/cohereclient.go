package cohereclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cohere-ai/tokenizer"
)

var (
	ErrEmptyResponse = errors.New("empty response")
	ErrModelNotFound = errors.New("model not found")
)

type Client struct {
	token      string
	baseURL    string
	model      string
	httpClient Doer
	encoder    *tokenizer.Encoder
}

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

func New(token string, baseURL string, model string, opts ...Option) (*Client, error) {
	encoder, err := tokenizer.NewFromPrebuilt("coheretext-50k")
	if err != nil {
		return nil, fmt.Errorf("create tokenizer: %w", err)
	}

	c := &Client{
		token:      token,
		baseURL:    baseURL,
		model:      model,
		httpClient: http.DefaultClient,
		encoder:    encoder,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

type GenerationRequest struct {
	Prompt string `json:"prompt"`
}

type Generation struct {
	Text string `json:"text"`
}

type generateRequestPayload struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

type generateResponsePayload struct {
	ID          string `json:"id,omitempty"`
	Message     string `json:"message,omitempty"`
	Generations []struct {
		ID   string `json:"id,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"generations,omitempty"`
}

func (c *Client) CreateGeneration(ctx context.Context, r *GenerationRequest) (*Generation, error) {
	if c.baseURL == "" {
		c.baseURL = "https://api.cohere.ai"
	}

	payload := generateRequestPayload{
		Prompt: r.Prompt,
		Model:  c.model,
	}

	payloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/generate", c.baseURL),
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	var response generateResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(response.Generations) == 0 {
		if strings.HasPrefix(response.Message, "model not found") {
			return nil, ErrModelNotFound
		}
		return nil, ErrEmptyResponse
	}

	var generation Generation
	generation.Text = response.Generations[0].Text

	return &generation, nil
}

func (c *Client) GetNumTokens(text string) int {
	encoded, _ := c.encoder.Encode(text)
	return len(encoded)
}
