package gigachatclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrNotSetAuth      = errors.New("both accessToken and apiKey secretKey are not set")
	ErrAccessTokenCode = errors.New("get access_token API returned unexpected status code")
	ErrEmbeddingCode   = errors.New("embedding API returned unexpected status code")
)

// Client is a client for the GigaChat API.
type Client struct {
	baseURL      string
	clientId     string
	clientSecret string
	scope        string
	accessToken  string
	httpClient   Doer
	Model        string
}

const (
	DefaultBaseURL = "https://gigachat.devices.sberbank.ru/api/v1"
	DefaultModel   = "GigaChat-Pro"
	tryPeriod      = 3 // minutes
)

// Option is an option for the Gigachat client.
type Option func(*Client) error

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
		Usage     struct {
			PromptTokens int `json:"prompt_tokens"`
		} `json:"usage"`
	} `json:"data"`
	Model string `json:"model,omitempty"`
}

type ErrorMessage struct {
	ErrorCode int    `json:"status"`
	Message   string `json:"message"`
}

type authResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// WithHTTPClient allows setting a custom HTTP client.
func WithHTTPClient(client Doer) Option {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

func WithBaseURL(url string) Option {
	return func(c *Client) error {
		c.baseURL = url
		return nil
	}
}

// WithClientIdAndSecret allows setting clientId, clientSecret.
func WithClientIdAndSecret(clientId, clientSecret string) Option {
	return func(c *Client) error {
		c.clientId = clientId
		c.clientSecret = clientSecret
		return nil
	}
}

func WithScope(scope string) Option {
	return func(c *Client) error {
		c.scope = scope
		return nil
	}
}

func WithModel(model string) Option {
	return func(c *Client) error {
		c.Model = model
		return nil
	}
}

// WithAccessToken is usually used for dev, Prod env recommend use WithClientIdAndSecret.
func WithAccessToken(accessToken string) Option {
	return func(c *Client) error {
		c.accessToken = accessToken
		return nil
	}
}

// New returns a new Gigachat client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.accessToken == "" && (c.clientId == "" || c.clientSecret == "") {
		return nil, ErrNotSetAuth
	}

	if c.clientId != "" && c.clientSecret != "" && c.accessToken == "" {
		err := autoRefresh(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func autoRefresh(c *Client) error {
	authResp, err := c.getAccessToken(context.Background())
	if err != nil {
		return err
	}
	c.accessToken = authResp.AccessToken
	go func() { // 30 min expiration: https://developers.sber.ru/docs/ru/gigachat/quickstart/ind-using-api
		for {
			authResp, err := c.getAccessToken(context.Background())
			if err != nil {
				time.Sleep(tryPeriod * time.Minute) // try
				continue
			}
			c.accessToken = authResp.AccessToken
			time.Sleep(25 * time.Minute)
		}
	}()
	return nil
}

// CreateEmbedding use Gigachat Embedding-V1.
func (c *Client) CreateEmbedding(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	payload := make(map[string]any)
	payload["input"] = texts
	payload["model"] = "Embeddings"

	body, e := json.Marshal(payload)
	if e != nil {
		return nil, e
	}

	resp, e := c.do(ctx, "/embeddings", body)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	var response EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

// accessToken 30 min expiration:https://developers.sber.ru/docs/ru/gigachat/quickstart/ind-using-api
func (c *Client) getAccessToken(ctx context.Context) (*authResponse, error) {
	url := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"

	data := strings.NewReader("scope=" + c.scope)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, data)
	if e != nil {
		return nil, e
	}

	// Create the base64 encoded auth string
	auth := base64.StdEncoding.EncodeToString([]byte(c.clientId + ":" + c.clientSecret))

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("RqUID", uuid.New().String())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	var response authResponse
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
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

func (c *Client) decodeError(resp *http.Response) error {
	msg := fmt.Sprintf("API returned unexpected status code: %d", resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(msg) // nolint:goerr113
	}
	var errResp ErrorMessage
	if err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&errResp); err != nil {
		return fmt.Errorf("%s, %s", msg, string(bodyBytes))
	}
	return fmt.Errorf("%s: %s", msg, errResp.Message) // nolint:goerr113
}
