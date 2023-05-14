package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

const (
	defaultChatModel = "gpt-3.5-turbo"
)

// ChatRequest is a request to create an embedding.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        int           `json:"top_p,omitempty"`
	N           int           `json:"n,omitempty"`
	StopWords   []string      `json:"stop,omitempty"`
}

// ChatMessage is a message in a chat request.
type ChatMessage struct {
	// The role of the author of this message. One of system, user, or assistant.
	Role string `json:"role"`
	// The content of the message.
	Content string `json:"content"`
	// The name of the author of this message. May contain a-z, A-Z, 0-9, and underscores,
	// with a maximum length of 64 characters.
	Name string `json:"name,omitempty"`
}

// ChatChoice is a choice in a chat response.
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage is the usage of a chat completion request.
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse is a response to a chat request.
type ChatResponse struct {
	ID      string  `json:"id,omitempty"`
	Created float64 `json:"created,omitempty"`
	Choices []struct {
		FinishReason string      `json:"finish_reason,omitempty"`
		Index        float64     `json:"index,omitempty"`
		Message      ChatMessage `json:"message,omitempty"`
	} `json:"choices,omitempty"`
	Model  string `json:"model,omitempty"`
	Object string `json:"object,omitempty"`
	Usage  struct {
		CompletionTokens float64 `json:"completion_tokens,omitempty"`
		PromptTokens     float64 `json:"prompt_tokens,omitempty"`
		TotalTokens      float64 `json:"total_tokens,omitempty"`
	} `json:"usage,omitempty"`
}

func (c *Client) createChat(ctx context.Context, payload *ChatRequest) (*ChatResponse, error) {
	// Build request payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Build request
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Send request
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Parse response
	var response ChatResponse
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
