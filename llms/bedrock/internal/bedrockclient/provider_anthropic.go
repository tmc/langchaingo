package bedrockclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/tmc/langchaingo/llms"
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
	// One of: "text", "image", "tool_result", "tool_use"
	Type string `json:"type"`
	// The source of the content. Required if type is "image"
	Source *anthropicBinGenerationInputSource `json:"source,omitempty"`
	// The text content. Required if type is "text"
	Text string `json:"text,omitempty"`
	// Tool result fields
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	// Tool use fields (for tool calls from AI)
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
	// CacheControl marks this content block as a prompt-cache breakpoint.
	// The cached prefix extends from the start of the prompt up to and
	// including this block.
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

// anthropicCacheControl maps llms.CacheControl onto Bedrock's wire format for
// Anthropic models.
type anthropicCacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

// anthropicSystemBlock is used when the system field is emitted as an array of
// blocks (only when at least one system part carries cache control).
type anthropicSystemBlock struct {
	Type         string                 `json:"type"`
	Text         string                 `json:"text"`
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
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
	// The system prompt to use. Optional.
	// Either a string (legacy form, no caching) or []anthropicSystemBlock
	// when at least one system part carries cache control.
	System interface{} `json:"system,omitempty"`
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
	// Tools available to the model. Optional
	Tools []BedrockTool `json:"tools,omitempty"`
	// Tool choice configuration. Optional
	ToolChoice *BedrockToolChoice `json:"tool_choice,omitempty"`
	// Guardrail configuration. Optional
	GuardrailConfig *struct {
		TagSuffix            string `json:"tagSuffix,omitempty"`
		StreamProcessingMode string `json:"streamProcessingMode,omitempty"`
	} `json:"amazon-bedrock-guardrailConfig,omitempty"`
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
	// Can be "text" or "tool_use"
	Content []anthropicContentBlock `json:"content"`
	// The reason for the completion of the generation.
	// One of: ["end_turn", "max_tokens", "stop_sequence", "tool_use"]
	StopReason string `json:"stop_reason"`
	// Which custom stop sequence was matched, if any.
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	} `json:"usage"`
}

// anthropicContentBlock represents a content block in Anthropic response
type anthropicContentBlock struct {
	Type string `json:"type"` // "text" or "tool_use"
	Text string `json:"text,omitempty"`
	// Tool use fields
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
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

	input := anthropicTextGenerationInput{
		AnthropicVersion: AnthropicLatestVersion,
		MaxTokens:        getMaxTokens(options.MaxTokens, 2048),
		System:           systemPrompt,
		Messages:         inputContents,
		Temperature:      options.Temperature,
		TopP:             options.TopP,
		TopK:             options.TopK,
		StopSequences:    options.StopWords,
	}

	// Add tools if provided
	if len(options.Tools) > 0 {
		bedrockTools, err := convertToolsToBedrockTools(options.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		input.Tools = bedrockTools

		// Add tool choice if provided
		if options.ToolChoice != nil {
			toolChoice, err := convertToolChoiceToBedrockToolChoice(options.ToolChoice)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool choice: %w", err)
			}
			input.ToolChoice = toolChoice
		}
	}

	// Add guardrail tag suffix and stream processing mode if provided in safety config
	if options.SafetyConfig != nil {
		input.GuardrailConfig = &struct {
			TagSuffix            string `json:"tagSuffix,omitempty"`
			StreamProcessingMode string `json:"streamProcessingMode,omitempty"`
		}{}
		if tagSuffix, ok := options.SafetyConfig["tag_suffix"]; ok {
			if str, ok := tagSuffix.(string); ok {
				input.GuardrailConfig.TagSuffix = str
			} else {
				return nil, errors.New("tag_suffix in safety config must be a string")
			}
		}
		if streamProcessingMode, ok := options.SafetyConfig["stream_processing_mode"]; ok {
			if str, ok := streamProcessingMode.(string); ok {
				input.GuardrailConfig.StreamProcessingMode = str
			} else {
				return nil, errors.New("stream_processing_mode in safety config must be a string")
			}
		}
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if options.StreamingFunc != nil {
		return parseStreamingCompletionResponse(ctx, client, modelID, body, options)
	}

	resp, err := invokeModel(ctx, client, modelID, body, options)
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
	} else if stopReason := output.StopReason; stopReason != AnthropicCompletionReasonEndTurn && stopReason != AnthropicCompletionReasonStopSequence && stopReason != "tool_use" {
		return nil, errors.New("completed due to " + stopReason + ". Maybe try increasing max tokens")
	}

	// Process content blocks and build a single ContentChoice
	choice := &llms.ContentChoice{
		StopReason: output.StopReason,
		GenerationInfo: map[string]interface{}{
			"input_tokens":             output.Usage.InputTokens,
			"output_tokens":            output.Usage.OutputTokens,
			"CacheCreationInputTokens": output.Usage.CacheCreationInputTokens,
			"CacheReadInputTokens":     output.Usage.CacheReadInputTokens,
		},
	}

	var textContent string
	var toolCalls []llms.ToolCall

	for _, block := range output.Content {
		switch block.Type {
		case "text":
			textContent += block.Text
		case "tool_use":
			toolCall, err := convertBedrockToolCallToLLMToolCall(BedrockToolCall{
				Type:  block.Type,
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool call: %w", err)
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	choice.Content = textContent
	choice.ToolCalls = toolCalls

	// Set legacy FuncCall field for backward compatibility
	if len(toolCalls) > 0 {
		choice.FuncCall = toolCalls[0].FunctionCall
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{choice},
	}, nil
}

func parseStreamingCompletionResponse(ctx context.Context, client invokeModelWithResponseStreamAPI, modelID string, body []byte, options llms.CallOptions) (*llms.ContentResponse, error) {
	output, err := invokeModelWithResponseStream(ctx, client, modelID, body, options)
	if err != nil {
		return nil, err
	}
	stream := output.GetStream()
	if stream == nil {
		return nil, errors.New("no stream")
	}
	defer stream.Close()

	contentchoices := []*llms.ContentChoice{{GenerationInfo: map[string]interface{}{}}}
	for e := range stream.Events() {
		if err = stream.Err(); err != nil {
			return nil, err
		}

		if v, ok := e.(*types.ResponseStreamMemberChunk); ok {
			var resp streamingCompletionResponseChunk
			if err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp); err != nil {
				return nil, err
			}
			if err := checkGuardrailError(&resp); err != nil {
				return nil, err
			}

			switch resp.Type {
			case "message_start":
				contentchoices[0].GenerationInfo["input_tokens"] = resp.Message.Usage.InputTokens
				contentchoices[0].GenerationInfo["CacheCreationInputTokens"] = resp.Message.Usage.CacheCreationInputTokens
				contentchoices[0].GenerationInfo["CacheReadInputTokens"] = resp.Message.Usage.CacheReadInputTokens
			case "content_block_delta":
				if err = options.StreamingFunc(ctx, []byte(resp.Delta.Text)); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.Delta.Text
			case "message_delta":
				contentchoices[0].StopReason = resp.Delta.StopReason
				contentchoices[0].GenerationInfo["output_tokens"] = resp.Usage.OutputTokens
				if resp.Usage.CacheCreationInputTokens > 0 {
					contentchoices[0].GenerationInfo["CacheCreationInputTokens"] = resp.Usage.CacheCreationInputTokens
				}
				if resp.Usage.CacheReadInputTokens > 0 {
					contentchoices[0].GenerationInfo["CacheReadInputTokens"] = resp.Usage.CacheReadInputTokens
				}
			}
		}
	}
	if err = stream.Err(); err != nil {
		return nil, err
	}

	return &llms.ContentResponse{
		Choices: contentchoices,
	}, nil
}

// converts the flattened Bedrock messages into
// Anthropic's request shape. The returned system value is one of:
//   - nil: no system messages were provided.
//   - string: concatenated system text (legacy form, backward-compatible when
//     no system part carries cache control).
//   - []anthropicSystemBlock: emitted only when at least one system part has
//     a cache control, so that per-block cache_control markers are preserved.
func processInputMessagesAnthropic(messages []Message) ([]*anthropicTextGenerationInputMessage, interface{}, error) {
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
	var systemParts []Message
	for _, chunk := range chunkedMessages {
		role, err := getAnthropicRole(chunk[0].Role)
		if err != nil {
			return nil, nil, err
		}
		if role == AnthropicSystem {
			if len(systemParts) > 0 {
				return nil, nil, errors.New("multiple system prompts")
			}
			systemParts = append(systemParts, chunk...)
			continue
		}
		content := make([]anthropicTextGenerationInputContent, 0, len(chunk))
		for _, message := range chunk {
			content = append(content, getAnthropicInputContent(message))
		}
		inputContents = append(inputContents, &anthropicTextGenerationInputMessage{
			Role:    role,
			Content: content,
		})
	}

	system, err := buildAnthropicSystem(systemParts)
	if err != nil {
		return nil, nil, err
	}
	return inputContents, system, nil
}

// buildAnthropicSystem emits the system field in array form when any part
// carries cache control, otherwise concatenates into a single string for
// backward compatibility.
func buildAnthropicSystem(parts []Message) (interface{}, error) {
	if len(parts) == 0 {
		return nil, nil
	}
	for _, m := range parts {
		if m.Type != AnthropicMessageTypeText {
			return nil, errors.New("system prompt must be text")
		}
	}
	hasCache := false
	for _, m := range parts {
		if m.CacheControl != nil {
			hasCache = true
			break
		}
	}
	if !hasCache {
		var s string
		for _, m := range parts {
			s += m.Content
		}
		return s, nil
	}
	blocks := make([]anthropicSystemBlock, 0, len(parts))
	for _, m := range parts {
		blocks = append(blocks, anthropicSystemBlock{
			Type:         AnthropicMessageTypeText,
			Text:         m.Content,
			CacheControl: cacheControlToAnthropic(m.CacheControl),
		})
	}
	return blocks, nil
}

// cacheControlToAnthropic converts the provider-agnostic llms.CacheControl
// into Anthropic's Bedrock wire format.
func cacheControlToAnthropic(cc *llms.CacheControl) *anthropicCacheControl {
	if cc == nil {
		return nil
	}
	t := cc.Type
	if t == "" {
		t = "ephemeral"
	}
	out := &anthropicCacheControl{Type: t}
	if cc.Duration > 0 {
		out.TTL = "1h"
	}
	return out
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
	case llms.ChatMessageTypeFunction, llms.ChatMessageTypeTool:
		return AnthropicRoleUser, nil // Tool results are sent as user messages
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
	case AnthropicMessageTypeImage:
		c = anthropicTextGenerationInputContent{
			Type: message.Type,
			Source: &anthropicBinGenerationInputSource{
				Type:      "base64",
				MediaType: message.MimeType,
				Data:      base64.StdEncoding.EncodeToString([]byte(message.Content)),
			},
		}
	case "tool_result":
		// Handle tool results from tool response messages
		c = anthropicTextGenerationInputContent{
			Type:      "tool_result",
			ToolUseID: message.ToolUseID,
			Content:   message.Content,
		}
	case "tool_call":
		// Handle tool calls from AI messages - convert to tool_use format for Anthropic
		var input interface{}
		if message.ToolArgs != "" {
			// Try to parse the arguments as JSON
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(message.ToolArgs), &args); err == nil {
				input = args
			} else {
				// If parsing fails, wrap in a simple structure
				input = map[string]interface{}{"arguments": message.ToolArgs}
			}
		} else {
			input = map[string]interface{}{}
		}

		c = anthropicTextGenerationInputContent{
			Type:  "tool_use",
			ID:    message.ToolCallID,
			Name:  message.ToolName,
			Input: input,
		}
	}
	c.CacheControl = cacheControlToAnthropic(message.CacheControl)
	return c
}
