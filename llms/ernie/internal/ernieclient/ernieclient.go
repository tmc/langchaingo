package ernieclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrNotSetAuth      = errors.New("both accessToken and apiKey secretKey are not set")
	ErrCompletionCode  = errors.New("completion API returned unexpected status code")
	ErrAccessTokenCode = errors.New("get access_token API returned unexpected status code")
	ErrEmbeddingCode   = errors.New("embedding API returned unexpected status code")
)

// Client is a client for the ERNIE API.
type Client struct {
	apiKey      string
	secretKey   string
	accessToken string
	httpClient  Doer
}

// ModelPath ERNIE API URL path suffix distinguish models.
type ModelPath string

// DefaultCompletionModelPath default model.
const (
	DefaultCompletionModelPath = "completions"
	tryPeriod                  = 3 // minutes
)

// Option is an option for the ERNIE client.
type Option func(*Client) error

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Messages      []Message                                     `json:"messages"`
	Temperature   float64                                       `json:"temperature,omitempty"`
	TopP          float64                                       `json:"top_p,omitempty"`
	PenaltyScore  float64                                       `json:"penalty_score,omitempty"`
	Stream        bool                                          `json:"stream,omitempty"`
	UserID        string                                        `json:"user_id,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

// Completion is a completion.
type Completion struct {
	ID               string `json:"id"`
	Object           string `json:"object"`
	Created          int    `json:"created"`
	SentenceID       int    `json:"sentence_id"`
	IsEnd            bool   `json:"is_end"`
	IsTruncated      bool   `json:"is_truncated"`
	Result           string `json:"result"`
	NeedClearHistory bool   `json:"need_clear_history"`
	Usage            struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	// for error
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type EmbeddingResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Data    []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	// for error
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type authResponse struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	Scope         string `json:"scope"`
	SessionSecret string `json:"session_secret"`
}

// WithHTTPClient allows setting a custom HTTP client.
func WithHTTPClient(client Doer) Option {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// WithAKSK allows setting apiKey, secretKey.
func WithAKSK(apiKey, secretKey string) Option {
	return func(c *Client) error {
		c.apiKey = apiKey
		c.secretKey = secretKey
		return nil
	}
}

// Usually used for dev, Prod env recommend use WithAKSK.
func WithAccessToken(accessToken string) Option {
	return func(c *Client) error {
		c.accessToken = accessToken
		return nil
	}
}

// New returns a new ERNIE client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.accessToken == "" && (c.apiKey == "" || c.secretKey == "") {
		return nil, ErrNotSetAuth
	}

	if c.apiKey != "" && c.secretKey != "" {
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
	go func() { // 30 day expiration, auto refresh access token per 10 days
		for {
			authResp, err := c.getAccessToken(context.Background())
			if err != nil {
				time.Sleep(tryPeriod * time.Minute) // try
				continue
			}
			c.accessToken = authResp.AccessToken
			time.Sleep(10 * 24 * time.Hour)
		}
	}()
	return nil
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, modelPath ModelPath, r *CompletionRequest) (*Completion, error) {
	if modelPath == "" {
		modelPath = DefaultCompletionModelPath
	}

	url := "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/" + string(modelPath) +
		"?access_token=" + c.accessToken
	body, e := json.Marshal(r)
	if e != nil {
		return nil, e
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrCompletionCode, resp.StatusCode)
	}

	if r.Stream {
		return parseStreamingCompletionResponse(ctx, resp, r)
	}

	var response Completion
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

// CreateEmbedding use ernie Embedding-V1.
func (c *Client) CreateEmbedding(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	url := "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1?access_token=" +
		c.accessToken

	payload := make(map[string]any)
	payload["input"] = texts

	body, e := json.Marshal(payload)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrEmbeddingCode, resp.StatusCode)
	}

	var response EmbeddingResponse
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

// accessToken 30 day expiration https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Ilkkrb0i5
func (c *Client) getAccessToken(ctx context.Context) (*authResponse, error) {
	url := fmt.Sprintf(
		"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%v&client_secret=%v",
		c.apiKey, c.secretKey)

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("")))
	if e != nil {
		return nil, e
	}

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrAccessTokenCode, resp.StatusCode)
	}

	var response authResponse
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

func parseStreamingCompletionResponse(ctx context.Context, resp *http.Response, req *CompletionRequest) (*Completion, error) { // nolint:lll
	scanner := bufio.NewScanner(resp.Body)
	responseChan := make(chan *Completion)
	go func() {
		defer close(responseChan)
		dataPrefix := "data: "
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, dataPrefix) {
				continue
			}
			data := strings.TrimPrefix(line, dataPrefix)
			streamPayload := &Completion{}
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				log.Fatalf("failed to decode stream payload: %v", err)
			}
			responseChan <- streamPayload
		}
		if err := scanner.Err(); err != nil {
			log.Println("issue scanning response:", err)
		}
	}()
	// Parse response
	response := Completion{}

	var lastResponse *Completion
	for streamResponse := range responseChan {
		response.Result += streamResponse.Result
		if req.StreamingFunc != nil {
			err := req.StreamingFunc(ctx, []byte(streamResponse.Result))
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
		lastResponse = streamResponse
	}
	// update
	lastResponse.Result = response.Result
	lastResponse.Usage.CompletionTokens = lastResponse.Usage.TotalTokens - lastResponse.Usage.PromptTokens
	return lastResponse, nil
}
