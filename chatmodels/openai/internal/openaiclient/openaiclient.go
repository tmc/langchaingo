package openaiclient

import (
	"context"
	"errors"
)

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// Client is a client for the OpenAI API.
type Client struct {
	token string
}

// New returns a new OpenAI client.
func New(token string) (*Client, error) {
	c := &Client{token: token}
	return c, nil
}

// ChatRequest is a request to the text-only chat completion including some configurable parameters.
// Note that the parameters are not exhaustive but only the ones that are used by langchain.
type ChatRequest struct {
	Model string
	// Messages    []schema.ChatMessage // FIXME: incompatible type with LLM
	Messages    []string
	Temperature float64  `json:"temperature"`
	N           int      `json:"n"`
	MaxTokens   int      `json:"max_tokens"`
	Stop        []string `json:"stop,omitempty"`
}

// ChatCompletion result of the text-only chat completion.
type ChatCompletion struct {
	Text string `json:"text"`
}

// Chat creates a new chat completion.
func (c *Client) Chat(ctx context.Context, r *ChatRequest) (*ChatCompletion, error) {
	// set defaults according to langchain
	if r.Model == "" {
		r.Model = "gpt-3.5-turbo"
	}
	if r.Temperature == 0.0 {
		r.Temperature = 0.7
	}
	if r.N == 0 {
		r.N = 1
	}
	// this is a choice by langchaingo
	if r.MaxTokens == 0 {
		r.MaxTokens = 256
	}

	var messages []ChatPayloadMessage

	for _, msg := range r.Messages {
		messages = append(messages, ChatPayloadMessage{
			Content: msg,
			Role:    "user",
			// FIXME: incompatible type with LLM
			// Content: msg.GetText(),
			// Role:    string(msg.GetType()),
		})
	}

	resp, err := c.createChat(ctx, &chatPayload{
		Model:       r.Model,
		Messages:    messages,
		Temperature: r.Temperature,
		N:           r.N,
		MaxTokens:   r.MaxTokens,
		Stop:        r.Stop,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	return &ChatCompletion{
		Text: resp.Choices[0].Message.Content,
	}, nil
}
