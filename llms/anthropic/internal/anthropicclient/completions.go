package anthropicclient

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
)

const (
	defaultCompletionModel = "claude-instant-1"
)

type completionPayload struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens_to_sample,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	StopWords   []string `json:"stop_sequences,omitempty"`
	Stream      bool     `json:"stream,omitempty"`

	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type CompletionResponsePayload struct {
	Completion string `json:"completion,omitempty"`
	LogID      string `json:"log_id,omitempty"`
	Model      string `json:"model,omitempty"`
	Stop       string `json:"stop,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (c *Client) setCompletionDefaults(payload *completionPayload) {
	// Set defaults
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 256
	}

	if len(payload.StopWords) == 0 {
		payload.StopWords = nil
	}

	switch {
	// Prefer the model specified in the payload.
	case payload.Model != "":

	// If no model is set in the payload, take the one specified in the client.
	case c.Model != "":
		payload.Model = c.Model
	// Fallback: use the default model
	default:
		payload.Model = defaultCompletionModel
	}
	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
}

// nolint:lll
func (c *Client) createCompletion(ctx context.Context, payload *completionPayload) (*CompletionResponsePayload, error) {
	c.setCompletionDefaults(payload)

	// Build request payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}

	url := fmt.Sprintf("%s/complete", c.baseURL)
	// Build request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)

	// Send request
	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg) // nolint:goerr113
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
	}
	if payload.StreamingFunc != nil {
		// Read chunks
		return parseStreamingCompletionResponse(ctx, r, payload)
	}

	// Parse response
	var response CompletionResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

func parseStreamingCompletionResponse(ctx context.Context, r *http.Response, payload *completionPayload) (*CompletionResponsePayload, error) { // nolint:lll
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan *CompletionResponsePayload)
	go func() {
		defer close(responseChan)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			streamPayload := &CompletionResponsePayload{}
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
	response := CompletionResponsePayload{}

	var lastResponse *CompletionResponsePayload
	for streamResponse := range responseChan {
		response.Completion += streamResponse.Completion
		if payload.StreamingFunc != nil {
			err := payload.StreamingFunc(ctx, []byte(streamResponse.Completion))
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
		lastResponse = streamResponse
	}
	response.Model = lastResponse.Model
	response.LogID = lastResponse.LogID
	response.Stop = lastResponse.Stop
	response.StopReason = lastResponse.StopReason
	return &response, nil
}
