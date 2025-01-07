package gigachatclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type PayloadMessage struct {
	Role            string   `json:"role"`
	Content         any      `json:"content"`
	FunctionStateId string   `json:"functions_state_id,omitempty"`
	DataForContext  []any    `json:"data_for_context,omitempty"`
	Attachments     []string `json:"attachments,omitempty"`
}

type ChatPayload struct {
	Model             string           `json:"model"`
	Messages          []PayloadMessage `json:"messages"`
	FunctionCalls     any              `json:"function_calls,omitempty"`
	Functions         []Function       `json:"functions,omitempty"`
	Temperature       float64          `json:"temperature,omitempty"`
	TopP              float64          `json:"top_p,omitempty"`
	Stream            bool             `json:"stream,omitempty"`
	MaxTokens         int              `json:"max_tokens,omitempty"`
	RepetitionPenalty float64          `json:"repetition_penalty,omitempty"`
	UpdateInterval    float64          `json:"update_interval,omitempty"`
}

// Function used for the request message payload.
type Function struct {
	Name             string                `json:"name"`
	Description      string                `json:"description,omitempty"`
	Parameters       map[string]any        `json:"parameters"`
	FewShotExamples  []FunctionCallExample `json:"few_shot_examples,omitempty"`
	ReturnParameters any                   `json:"return_parameters,omitempty"`
}

type FunctionCallExample struct {
	Request string         `json:"request"`
	Params  map[string]any `json:"params"`
}

// ChatResponse represents the top-level response from the GigaChat API
type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Object  string   `json:"object"`
}

// Choice represents a single response choice from the model
type Choice struct {
	Message      ResponseMessage `json:"message"`
	Index        int             `json:"index"`
	FinishReason string          `json:"finish_reason"` // "stop", "length", "function_call", "blacklist", "error"
}

// ResponseMessage represents the generated message content
type ResponseMessage struct {
	Role             string        `json:"role"` // "assistant" or "function_in_progress"
	Content          string        `json:"content"`
	FunctionsStateID string        `json:"functions_state_id,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
	DataForContext   []any         `json:"data_for_context,omitempty"`
}

// FunctionCall represents a function call in the response
type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (c *Client) setMessageDefaults(payload *ChatPayload) {
	// Set defaults
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 2048
	}

	switch {
	// Prefer the model specified in the payload.
	case payload.Model != "":

	// If no model is set in the payload, take the one specified in the client.
	case c.Model != "":
		payload.Model = c.Model
	// Fallback: use the default model
	default:
		payload.Model = DefaultModel
	}
}

func (c *Client) CreateCompletion(ctx context.Context, payload *ChatPayload) (*ChatResponse, error) {
	c.setMessageDefaults(payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := c.do(ctx, "/chat/completions", payloadBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}
