package databricks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/tmc/langchaingo/llms"
)

// Option is a function that applies a configuration to the LLM.
type Option func(*LLM)

// LLM is a databricks LLM implementation.
type LLM struct {
	url        string // The constructed or provided URL
	token      string // The token for authentication
	httpClient *http.Client
	model      Model
}

var _ llms.Model = (*LLM)(nil)

// New creates a new llamafile LLM implementation.
func New(model Model, opts ...Option) (*LLM, error) {
	llm := &LLM{
		model: model,
	}

	// Apply all options to customize the LLM.
	for _, opt := range opts {
		opt(llm)
	}

	// Validate URL
	if llm.url == "" {
		return nil, fmt.Errorf("URL must be provided or constructed using options")
	}

	if llm.httpClient == nil {
		if llm.token == "" {
			return nil, fmt.Errorf("token must be provided")
		}
		llm.httpClient = NewHTTPClient(llm.token)
	}

	return llm, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop, funlen
	payload, err := o.model.FormatPayload(ctx, messages, options...)
	if err != nil {
		return nil, err
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	fmt.Printf("payload: %v\n", string(payload))

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := o.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	// Create a buffer to save a copy of the body
	var buffer bytes.Buffer
	teeReader := io.TeeReader(resp.Body, &buffer)

	if err := o.stream(ctx, &buffer, opts); err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(teeReader)
	if err != nil {
		return nil, err
	}
	fmt.Printf("bodyBytes: %v\n", string(bodyBytes))
	return o.model.FormatResponse(ctx, bodyBytes)
}

func (o *LLM) stream(ctx context.Context, resp io.Reader, opts llms.CallOptions) error {
	if opts.StreamingFunc == nil {
		return nil
	}

	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		contentResponse, err := o.model.FormatResponse(ctx, scanner.Bytes())
		if err != nil {
			return err
		}

		fmt.Printf("contentResponse: %v\n", *contentResponse)

		if len(contentResponse.Choices) == 0 {
			continue
		}

		if err := opts.StreamingFunc(ctx, []byte(contentResponse.Choices[0].Content)); err != nil {
			return err
		}
	}

	return scanner.Err()
}
