package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	defaultCompletionModel = "text-davinci-003"
)

type completionPayload struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Temperature      float64  `json:"temperature,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	N                int      `json:"n,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	StopWords        []string `json:"stop,omitempty"`
}

type completionResponsePayload struct {
	ID      string  `json:"id,omitempty"`
	Created float64 `json:"created,omitempty"`
	Choices []struct {
		FinishReason string      `json:"finish_reason,omitempty"`
		Index        float64     `json:"index,omitempty"`
		Logprobs     interface{} `json:"logprobs,omitempty"`
		Text         string      `json:"text,omitempty"`
	} `json:"choices,omitempty"`
	Model  string `json:"model,omitempty"`
	Object string `json:"object,omitempty"`
	Usage  struct {
		CompletionTokens float64 `json:"completion_tokens,omitempty"`
		PromptTokens     float64 `json:"prompt_tokens,omitempty"`
		TotalTokens      float64 `json:"total_tokens,omitempty"`
	} `json:"usage,omitempty"`
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
}

// nolint:lll
func (c *Client) createCompletion(ctx context.Context, payload *completionPayload) (*completionResponsePayload, error) {
	c.setCompletionDefaults(payload)

	// Build request payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL("/completions", c.Model), bytes.NewReader(payloadBytes))
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

	// Parse response
	var response completionResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}
