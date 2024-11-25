package databricksclientsmistralv1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/databricks"
)

// Mistral1 represents the payload structure for the Mistral model.
type Mistral1 struct {
	Model string `json:"model"`
}

var _ databricks.Model = (*Mistral1)(nil)

// NewMistral1 creates a new Mistral1 instance.
func NewMistral1(model string) *Mistral1 {
	return &Mistral1{
		Model: model,
	}
}

// FormatPayload implements databricks.Model to convert langchaingo llms.MessageContent to llama payload.
func (l *Mistral1) FormatPayload(_ context.Context, messages []llms.MessageContent, options ...llms.CallOption) ([]byte, error) {
	// Initialize payload options with defaults
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Convert langchaingo MessageContent to Mistral ChatMessage
	var mistralMessages []ChatMessage // nolint: prealloc
	for _, msg := range messages {
		// Process parts to handle the actual content and tool calls correctly
		var toolCalls []ToolCall
		var contentParts []string
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				// Text parts go directly as content
				contentParts = append(contentParts, p.Text)
			case llms.ImageURLContent:
				// Append structured description for the image
				contentParts = append(contentParts, fmt.Sprintf("Image: %s", p.URL))
				if p.Detail != "" {
					contentParts = append(contentParts, fmt.Sprintf("Detail: %s", p.Detail))
				}
			case llms.ToolCall:
				// Convert tool calls into structured Mistral ToolCall objects
				toolCalls = append(toolCalls, ToolCall{
					ID:   p.ID,
					Type: ToolTypeFunction, // Assuming ToolTypeFunction
					Function: FunctionCall{
						Name:      p.FunctionCall.Name,
						Arguments: p.FunctionCall.Arguments,
					},
				})
			case llms.ToolCallResponse:
				// Handle tool call responses as content
				contentParts = append(contentParts, p.Content)
			default:
				return nil, fmt.Errorf("unknown content part type: %T", part)
			}
		}

		mistralMessage := ChatMessage{
			Role:      MapRole(msg.Role),
			Content:   fmt.Sprintf("%s", contentParts),
			ToolCalls: toolCalls,
		}
		mistralMessages = append(mistralMessages, mistralMessage)
	}

	// Handle options (example: temperature, max_tokens, etc.)
	payload := ChatCompletionPayload{
		Model:       "mistral-7b", // Replace with the desired model
		Messages:    mistralMessages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		TopP:        opts.TopP,
		RandomSeed:  opts.Seed,
	}

	if opts.StreamingFunc != nil {
		payload.Stream = true
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return payloadBytes, nil
}

// Refactored FormatResponse.
func (l *Mistral1) FormatResponse(_ context.Context, response []byte) (*llms.ContentResponse, error) {
	var resp ChatCompletionResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &llms.ContentResponse{
		Choices: mapChoices(resp.Choices),
	}, nil
}

// Refactored FormatStreamResponse.
func (l *Mistral1) FormatStreamResponse(_ context.Context, response []byte) (*llms.ContentResponse, error) {
	// The "data:" prefix is commonly used in Server-Sent Events (SSE) or streaming APIs
	// to delimit individual chunks of data being sent from the server. It indicates
	// that the following text is a data payload. Before parsing the JSON, we remove
	// this prefix to work with the raw JSON payload.
	response = bytes.TrimPrefix(response, []byte("data: "))

	if string(response) == "[DONE]" || len(response) == 0 {
		return &llms.ContentResponse{
			Choices: []*llms.ContentChoice{{
				Content: "",
			}},
		}, nil
	}

	var streamResp ChatCompletionStreamResponse
	if err := json.Unmarshal(response, &streamResp); err != nil {
		return nil, fmt.Errorf("failed to parse streaming response: %w", err)
	}

	return &llms.ContentResponse{
		Choices: mapChoices(streamResp.Choices),
	}, nil
}

// Helper function to map choices.
func mapChoices[T ChatCompletionResponseChoice | ChatCompletionResponseChoiceStream](choices []T) []*llms.ContentChoice {
	var contentChoices []*llms.ContentChoice // nolint: prealloc

	for _, choice := range choices {
		var index int
		var message ChatMessage
		var finishReason FinishReason
		switch c := any(choice).(type) {
		case ChatCompletionResponseChoice:
			index = c.Index
			message = c.Message
			finishReason = c.FinishReason
		case ChatCompletionResponseChoiceStream:
			index = c.Index
			message = c.Delta
			finishReason = c.FinishReason
		}

		contentChoice := &llms.ContentChoice{
			Content:    message.Content,
			StopReason: string(finishReason),
			GenerationInfo: map[string]any{
				"index": index,
			},
		}

		for _, toolCall := range message.ToolCalls {
			contentChoice.ToolCalls = append(contentChoice.ToolCalls, llms.ToolCall{
				ID:   toolCall.ID,
				Type: string(toolCall.Type),
				FunctionCall: &llms.FunctionCall{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			})
		}

		contentChoices = append(contentChoices, contentChoice)
	}

	return contentChoices
}
