package anthropicclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type completionPayload struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature"`
	MaxTokens   int      `json:"max_tokens_to_sample,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	StopWords   []string `json:"stop_sequences,omitempty"`
	Stream      bool     `json:"stream,omitempty"`

	StreamingFunc streaming.Callback `json:"-"`
}

type CompletionResponsePayload struct {
	Completion string `json:"completion,omitempty"`
	LogID      string `json:"log_id,omitempty"`
	Model      string `json:"model,omitempty"`
	Stop       string `json:"stop,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

func (c *Client) setCompletionDefaults(payload *completionPayload) {
	// Set defaults
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 2048
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
		payload.Model = defaultModel
	}
	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
}

// nolint:lll
func (c *Client) createCompletion(ctx context.Context, payload *completionPayload) (*CompletionResponsePayload, error) {
	c.setCompletionDefaults(payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := c.do(ctx, "/complete", payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	if payload.StreamingFunc != nil {
		return parseStreamingCompletionResponse(ctx, resp, payload)
	}

	var response CompletionResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

type CompletionEvent struct {
	Response *CompletionResponsePayload
	Err      error
}

func parseStreamingCompletionResponse(ctx context.Context, r *http.Response, payload *completionPayload) (*CompletionResponsePayload, error) { // nolint:lll
	scanner := bufio.NewScanner(r.Body)
	scanner.Buffer(make([]byte, 0, initialStreamBuffer), maxStreamLine)
	responseChan := make(chan CompletionEvent)
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
				responseChan <- CompletionEvent{Response: nil, Err: fmt.Errorf("failed to parse stream event: %w", err)}
				return
			}
			responseChan <- CompletionEvent{Response: streamPayload, Err: nil}
		}
		if err := scanner.Err(); err != nil {
			responseChan <- CompletionEvent{Response: nil, Err: fmt.Errorf("failed to read stream: %w", err)}
		}
	}()
	// Parse response
	response := CompletionResponsePayload{}
	defer streaming.CallWithDone(ctx, payload.StreamingFunc) //nolint:errcheck

	var lastResponse *CompletionResponsePayload
	for streamResponse := range responseChan {
		if streamResponse.Err != nil {
			return nil, streamResponse.Err
		}
		content := streamResponse.Response.Completion
		response.Completion += content
		if err := streaming.CallWithText(ctx, payload.StreamingFunc, content); err != nil {
			return nil, fmt.Errorf("streaming func returned an error: %w", err)
		}
		lastResponse = streamResponse.Response
	}
	if lastResponse == nil {
		return nil, ErrEmptyResponse
	}
	response.Model = lastResponse.Model
	response.LogID = lastResponse.LogID
	response.Stop = lastResponse.Stop
	response.StopReason = lastResponse.StopReason
	return &response, nil
}
