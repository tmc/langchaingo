package databricksclientsllama31

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/databricks"
)

type Llama31 struct{}

var _ databricks.Model = (*Llama31)(nil)

func NewLlama31() *Llama31 {
	return &Llama31{}
}

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
		payload.Streaming = true
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
	return formatResponse(response)
}

// FormatStreamResponse parses the LlamaResponse JSON and converts it to a ContentResponse structure.
func (l *Llama31) FormatStreamResponse(_ context.Context, response []byte) (*llms.ContentResponse, error) {
	return formatResponse(response)
}

func formatResponse(response []byte) (*llms.ContentResponse, error) {
	// Parse the LlamaResponse JSON
	var llamaResp LlamaResponse
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
		contentChoice := &llms.ContentChoice{
			Content:    llamaChoice.Message.Content,
			StopReason: llamaChoice.FinishReason,
			GenerationInfo: map[string]any{
				"index": llamaChoice.Index,
			},
		}

		// If the LlamaMessage indicates a function/tool call, populate FuncCall or ToolCalls
		if llamaChoice.Message.Role == RoleIPython {
			funcCall := &llms.FunctionCall{
				Name:      "tool_function_name",        // Replace with actual function name if included in response
				Arguments: llamaChoice.Message.Content, // Replace with parsed arguments if available
			}
			contentChoice.FuncCall = funcCall
			contentChoice.ToolCalls = []llms.ToolCall{
				{
					ID:           fmt.Sprintf("tool-call-%d", llamaChoice.Index),
					Type:         "function",
					FunctionCall: funcCall,
				},
			}
		}

		contentResponse.Choices = append(contentResponse.Choices, contentChoice)
	}

	return contentResponse, nil
}
