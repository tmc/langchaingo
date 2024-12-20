package qwenclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	ErrNotSetAuth     = errors.New("api key is not set")
	ErrCompletionCode = errors.New("completion API returned unexpected status code")
)

const (
	DefaultTextModelPath = "qwen-turbo"
	DefaultVLModelPath   = "qwen-vl-plus"
)

type Client struct {
	apiKey     string
	httpClient Doer
	Model      string
}

type ModelPath string

type Option func(*Client) error

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Parameters struct {
	Temperature       float64 `json:"temperature"`
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	MaxTokens         int     `json:"max_tokens,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
	ResultFormat      string  `json:"result_format"`
}

type CompletionRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []Message `json:"messages"`
	} `json:"input"`
	Parameters    Parameters                                    `json:"parameters"`
	Stream        bool                                          `json:"stream,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type Completion struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Output    struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
		Choices      []struct {
			FinishReason string  `json:"finish_reason"`
			Message      Message `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		TotalTokens  int `json:"total_tokens"`
		OutputTokens int `json:"output_tokens"`
		InputTokens  int `json:"input_tokens"`
	} `json:"usage"`
}

type VLContent struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

type VLMessage struct {
	Role    string      `json:"role"`
	Content []VLContent `json:"content"`
}

type VLRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []VLMessage `json:"messages"`
	} `json:"input"`
	Parameters    Parameters                                    `json:"parameters,omitempty"`
	Stream        bool                                          `json:"stream,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type VLResponse struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Output    struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
		Choices      []struct {
			FinishReason string    `json:"finish_reason"`
			Message      VLMessage `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
		InputTokens  int `json:"input_tokens"`
		ImageTokens  int `json:"image_tokens"`
	} `json:"usage"`
}

func WithHTTPClient(client Doer) Option {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

func WithAK(apiKey string) Option {
	return func(c *Client) error {
		c.apiKey = apiKey
		return nil
	}
}

func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.apiKey == "" {
		return nil, ErrNotSetAuth
	}

	return c, nil
}

func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	body, e := json.Marshal(r)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	c.setHeader(req, r.Stream)

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Read and print the full response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		// Reset the response body to be read again later
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		fmt.Printf("full response body: %s\n", string(bodyBytes))
		return nil, fmt.Errorf("%w: %d", ErrCompletionCode, resp.StatusCode)
	}

	if r.Stream {
		return parseStreamingCompletionResponse(ctx, resp, r)
	}

	var response Completion
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

func (c *Client) CreateVLChat(ctx context.Context, r *VLRequest) (*VLResponse, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	body, e := json.Marshal(r)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	c.setHeader(req, r.Stream)

	resp, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrCompletionCode, resp.StatusCode)
	}

	var response VLResponse
	return &response, json.NewDecoder(resp.Body).Decode(&response)
}

func (c *Client) setHeader(req *http.Request, stream bool) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("X-DashScope-SSE", "enable")
	}
}

func parseStreamingCompletionResponse(ctx context.Context, resp *http.Response, r *CompletionRequest) (*Completion, error) {
	scanner := bufio.NewScanner(resp.Body)
	responseChan := make(chan *Completion)
	go func() {
		defer close(responseChan)
		dataPrefix := "data:"
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, dataPrefix) && !strings.HasPrefix(line, "{") {
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
		var text = ""
		if len(streamResponse.Output.Choices) == 1 {
			text = streamResponse.Output.Choices[0].Message.Content
		}
		response.Output.Text += text
		if r.StreamingFunc != nil {
			err := r.StreamingFunc(ctx, []byte(text))
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
		lastResponse = streamResponse
	}
	// update
	lastResponse.Output.Text = response.Output.Text
	lastResponse.Usage.OutputTokens = lastResponse.Usage.TotalTokens - lastResponse.Usage.InputTokens
	return lastResponse, nil
}
