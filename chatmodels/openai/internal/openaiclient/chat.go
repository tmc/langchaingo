package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// ChatPayloadMessage is a message in the format expected by the OpenAI API.
type ChatPayloadMessage struct {
	// Role is the role of the message, either "system", "user", or "assistant"
	Role string `json:"role"`
	// Content is the content of the message
	Content string `json:"content"`
}

type chatPayload struct {
	Model       string               `json:"model"`
	Messages    []ChatPayloadMessage `json:"messages"`
	Temperature float64              `json:"temperature"`
	N           int                  `json:"n"`
	MaxTokens   int                  `json:"max_tokens"`
	Stop        []string             `json:"stop,omitempty"`
}

type chatResponsePayload struct {
	ID      string  `json:"id,omitempty"`
	Created float64 `json:"created,omitempty"`
	Choices []struct {
		FinishReason string      `json:"finish_reason,omitempty"`
		Index        float64     `json:"index,omitempty"`
		Logprobs     interface{} `json:"logprobs,omitempty"`
		Message      struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `message:"text,omitempty"`
	} `json:"choices,omitempty"`
	Model  string `json:"model,omitempty"`
	Object string `json:"object,omitempty"`
	Usage  struct {
		CompletionTokens float64 `json:"completion_tokens,omitempty"`
		PromptTokens     float64 `json:"prompt_tokens,omitempty"`
		TotalTokens      float64 `json:"total_tokens,omitempty"`
	} `json:"usage,omitempty"`
}

// Chat creates a chat completion for the given payload.
func (c *Client) createChat(ctx context.Context, payload *chatPayload) (*chatResponsePayload, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	var response chatResponsePayload

	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
