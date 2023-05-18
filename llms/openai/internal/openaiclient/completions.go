package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	defaultCompletionModel = "text-davinci-003"
)

type completionPayload struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature int      `json:"temperature"`
	MaxTokens   int      `json:"max_tokens"`
	StopWords   []string `json:"stop,omitempty"`
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

func (c *Client) createCompletion(ctx context.Context, payload *completionPayload) (*completionResponsePayload, error) {
	// Set defaults
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 256
	}

	if len(payload.StopWords) == 0 {
		payload.StopWords = nil
	}

	// Build request payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/completions", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Send request
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d", r.StatusCode)
	}

	// Parse response
	var response completionResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}
