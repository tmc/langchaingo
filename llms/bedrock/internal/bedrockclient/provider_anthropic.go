package bedrockclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-anthropic-claude-messages.html
// Also: https://docs.anthropic.com/claude/reference/messages_post

// anthropicBinGenerationInputSource is the source of the content.
type anthropicBinGenerationInputSource struct {
	// The type of the source. Required
	// One of: "base64"
	Type string `json:"type"`
	// The MIME type of the source. Required
	// One of: []"image/jpeg", "image/png", "image/gif", "image/bmp", "image/webp"]
	MediaType string `json:"media_type"`
	// The data of the source. Required
	// For example if type is "base64" then data is a base64 encoded string
	Data string `json:"data"`
}

// anthropicTextGenerationInputContent is a single message in the input.
type anthropicTextGenerationInputContent struct {
	// The type of the content. Required.
	// One of: "text", "image", "tool_use", "tool_result", "thinking"
	Type string `json:"type"`
	// The source of the content. Required if type is "image"
	Source *anthropicBinGenerationInputSource `json:"source,omitempty"`
	// The text content. Required if type is "text"
	Text string `json:"text,omitempty"`
	// Tool use fields
	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
	// Tool result fields
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	// Thinking fields
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
	Data      string `json:"data,omitempty"`
	// Cache control for prompt caching (Bedrock supports Anthropic cache_control format)
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

type anthropicCacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

type anthropicTextGenerationInputMessage struct {
	// The role of the message. Required
	// One of: ["user", "assistant"]
	// For system prompt, use the system field in the input
	Role string `json:"role"`
	// The content of the message. Required
	Content []anthropicTextGenerationInputContent `json:"content"`
}

// anthropicTextGenerationInput is the input to the model.
type anthropicTextGenerationInput struct {
	// The version of the model to use. Required
	AnthropicVersion string `json:"anthropic_version"`
	// The maximum number of tokens to generate per result. Required
	MaxTokens int `json:"max_tokens"`
	// The system prompt to use. Optional
	// Can be string or []anthropicTextGenerationInputContent for caching support
	System any `json:"system,omitempty"`
	// The messages to use. Required
	Messages []*anthropicTextGenerationInputMessage `json:"messages"`
	// The amount of randomness injected into the response. Optional, default = 1
	Temperature float64 `json:"temperature,omitempty"`
	// The probability mass from which tokens are sampled. Optional, default = 1
	TopP float64 `json:"top_p,omitempty"`
	// Only sample from the top K options for each subsequent token.
	// Use top_k to remove long tail low probability responses.
	// Optional, default = 250
	TopK int `json:"top_k,omitempty"`
	// Sequences that will cause the model to stop generating tokens. Optional
	StopSequences []string `json:"stop_sequences,omitempty"`
	// Tools available for the model to use. Optional
	Tools []anthropicTool `json:"tools,omitempty"`
	// Thinking configuration for reasoning models. Optional
	Thinking *anthropicThinkingPayload `json:"thinking,omitempty"`
}

type anthropicTool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	InputSchema  any                    `json:"input_schema,omitempty"`
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

type anthropicThinkingPayload struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// anthropicTextGenerationOutput is the generated output.
type anthropicTextGenerationOutput struct {
	// Type of the content.
	// For messages, it is "message"
	Type string `json:"type"`
	// Conversational role of the generated message.
	// This will always be "assistant".
	Role string `json:"role"`
	// This is an array of content blocks, each of which has a type that determines its shape.
	// Currently, the only type in responses is "text" or "tool_use".
	Content []anthropicContentBlock `json:"content"`
	// The reason for the completion of the generation.
	// One of: ["end_turn", "max_tokens", "stop_sequence", "tool_use"]
	StopReason string `json:"stop_reason"`
	// Which custom stop sequence was matched, if any.
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens              int32 `json:"input_tokens"`
		OutputTokens             int32 `json:"output_tokens"`
		CacheCreationInputTokens int32 `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int32 `json:"cache_read_input_tokens,omitempty"`
		CacheCreation            struct {
			Ephemeral5mInputTokens int32 `json:"ephemeral_5m_input_tokens,omitempty"`
			Ephemeral1hInputTokens int32 `json:"ephemeral_1h_input_tokens,omitempty"`
		} `json:"cache_creation,omitempty"`
	} `json:"usage"`
}

type anthropicContentBlock struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	Thinking  string         `json:"thinking,omitempty"`
	Signature string         `json:"signature,omitempty"`
	Data      string         `json:"data,omitempty"`
}

// Finish reason for the completion of the generation.
const (
	AnthropicCompletionReasonEndTurn      = "end_turn"
	AnthropicCompletionReasonMaxTokens    = "max_tokens"
	AnthropicCompletionReasonStopSequence = "stop_sequence"
)

// The latest version of the model.
const (
	AnthropicLatestVersion = "bedrock-2023-05-31"
)

// Role attribute for the anthropic message.
const (
	AnthropicSystem        = "system"
	AnthropicRoleUser      = "user"
	AnthropicRoleAssistant = "assistant"
)

// Type attribute for the anthropic message.
const (
	AnthropicMessageTypeText       = "text"
	AnthropicMessageTypeImage      = "image"
	AnthropicMessageTypeToolUse    = "tool_use"
	AnthropicMessageTypeToolResult = "tool_result"
)

func createAnthropicCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	inputContents, systemPrompt, err := processInputMessagesAnthropic(messages)
	if err != nil {
		return nil, err
	}

	tools := make([]anthropicTool, len(options.Tools))
	for i, tool := range options.Tools {
		tools[i] = anthropicTool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters,
		}
	}

	// Prepare system prompt - omit if empty
	var system any
	if systemPrompt != "" {
		system = systemPrompt
	}

	input := anthropicTextGenerationInput{
		AnthropicVersion: AnthropicLatestVersion,
		MaxTokens:        getMaxTokens(options.GetMaxTokens(), 2048),
		System:           system,
		Messages:         inputContents,
		Temperature:      options.GetTemperature(),
		TopP:             options.GetTopP(),
		TopK:             options.GetTopK(),
		StopSequences:    options.StopWords,
		Tools:            tools,
	}

	// Add thinking configuration for reasoning models
	if options.Reasoning != nil && supportsAnthropicReasoning(modelID) {
		maxTokens := options.GetMaxTokens()
		if maxTokens == 0 {
			maxTokens = 2048
		}
		tokens := options.Reasoning.GetTokens(maxTokens)
		if tokens > 0 {
			input.Thinking = &anthropicThinkingPayload{
				Type:         "enabled",
				BudgetTokens: tokens,
			}
		}
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if options.StreamingFunc != nil {
		modelInput := &bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String(modelID),
			Accept:      aws.String("*/*"),
			ContentType: aws.String("application/json"),
			Body:        body,
		}
		return parseStreamingCompletionResponse(ctx, client, modelInput, options)
	}

	modelInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Accept:      aws.String("*/*"),
		ContentType: aws.String("application/json"),
		Body:        body,
	}
	resp, err := client.InvokeModel(ctx, modelInput)
	if err != nil {
		return nil, err
	}

	var output anthropicTextGenerationOutput
	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	if len(output.Content) == 0 {
		return nil, errors.New("no results")
	}

	Contentchoices := make([]*llms.ContentChoice, 0)

	// Group content blocks by type for this choice
	var textContent string
	var reasoningContent string
	var signature []byte
	var toolCalls []llms.ToolCall

	for _, c := range output.Content {
		switch c.Type {
		case "text":
			textContent += c.Text
		case "thinking":
			reasoningContent += c.Thinking
			if len(c.Signature) > 0 {
				signature = []byte(c.Signature)
			}
		case "tool_use":
			argumentsJSON, err := json.Marshal(c.Input)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tool use arguments: %w", err)
			}
			toolCalls = append(toolCalls, llms.ToolCall{
				ID:   c.ID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      c.Name,
					Arguments: string(argumentsJSON),
				},
			})
		}
	}

	// Create single choice with all content
	choice := &llms.ContentChoice{
		Content:    textContent,
		Reasoning:  processReasoning(reasoningContent, signature),
		ToolCalls:  toolCalls,
		StopReason: output.StopReason,
		GenerationInfo: map[string]any{
			"input_tokens":     output.Usage.InputTokens,
			"output_tokens":    output.Usage.OutputTokens,
			"PromptTokens":     output.Usage.InputTokens,
			"CompletionTokens": output.Usage.OutputTokens,
			"TotalTokens":      output.Usage.InputTokens + output.Usage.OutputTokens,
			// Cache metrics
			"CacheReadInputTokens":                output.Usage.CacheReadInputTokens,
			"CacheCreationInputTokens":            output.Usage.CacheCreationInputTokens,
			"CacheCreationEphemeral5mInputTokens": output.Usage.CacheCreation.Ephemeral5mInputTokens,
			"CacheCreationEphemeral1hInputTokens": output.Usage.CacheCreation.Ephemeral1hInputTokens,
			"PromptCachedTokens":                  output.Usage.CacheReadInputTokens,
		},
	}
	Contentchoices = append(Contentchoices, choice)

	return &llms.ContentResponse{
		Choices: Contentchoices,
	}, nil
}

func processReasoning(reasoningContent string, signature []byte) *reasoning.ContentReasoning {
	if reasoningContent == "" && len(signature) == 0 {
		return nil
	}

	return &reasoning.ContentReasoning{
		Content:   reasoningContent,
		Signature: signature,
	}
}

type streamingCompletionResponseChunk struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type         string `json:"type"`
		Text         string `json:"text"`
		PartialJSON  string `json:"partial_json"`
		StopReason   string `json:"stop_reason"`
		StopSequence any    `json:"stop_sequence"`
		Thinking     string `json:"thinking,omitempty"`
		Signature    string `json:"signature,omitempty"`
	} `json:"delta"`
	ContentBlock struct {
		Type  string         `json:"type"`
		ID    string         `json:"id"`
		Name  string         `json:"name"`
		Input map[string]any `json:"input"`
	} `json:"content_block"`
	AmazonBedrockInvocationMetrics struct {
		InputTokenCount   int32 `json:"inputTokenCount"`
		OutputTokenCount  int32 `json:"outputTokenCount"`
		InvocationLatency int32 `json:"invocationLatency"`
		FirstByteLatency  int32 `json:"firstByteLatency"`
	} `json:"amazon-bedrock-invocationMetrics"`
	Usage struct {
		OutputTokens int32 `json:"output_tokens"`
	} `json:"usage"`
	Message struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Role         string `json:"role"`
		Content      []any  `json:"content"`
		Model        string `json:"model"`
		StopReason   any    `json:"stop_reason"`
		StopSequence any    `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int32 `json:"input_tokens"`
			OutputTokens int32 `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

func parseStreamingCompletionResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
	output, err := client.InvokeModelWithResponseStream(ctx, modelInput)
	if err != nil {
		return nil, err
	}
	stream := output.GetStream()
	if stream == nil {
		return nil, errors.New("no stream")
	}
	defer stream.Close()
	defer streaming.CallWithDone(ctx, options.StreamingFunc) //nolint:errcheck

	contentchoices := []*llms.ContentChoice{{GenerationInfo: map[string]any{}}}
	var currentToolCall *streaming.ToolCall
	var toolCalls []llms.ToolCall
	var signature strings.Builder

	for e := range stream.Events() {
		if err = stream.Err(); err != nil {
			return nil, err
		}

		if v, ok := e.(*types.ResponseStreamMemberChunk); ok {
			var resp streamingCompletionResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			switch resp.Type {
			case "message_start":
				contentchoices[0].GenerationInfo["input_tokens"] = resp.Message.Usage.InputTokens
				contentchoices[0].GenerationInfo["PromptTokens"] = resp.Message.Usage.InputTokens
				contentchoices[0].GenerationInfo["TotalTokens"] = resp.Message.Usage.InputTokens
			case "content_block_start":
				if resp.ContentBlock.Type == "tool_use" {
					currentToolCall = &streaming.ToolCall{
						ID:   resp.ContentBlock.ID,
						Name: resp.ContentBlock.Name,
					}
				}
			case "content_block_delta":
				switch resp.Delta.Type {
				case "text_delta":
					if err = streaming.CallWithText(ctx, options.StreamingFunc, resp.Delta.Text); err != nil {
						return nil, err
					}
					contentchoices[0].Content += resp.Delta.Text
				case "thinking_delta":
					if resp.Delta.Thinking != "" {
						chunk := streaming.Chunk{
							Type:      streaming.ChunkTypeReasoning,
							Reasoning: &reasoning.ContentReasoning{Content: resp.Delta.Thinking},
						}
						if err = options.StreamingFunc(ctx, chunk); err != nil {
							return nil, err
						}
						contentchoices[0].Reasoning = appendReasoning(contentchoices[0].Reasoning, resp.Delta.Thinking)
					}
					if resp.Delta.Signature != "" {
						signature.WriteString(resp.Delta.Signature)
					}
				case "signature_delta":
					if resp.Delta.Signature != "" {
						signature.WriteString(resp.Delta.Signature)
					}
				case "input_json_delta":
					if currentToolCall != nil {
						// Bedrock already sends deltas in PartialJSON, not full accumulated JSON
						delta := resp.Delta.PartialJSON

						// Accumulate for final tool call response
						currentToolCall.Arguments += delta

						// Send delta to maintain compatibility with original Anthropic behavior
						if delta != "" {
							deltaToolCall := streaming.NewToolCall(currentToolCall.ID, currentToolCall.Name, delta)
							if err = streaming.CallWithToolCall(ctx, options.StreamingFunc, deltaToolCall); err != nil {
								return nil, err
							}
						}
					}
				}
			case "content_block_stop":
				if currentToolCall != nil {
					// Add completed tool call to final response
					toolCalls = append(toolCalls, llms.ToolCall{
						ID:   currentToolCall.ID,
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      currentToolCall.Name,
							Arguments: currentToolCall.Arguments,
						},
					})
					currentToolCall = nil
				}
			case "message_delta":
				contentchoices[0].StopReason = resp.Delta.StopReason
				inputTokens := resp.Message.Usage.InputTokens
				outputTokens := resp.Message.Usage.OutputTokens
				if inputTokens == 0 {
					if v, ok := contentchoices[0].GenerationInfo["input_tokens"].(int32); ok {
						inputTokens = v
					}
				}
				contentchoices[0].GenerationInfo["output_tokens"] = outputTokens
				contentchoices[0].GenerationInfo["CompletionTokens"] = outputTokens
				contentchoices[0].GenerationInfo["TotalTokens"] = inputTokens + outputTokens
			}
		}
	}
	if err = stream.Err(); err != nil {
		return nil, err
	}

	// Add tool calls to the final response
	contentchoices[0].ToolCalls = toolCalls

	// Add signature to reasoning if accumulated
	if signature.Len() > 0 && contentchoices[0].Reasoning != nil {
		contentchoices[0].Reasoning.Signature = []byte(signature.String())
	}

	return &llms.ContentResponse{
		Choices: contentchoices,
	}, nil
}

func appendReasoning(reasoning *reasoning.ContentReasoning, reasoningContent string) *reasoning.ContentReasoning {
	if reasoning == nil {
		return processReasoning(reasoningContent, nil)
	}

	reasoning.Content += reasoningContent
	return reasoning
}

// process the input messages to anthropic supported input
// returns the input content and system prompt.
func processInputMessagesAnthropic(messages []Message) ([]*anthropicTextGenerationInputMessage, string, error) {
	chunkedMessages := make([][]Message, 0, len(messages))
	currentChunk := make([]Message, 0, len(messages))
	var lastRole llms.ChatMessageType
	for _, message := range messages {
		if message.Role != lastRole {
			if len(currentChunk) > 0 {
				chunkedMessages = append(chunkedMessages, currentChunk)
			}
			currentChunk = make([]Message, 0, len(messages))
		}
		currentChunk = append(currentChunk, message)
		lastRole = message.Role
	}
	if len(currentChunk) > 0 {
		chunkedMessages = append(chunkedMessages, currentChunk)
	}

	inputContents := make([]*anthropicTextGenerationInputMessage, 0, len(messages))
	var systemPrompt string
	for _, chunk := range chunkedMessages {
		role, err := getAnthropicRole(chunk[0].Role)
		if err != nil {
			return nil, "", err
		}
		if role == AnthropicSystem {
			if systemPrompt != "" {
				return nil, "", errors.New("multiple system prompts")
			}
			for _, message := range chunk {
				c := getAnthropicInputContent(message)
				if c.Type != AnthropicMessageTypeText {
					return nil, "", errors.New("system prompt must be text")
				}
				systemPrompt += c.Text
			}
			continue
		}
		content := make([]anthropicTextGenerationInputContent, 0, len(chunk))
		for _, message := range chunk {
			// For AI messages with reasoning, add thinking blocks before text
			if message.Role == llms.ChatMessageTypeAI && message.Reasoning != nil {
				// Add thinking block if present
				if message.Reasoning.Content != "" || len(message.Reasoning.Signature) > 0 {
					thinkingBlock := anthropicTextGenerationInputContent{
						Type:     "thinking",
						Thinking: message.Reasoning.Content,
					}
					if len(message.Reasoning.Signature) > 0 {
						thinkingBlock.Signature = string(message.Reasoning.Signature)
					}
					content = append(content, thinkingBlock)
				}
			}
			// Add regular content (text, tool_use, tool_result, etc.)
			content = append(content, getAnthropicInputContent(message))
		}
		inputContents = append(inputContents, &anthropicTextGenerationInputMessage{
			Role:    role,
			Content: content,
		})
	}
	return inputContents, systemPrompt, nil
}

// process the role of the message to anthropic supported role.
func getAnthropicRole(role llms.ChatMessageType) (string, error) {
	switch role {
	case llms.ChatMessageTypeSystem:
		return AnthropicSystem, nil

	case llms.ChatMessageTypeAI:
		return AnthropicRoleAssistant, nil

	case llms.ChatMessageTypeGeneric:
		fallthrough
	case llms.ChatMessageTypeHuman:
		return AnthropicRoleUser, nil
	case llms.ChatMessageTypeTool:
		return AnthropicRoleUser, nil
	case llms.ChatMessageTypeFunction:
		fallthrough
	default:
		return "", errors.New("role not supported")
	}
}

func getAnthropicInputContent(message Message) anthropicTextGenerationInputContent {
	var c anthropicTextGenerationInputContent
	switch message.Type {
	case AnthropicMessageTypeText:
		c = anthropicTextGenerationInputContent{
			Type: message.Type,
			Text: message.Content,
		}
		if message.CacheControl != nil {
			c.CacheControl = &anthropicCacheControl{
				Type: message.CacheControl.Type,
				TTL:  message.CacheControl.TTL,
			}
		}
	case AnthropicMessageTypeImage:
		c = anthropicTextGenerationInputContent{
			Type: message.Type,
			Source: &anthropicBinGenerationInputSource{
				Type:      "base64",
				MediaType: message.MimeType,
				Data:      base64.StdEncoding.EncodeToString([]byte(message.Content)),
			},
		}
	case AnthropicMessageTypeToolUse:
		if message.ToolCall != nil {
			c = anthropicTextGenerationInputContent{
				Type:  message.Type,
				ID:    message.ToolCall.ID,
				Name:  message.ToolCall.Name,
				Input: message.ToolCall.Arguments,
			}
		}
	case AnthropicMessageTypeToolResult:
		if message.ToolResult != nil {
			c = anthropicTextGenerationInputContent{
				Type:      message.Type,
				ToolUseID: message.ToolResult.ToolCallID,
				Content:   message.ToolResult.Content,
			}
			if message.CacheControl != nil {
				c.CacheControl = &anthropicCacheControl{
					Type: message.CacheControl.Type,
					TTL:  message.CacheControl.TTL,
				}
			}
		}
	}
	return c
}

// supportsAnthropicReasoning checks if the model supports reasoning
func supportsAnthropicReasoning(modelID string) bool {
	reasoningModels := []string{
		"anthropic.claude-opus-4-",
		"anthropic.claude-sonnet-4-",
		"anthropic.claude-haiku-4-",
		"anthropic.claude-3-7-",
	}

	for _, model := range reasoningModels {
		if strings.Contains(modelID, model) {
			return true
		}
	}
	return false
}
