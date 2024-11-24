package databricks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	if opts.StreamingFunc != nil {
		return o.stream(ctx, resp.Body, opts)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("bodyBytes: %v\n", string(bodyBytes))
	return o.model.FormatResponse(ctx, bodyBytes)
}

func (o *LLM) stream(ctx context.Context, body io.Reader, opts llms.CallOptions) (*llms.ContentResponse, error) {
	fullChoiceContent := []strings.Builder{}
	scanner := bufio.NewScanner(body)
	finalResponse := &llms.ContentResponse{}

	for scanner.Scan() {
		scannedBytes := scanner.Bytes()
		if len(scannedBytes) == 0 {
			continue
		}

		contentResponse, err := o.model.FormatStreamResponse(ctx, scannedBytes)
		if err != nil {
			return nil, err
		}

		if len(contentResponse.Choices) == 0 {
			continue
		}

		index, err := concatenateAnswers(contentResponse.Choices, &fullChoiceContent)
		if err != nil {
			return nil, err
		}

		if index == nil {
			continue
		}

		if err := opts.StreamingFunc(ctx, []byte(fullChoiceContent[*index].String())); err != nil {
			return nil, err
		}

		finalResponse = contentResponse
	}

	for index := range finalResponse.Choices {
		finalResponse.Choices[index].Content = fullChoiceContent[index].String()
	}

	return finalResponse, nil
}

func concatenateAnswers(choices []*llms.ContentChoice, fullChoiceContent *[]strings.Builder) (*int, error) {
	var lastModifiedIndex *int

	for choiceIndex := range choices {
		if len(*fullChoiceContent) <= choiceIndex {
			*fullChoiceContent = append(*fullChoiceContent, strings.Builder{})
		}

		if choices[choiceIndex].Content == "" {
			continue
		}

		lastModifiedIndex = &choiceIndex

		if _, err := (*fullChoiceContent)[choiceIndex].WriteString(choices[choiceIndex].Content); err != nil {
			return lastModifiedIndex, err
		}
	}

	return lastModifiedIndex, nil
}
