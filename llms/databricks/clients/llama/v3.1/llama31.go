package databricksclientsllama31

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/databricks"
)

// LlamaPayload represents the payload structure for the Llama model.
type Llama31 struct{}

var _ databricks.Model = (*Llama31)(nil)

// NewLlama31 creates a new Llama31 instance.
func NewLlama31() *Llama31 {
	return &Llama31{}
}

// FormatPayload implements databricks.Model to convert langchaingo llms.MessageContent to llama payload.
func (l *Llama31) FormatPayload(_ context.Context, messages []llms.MessageContent, options ...llms.CallOption) ([]byte, error) {
	// Initialize payload options with defaults
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Transform llms.MessageContent to LlamaMessage
	llamaMessages := []LlamaMessage{}
	for _, msg := range messages {
		var contentBuilder strings.Builder
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				contentBuilder.WriteString(p.Text)
			case llms.ImageURLContent:
				contentBuilder.WriteString(fmt.Sprintf("[Image: %s]", p.URL))
			case llms.BinaryContent:
				contentBuilder.WriteString(fmt.Sprintf("[Binary Content: %s]", p.MIMEType))
			default:
				return nil, fmt.Errorf("unsupported content part type: %T", p)
			}
		}

		llamaMessages = append(llamaMessages, LlamaMessage{
			Role:    MapRole(msg.Role),
			Content: contentBuilder.String(),
		})
	}

	// Construct the LlamaPayload
	payload := LlamaPayload{
		Model:            "llama-3.1",
		Messages:         llamaMessages,
		Temperature:      opts.Temperature,
		MaxTokens:        opts.MaxTokens,
		TopP:             opts.TopP,
		FrequencyPenalty: opts.FrequencyPenalty,
		PresencePenalty:  opts.PresencePenalty,
		Stop:             opts.StopWords, // Add stop sequences if needed
	}

	if opts.StreamingFunc != nil {
		payload.Stream = true
	}

	// Serialize to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return jsonPayload, nil
}

// FormatResponse parses the LlamaResponse JSON and converts it to a ContentResponse structure.
func (l *Llama31) FormatResponse(_ context.Context, response []byte) (*llms.ContentResponse, error) {
	return formatResponse[LlamaChoice](response)
}

// FormatStreamResponse parses the LlamaResponse JSON and converts it to a ContentResponse structure.
func (l *Llama31) FormatStreamResponse(_ context.Context, response []byte) (*llms.ContentResponse, error) {
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
	return formatResponse[LlamaChoiceDelta](response)
}

func formatResponse[T LlamaChoiceDelta | LlamaChoice](response []byte) (*llms.ContentResponse, error) {
	// Parse the LlamaResponse JSON
	var llamaResp LlamaResponse[T]
	err := json.Unmarshal(response, &llamaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal LlamaResponse: %w", err)
	}

	// Initialize ContentResponse
	contentResponse := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{},
	}

	// Map LlamaResponse choices to ContentChoice
	for _, llamaChoice := range llamaResp.Choices {
		var contentChoice *llms.ContentChoice
		switch choice := any(llamaChoice).(type) {
		case LlamaChoice:
			contentChoice = &llms.ContentChoice{
				Content:    choice.Message.Content,
				StopReason: choice.FinishReason,
				GenerationInfo: map[string]any{
					"index": choice.Index,
				},
			}

			// If the LlamaMessage indicates a function/tool call, populate FuncCall or ToolCalls
			if choice.Message.Role == RoleIPython {
				funcCall := &llms.FunctionCall{
					Name:      "tool_function_name",   // Replace with actual function name if included in response
					Arguments: choice.Message.Content, // Replace with parsed arguments if available
				}
				contentChoice.FuncCall = funcCall
				contentChoice.ToolCalls = []llms.ToolCall{
					{
						ID:           fmt.Sprintf("tool-call-%d", choice.Index),
						Type:         "function",
						FunctionCall: funcCall,
					},
				}
			}
		case LlamaChoiceDelta:
			contentChoice = &llms.ContentChoice{
				Content:    choice.Delta.Content,
				StopReason: choice.FinishReason,
				GenerationInfo: map[string]any{
					"index": choice.Index,
				},
			}
		}

		// Append the ContentChoice to the ContentResponse
		contentResponse.Choices = append(contentResponse.Choices, contentChoice)
	}

	return contentResponse, nil
}
