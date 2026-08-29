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

func (c anthropicTextGenerationInputContent) MarshalJSON() ([]byte, error) {
	type alias anthropicTextGenerationInputContent
	if c.Type != "thinking" || c.Thinking != "" {
		return json.Marshal(alias(c))
	}
	return json.Marshal(struct {
		Type      string `json:"type"`
		Thinking  string `json:"thinking"`
		Signature string `json:"signature,omitempty"`
	}{Type: c.Type, Signature: c.Signature})
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
	// ToolChoice constrains which tool the model may call. Optional
	ToolChoice any `json:"tool_choice,omitempty"`
	// Thinking configuration for reasoning models. Optional
	Thinking *anthropicThinkingPayload `json:"thinking,omitempty"`
	// OutputConfig carries the adaptive-thinking effort. It is a top-level
	// sibling of "thinking" in the Messages request body. Optional
	OutputConfig *anthropicOutputConfig `json:"output_config,omitempty"`
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
	// Display keeps adaptive reasoning text visible ("summarized"); Opus 4.7+
	// otherwise omits it by default.
	Display string `json:"display,omitempty"`
}

type anthropicOutputConfig struct {
	Effort string                 `json:"effort,omitempty"`
	Format *anthropicOutputFormat `json:"format,omitempty"`
}

// anthropicOutputFormat carries a structured-output JSON Schema for the legacy
// InvokeModel Anthropic payload, mirroring first-party output_config.format.
type anthropicOutputFormat struct {
	Type   string          `json:"type"` // "json_schema"
	Schema json.RawMessage `json:"schema,omitempty"`
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
	StopSequence string                `json:"stop_sequence"`
	StopDetails  *anthropicStopDetails `json:"stop_details,omitempty"`
	Usage        anthropicUsage        `json:"usage"`
}

type anthropicStopDetails struct {
	Category    string `json:"category,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type anthropicUsage struct {
	InputTokens              int32 `json:"input_tokens"`
	OutputTokens             int32 `json:"output_tokens"`
	CacheCreationInputTokens int32 `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int32 `json:"cache_read_input_tokens,omitempty"`
	CacheCreation            struct {
		Ephemeral5mInputTokens int32 `json:"ephemeral_5m_input_tokens,omitempty"`
		Ephemeral1hInputTokens int32 `json:"ephemeral_1h_input_tokens,omitempty"`
	} `json:"cache_creation,omitempty"`
	OutputTokensDetails struct {
		ThinkingTokens int32 `json:"thinking_tokens,omitempty"`
	} `json:"output_tokens_details,omitempty"`
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

	maxTokens := getMaxTokens(options.GetMaxTokens(), 2048)
	input := anthropicTextGenerationInput{
		AnthropicVersion: AnthropicLatestVersion,
		MaxTokens:        maxTokens,
		System:           system,
		Messages:         inputContents,
		Temperature:      options.GetTemperature(),
		TopP:             options.GetTopP(),
		TopK:             options.GetTopK(),
		StopSequences:    options.StopWords,
		Tools:            tools,
		ToolChoice:       anthropicToolChoiceOnWire(options.ToolChoice),
	}

	if err := applyAnthropicReasoning(&input, options.Reasoning, modelID, maxTokens); err != nil {
		return nil, err
	}
	if input.Thinking != nil {
		input.MaxTokens = reasoning.ClaudeMaxTokensForBudget(input.Thinking.BudgetTokens, input.MaxTokens)
	}

	// Structured output rides in output_config.format, merged with any effort set
	// by applyAnthropicReasoning (the two are independent).
	if err := applyAnthropicStructuredOutput(&input, modelID, options.StructuredOutput); err != nil {
		return nil, err
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
		streamResp, err := parseStreamingCompletionResponse(ctx, client, modelInput, options)
		if err != nil {
			return streamResp, err
		}
		// Validate the assembled stream too — a structured-output guarantee must
		// hold on both API paths.
		if err := validateStructuredResponse(options.StructuredOutput, modelID, streamResp); err != nil {
			return streamResp, err
		}
		return streamResp, nil
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
		Content:        textContent,
		Reasoning:      processReasoning(reasoningContent, signature),
		ToolCalls:      toolCalls,
		StopReason:     output.StopReason,
		Truncated:      llms.IsTruncated(output.StopReason),
		GenerationInfo: map[string]any{},
	}
	applyAnthropicUsage(choice.GenerationInfo, output.Usage)
	Contentchoices = append(Contentchoices, choice)

	contentResp := &llms.ContentResponse{
		Choices: Contentchoices,
	}
	if output.StopReason == "refusal" {
		refusal := &llms.ErrModelRefusal{
			Provider:                 "bedrock",
			Message:                  textContent,
			InputTokens:              int(output.Usage.InputTokens),
			OutputTokens:             int(output.Usage.OutputTokens),
			CacheCreationInputTokens: int(output.Usage.CacheCreationInputTokens),
			CacheReadInputTokens:     int(output.Usage.CacheReadInputTokens),
		}
		if output.StopDetails != nil {
			refusal.Category = output.StopDetails.Category
			refusal.Explanation = output.StopDetails.Explanation
		}
		return contentResp, refusal
	}
	if err := validateStructuredResponse(options.StructuredOutput, modelID, contentResp); err != nil {
		return contentResp, err
	}
	return contentResp, nil
}

func anthropicToolChoiceOnWire(choice any) any {
	switch kind, name := llms.ClassifyToolChoice(choice); kind {
	case llms.ToolChoiceNamed:
		return map[string]any{"type": "tool", "name": name}
	case llms.ToolChoiceAny:
		return map[string]any{"type": "any"}
	case llms.ToolChoiceAuto:
		return map[string]any{"type": "auto"}
	case llms.ToolChoiceNone:
		return map[string]any{"type": "none"}
	default:
		return nil
	}
}

func mergeAnthropicUsage(into anthropicUsage, updates ...anthropicUsage) anthropicUsage {
	for _, u := range updates {
		if u.InputTokens != 0 {
			into.InputTokens = u.InputTokens
		}
		if u.OutputTokens != 0 {
			into.OutputTokens = u.OutputTokens
		}
		if u.CacheCreationInputTokens != 0 {
			into.CacheCreationInputTokens = u.CacheCreationInputTokens
		}
		if u.CacheReadInputTokens != 0 {
			into.CacheReadInputTokens = u.CacheReadInputTokens
		}
		if u.CacheCreation.Ephemeral5mInputTokens != 0 {
			into.CacheCreation.Ephemeral5mInputTokens = u.CacheCreation.Ephemeral5mInputTokens
		}
		if u.CacheCreation.Ephemeral1hInputTokens != 0 {
			into.CacheCreation.Ephemeral1hInputTokens = u.CacheCreation.Ephemeral1hInputTokens
		}
		if u.OutputTokensDetails.ThinkingTokens != 0 {
			into.OutputTokensDetails.ThinkingTokens = u.OutputTokensDetails.ThinkingTokens
		}
	}
	return into
}

func applyAnthropicUsage(info map[string]any, usage anthropicUsage) {
	if info == nil {
		return
	}

	promptTokens := int(usage.InputTokens) +
		int(usage.CacheCreationInputTokens) + int(usage.CacheReadInputTokens)

	info["input_tokens"] = int(usage.InputTokens)
	info["output_tokens"] = int(usage.OutputTokens)
	info["PromptTokens"] = promptTokens
	info["CompletionTokens"] = int(usage.OutputTokens)
	info["TotalTokens"] = promptTokens + int(usage.OutputTokens)
	info["CacheReadInputTokens"] = int(usage.CacheReadInputTokens)
	info["CacheCreationInputTokens"] = int(usage.CacheCreationInputTokens)
	info["CacheCreationEphemeral5mInputTokens"] = int(usage.CacheCreation.Ephemeral5mInputTokens)
	info["CacheCreationEphemeral1hInputTokens"] = int(usage.CacheCreation.Ephemeral1hInputTokens)
	info["PromptCachedTokens"] = int(usage.CacheReadInputTokens)
	info["ReasoningTokens"] = int(usage.OutputTokensDetails.ThinkingTokens)
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
	Usage   anthropicUsage         `json:"usage"`
	Message anthropicStreamMessage `json:"message"`
}

type anthropicStreamMessage struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []any          `json:"content"`
	Model        string         `json:"model"`
	StopReason   any            `json:"stop_reason"`
	StopSequence any            `json:"stop_sequence"`
	Usage        anthropicUsage `json:"usage"`
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
	var usage anthropicUsage
	var streamedContent strings.Builder
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
				usage = mergeAnthropicUsage(usage, resp.Message.Usage, resp.Usage)
				applyAnthropicUsage(contentchoices[0].GenerationInfo, usage)
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
					streamedContent.WriteString(resp.Delta.Text)
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
				contentchoices[0].Truncated = llms.IsTruncated(resp.Delta.StopReason)
				usage = mergeAnthropicUsage(usage, resp.Message.Usage, resp.Usage)
				applyAnthropicUsage(contentchoices[0].GenerationInfo, usage)
			}
		}
	}
	if err = stream.Err(); err != nil {
		return nil, err
	}

	// Add tool calls to the final response
	contentchoices[0].ToolCalls = toolCalls

	// Add signature to reasoning if accumulated
	if signature.Len() > 0 {
		if contentchoices[0].Reasoning == nil {
			contentchoices[0].Reasoning = &reasoning.ContentReasoning{}
		}
		contentchoices[0].Reasoning.Signature = []byte(signature.String())
	}

	contentchoices[0].Content = streamedContent.String()

	response := &llms.ContentResponse{Choices: contentchoices}
	if contentchoices[0].StopReason == "refusal" {
		return response, &llms.ErrModelRefusal{
			Provider:                 "bedrock",
			Message:                  streamedContent.String(),
			InputTokens:              int(usage.InputTokens),
			OutputTokens:             int(usage.OutputTokens),
			CacheCreationInputTokens: int(usage.CacheCreationInputTokens),
			CacheReadInputTokens:     int(usage.CacheReadInputTokens),
		}
	}
	return response, nil
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

// applyAnthropicReasoning resolves the thinking mechanism from the model via the
// reasoning capability resolver: an adaptive-only generation gets adaptive
// thinking and drops sampling params, a budget-only generation gets budget
// thinking, and the caller's preference is honored where the model supports both.
func applyAnthropicReasoning(
	input *anthropicTextGenerationInput, cfg *llms.ReasoningConfig, modelID string, maxTokens int,
) error {
	// Adaptive-only models reject sampling params even without thinking.
	if reasoning.ClaudeRejectsSampling(modelID) {
		input.Temperature = 0
		input.TopP = 0
		input.TopK = 0
	}
	if reasoning.ClaudeMutuallyExclusiveSampling(modelID) && input.Temperature != 0 && input.TopP != 0 {
		input.TopP = 0
	}

	switch cfg.ResolveMode() {
	case llms.ReasoningDefault:
		return nil
	case llms.ReasoningOff:
		switch reasoning.ResolveOff(modelID, reasoning.ProviderBedrock) {
		case reasoning.OffDisableClaude:
			input.Thinking = &anthropicThinkingPayload{Type: "disabled"}
		case reasoning.OffUnsupported:
			return &reasoning.ErrReasoningOffUnsupported{Model: modelID}
		}
		return nil
	}

	setAdaptive := func() {
		input.Thinking = &anthropicThinkingPayload{Type: "adaptive", Display: "summarized"}
		input.OutputConfig = &anthropicOutputConfig{Effort: reasoning.ClaudeClampEffort(modelID, string(cfg.GetEffort(maxTokens)))}
		input.Temperature = 0
		input.TopP = 0
		input.TopK = 0
	}
	setBudget := func() error {
		tokens := reasoning.ClaudeClampBudget(modelID, cfg.GetTokens(maxTokens))
		if tokens <= 0 {
			return &reasoning.ErrEffortHasNoBudget{Model: modelID, Effort: string(cfg.GetEffort(maxTokens))}
		}
		input.Thinking = &anthropicThinkingPayload{Type: "enabled", BudgetTokens: tokens}
		if reasoning.ClaudeSupportsEffortWithBudget(modelID, reasoning.ProviderBedrock) {
			input.OutputConfig = &anthropicOutputConfig{Effort: reasoning.ClaudeClampEffort(modelID, string(cfg.GetEffort(maxTokens)))}
		}
		// Budget thinking requires temperature=1.0 and rejects top_p/top_k.
		input.Temperature = 1.0
		if reasoning.ClaudeRejectsSampling(modelID) {
			input.Temperature = 0
		}
		input.TopP = 0
		input.TopK = 0
		return nil
	}

	switch reasoning.ResolveMechanism(modelID, cfg.Adaptive, true, supportsAnthropicReasoning(modelID)) {
	case reasoning.MechanismAdaptive:
		setAdaptive()
	case reasoning.MechanismBudget:
		return setBudget()
	}
	return nil
}

func supportsAnthropicReasoning(modelID string) bool {
	return reasoning.IsReasoningModel(modelID)
}
