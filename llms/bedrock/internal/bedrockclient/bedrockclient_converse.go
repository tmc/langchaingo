package bedrockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// ConverseClient wraps the Bedrock Converse API client
type ConverseClient struct {
	client BedrockRuntimeClientInterface
}

// BedrockRuntimeClientInterface defines the interface for bedrock runtime operations
type BedrockRuntimeClientInterface interface {
	Converse(ctx context.Context, input *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
	ConverseStream(ctx context.Context, input *bedrockruntime.ConverseStreamInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseStreamOutput, error)
}

// NewConverseClient creates a new Converse API client
func NewConverseClient(client BedrockRuntimeClientInterface) *ConverseClient {
	return &ConverseClient{
		client: client,
	}
}

// CreateCompletionConverse creates a completion using the Converse API
func (c *ConverseClient) CreateCompletionConverse(ctx context.Context, input *ConverseInput) (*llms.ContentResponse, error) {
	converseInput, err := c.buildConverseInput(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build converse input: %w", err)
	}

	if input.StreamingFunc != nil {
		return c.handleStreamingResponse(ctx, converseInput, input.StreamingFunc)
	}

	return c.handleNonStreamingResponse(ctx, converseInput)
}

// ConverseInput represents input for the Converse API
type ConverseInput struct {
	Messages        []Message
	ModelID         string
	MaxTokens       *int
	Temperature     *float64
	TopP            *float64
	StopSequences   []string
	Tools           []llms.Tool
	StreamingFunc   streaming.Callback
	ReasoningConfig *llms.ReasoningConfig
	EnableCaching   bool
}

type converseThinkingPayload struct {
	Type         string `json:"type" document:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty" document:"budget_tokens,omitempty"`
}

type converseAdditionalModelRequestFields struct {
	Thinking *converseThinkingPayload `json:"thinking,omitempty" document:"thinking,omitempty"`
}

// buildConverseInput converts our input to AWS Converse format
func (c *ConverseClient) buildConverseInput(input *ConverseInput) (*bedrockruntime.ConverseInput, error) {
	// Convert messages
	converseMessages, systemPrompts, err := c.convertMessages(input.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Build inference configuration
	inferenceConfig := &types.InferenceConfiguration{}
	if input.MaxTokens != nil {
		inferenceConfig.MaxTokens = aws.Int32(int32(*input.MaxTokens))
	}
	if input.Temperature != nil {
		inferenceConfig.Temperature = aws.Float32(float32(*input.Temperature))
	}
	if input.TopP != nil {
		inferenceConfig.TopP = aws.Float32(float32(*input.TopP))
	}
	if len(input.StopSequences) > 0 {
		inferenceConfig.StopSequences = input.StopSequences
	}

	// Add cachePoint to messages if caching is enabled
	if input.EnableCaching && len(converseMessages) > 0 {
		c.addCachePointToMessages(converseMessages)
	}

	converseInput := &bedrockruntime.ConverseInput{
		ModelId:         aws.String(input.ModelID),
		Messages:        converseMessages,
		InferenceConfig: inferenceConfig,
	}

	// Add system prompts if any
	if len(systemPrompts) > 0 {
		// Add cachePoint after system if caching is enabled
		if input.EnableCaching {
			systemPrompts = append(systemPrompts, &types.SystemContentBlockMemberCachePoint{
				Value: types.CachePointBlock{
					Type: types.CachePointTypeDefault,
					Ttl:  types.CacheTTLFiveMinutes,
				},
			})
		}
		converseInput.System = systemPrompts
	}

	// Add tool configuration if tools are provided
	if len(input.Tools) > 0 {
		toolConfig, err := c.convertToolsToToolConfig(input.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		converseInput.ToolConfig = toolConfig
	}

	// Add additional model fields
	if input.ReasoningConfig != nil && c.supportsReasoning(input.ModelID) {
		additionalModelFields := converseAdditionalModelRequestFields{}
		maxTokens := 0 // Use 0 to let it use default maxTokens
		if input.MaxTokens != nil {
			maxTokens = *input.MaxTokens
		}
		tokens := input.ReasoningConfig.GetTokens(maxTokens)
		if tokens > 0 {
			additionalModelFields.Thinking = &converseThinkingPayload{
				Type:         "enabled",
				BudgetTokens: tokens,
			}
		}
		if additionalModelFields.Thinking != nil {
			converseInput.AdditionalModelRequestFields = document.NewLazyDocument(additionalModelFields)
		}
	}

	return converseInput, nil
}

// convertMessages converts our messages to Converse format
func (c *ConverseClient) convertMessages(messages []Message) ([]types.Message, []types.SystemContentBlock, error) {
	var converseMessages []types.Message
	var systemPrompts []types.SystemContentBlock

	for _, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			systemPrompts = append(systemPrompts, &types.SystemContentBlockMemberText{
				Value: msg.Content,
			})
		case llms.ChatMessageTypeHuman, llms.ChatMessageTypeAI:
			converseMsg, err := c.convertUserOrAssistantMessage(msg)
			if err != nil {
				return nil, nil, err
			}
			// Add cachePoint if message has cache control
			if msg.CacheControl != nil {
				cacheBlock := types.CachePointBlock{
					Type: types.CachePointTypeDefault,
				}
				if msg.CacheControl.TTL == "1h" {
					cacheBlock.Ttl = types.CacheTTLOneHour
				} else {
					cacheBlock.Ttl = types.CacheTTLFiveMinutes
				}
				converseMsg.Content = append(converseMsg.Content, &types.ContentBlockMemberCachePoint{
					Value: cacheBlock,
				})
			}
			converseMessages = append(converseMessages, converseMsg)
		case llms.ChatMessageTypeTool:
			converseMsg, err := c.convertToolMessage(msg)
			if err != nil {
				return nil, nil, err
			}
			converseMessages = append(converseMessages, converseMsg)
		}
	}

	return converseMessages, systemPrompts, nil
}

// addCachePointToMessages adds cachePoint to the last message for conversation history caching
func (c *ConverseClient) addCachePointToMessages(messages []types.Message) {
	if len(messages) == 0 {
		return
	}

	// Add cachePoint to the last message's content
	lastMsg := &messages[len(messages)-1]
	lastMsg.Content = append(lastMsg.Content, &types.ContentBlockMemberCachePoint{
		Value: types.CachePointBlock{
			Type: types.CachePointTypeDefault,
			Ttl:  types.CacheTTLFiveMinutes,
		},
	})
}

// convertUserOrAssistantMessage converts user or assistant messages
func (c *ConverseClient) convertUserOrAssistantMessage(msg Message) (types.Message, error) {
	var role types.ConversationRole
	if msg.Role == llms.ChatMessageTypeHuman {
		role = types.ConversationRoleUser
	} else {
		role = types.ConversationRoleAssistant
	}

	var contentBlocks []types.ContentBlock

	// For AI messages with reasoning, add reasoning blocks first
	if msg.Role == llms.ChatMessageTypeAI && msg.Reasoning != nil {
		// Add thinking block if present
		if msg.Reasoning.Content != "" || len(msg.Reasoning.Signature) > 0 {
			reasoningBlock := types.ReasoningContentBlockMemberReasoningText{
				Value: types.ReasoningTextBlock{
					Text:      ptrStringOrNil(msg.Reasoning.Content),
					Signature: ptrStringOrNil(string(msg.Reasoning.Signature)),
				},
			}
			contentBlocks = append(contentBlocks, &types.ContentBlockMemberReasoningContent{
				Value: &reasoningBlock,
			})
		}
		// Add redacted thinking block if present
		if len(msg.Reasoning.RedactedContent) > 0 {
			redactedBlock := types.ReasoningContentBlockMemberRedactedContent{
				Value: msg.Reasoning.RedactedContent,
			}
			contentBlocks = append(contentBlocks, &types.ContentBlockMemberReasoningContent{
				Value: &redactedBlock,
			})
		}
	}

	// Handle text content
	if msg.Content != "" {
		contentBlocks = append(contentBlocks, &types.ContentBlockMemberText{
			Value: msg.Content,
		})
	}

	// Handle tool calls
	if msg.ToolCall != nil {
		contentBlocks = append(contentBlocks, &types.ContentBlockMemberToolUse{
			Value: types.ToolUseBlock{
				ToolUseId: aws.String(msg.ToolCall.ID),
				Name:      aws.String(msg.ToolCall.Name),
				Input:     document.NewLazyDocument(msg.ToolCall.Arguments),
			},
		})
	}

	return types.Message{
		Role:    role,
		Content: contentBlocks,
	}, nil
}

func (c *ConverseClient) convertToolCallInput(args any) (any, error) {
	if isSmithyValidObject(args) {
		return args, nil
	}

	// Convert to Smithy-compatible format by re-encoding through JSON
	// This handles types like map[string]any with interface{} values
	jsonBytes := bytes.NewBuffer(nil)
	if err := json.NewEncoder(jsonBytes).Encode(args); err != nil {
		return nil, fmt.Errorf("failed to encode arguments: %w", err)
	}

	jsonDecoder := json.NewDecoder(jsonBytes)
	jsonDecoder.UseNumber()

	var jsonValue any
	if err := jsonDecoder.Decode(&jsonValue); err != nil {
		return nil, fmt.Errorf("failed to decode arguments: %w", err)
	}

	return jsonValue, nil
}

// convertToolMessage converts tool result messages
func (c *ConverseClient) convertToolMessage(msg Message) (types.Message, error) {
	if msg.ToolResult == nil {
		return types.Message{}, fmt.Errorf("tool message missing tool result")
	}

	contentBlocks := []types.ContentBlock{
		&types.ContentBlockMemberToolResult{
			Value: types.ToolResultBlock{
				ToolUseId: aws.String(msg.ToolResult.ToolCallID),
				Content: []types.ToolResultContentBlock{
					&types.ToolResultContentBlockMemberText{
						Value: msg.ToolResult.Content,
					},
				},
			},
		},
	}

	return types.Message{
		Role:    types.ConversationRoleUser,
		Content: contentBlocks,
	}, nil
}

// convertToolsToToolConfig converts llms.Tool to Converse ToolConfiguration
func (c *ConverseClient) convertToolsToToolConfig(tools []llms.Tool) (*types.ToolConfiguration, error) {
	var converseTools []types.Tool

	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}

		toolSpec := types.ToolSpecification{
			Name:        aws.String(tool.Function.Name),
			Description: aws.String(tool.Function.Description),
		}

		// Convert function parameters to tool input schema
		if tool.Function.Parameters != nil {
			parameters, err := c.convertToolCallInput(tool.Function.Parameters)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool call input: %w", err)
			}
			toolSpec.InputSchema = &types.ToolInputSchemaMemberJson{
				Value: document.NewLazyDocument(parameters),
			}
		}

		converseTools = append(converseTools, &types.ToolMemberToolSpec{
			Value: toolSpec,
		})
	}

	return &types.ToolConfiguration{
		Tools:      converseTools,
		ToolChoice: &types.ToolChoiceMemberAuto{},
	}, nil
}

// handleNonStreamingResponse handles non-streaming responses
func (c *ConverseClient) handleNonStreamingResponse(ctx context.Context, input *bedrockruntime.ConverseInput) (*llms.ContentResponse, error) {
	response, err := c.client.Converse(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("converse API call failed: %w", err)
	}

	return c.convertConverseResponse(response)
}

// handleStreamingResponse handles streaming responses
func (c *ConverseClient) handleStreamingResponse(ctx context.Context, input *bedrockruntime.ConverseInput, callback streaming.Callback) (*llms.ContentResponse, error) {
	streamInput := &bedrockruntime.ConverseStreamInput{
		ModelId:                      input.ModelId,
		Messages:                     input.Messages,
		System:                       input.System,
		InferenceConfig:              input.InferenceConfig,
		ToolConfig:                   input.ToolConfig,
		AdditionalModelRequestFields: input.AdditionalModelRequestFields,
	}

	response, err := c.client.ConverseStream(ctx, streamInput)
	if err != nil {
		return nil, fmt.Errorf("converse stream API call failed: %w", err)
	}

	return c.processStreamingResponse(ctx, response, callback)
}

// processStreamingResponse processes streaming events
func (c *ConverseClient) processStreamingResponse(ctx context.Context, response *bedrockruntime.ConverseStreamOutput, callback streaming.Callback) (*llms.ContentResponse, error) {
	var fullContent strings.Builder
	var reasoningContent strings.Builder
	var signature bytes.Buffer
	var redactedContent bytes.Buffer
	var toolCalls []llms.ToolCall
	currentToolCalls := make(map[string]*streaming.ToolCall) // Track streaming tool calls by ID

	defer streaming.CallWithDone(ctx, callback)

	stream := response.GetStream()
	defer stream.Close()

	for event := range stream.Events() {
		switch e := event.(type) {
		case *types.ConverseStreamOutputMemberContentBlockDelta:
			if e.Value.Delta != nil {
				switch delta := e.Value.Delta.(type) {
				case *types.ContentBlockDeltaMemberText:
					fullContent.WriteString(delta.Value)
					if callback != nil {
						chunk := streaming.Chunk{
							Type:    streaming.ChunkTypeText,
							Content: delta.Value,
						}
						if err := callback(ctx, chunk); err != nil {
							return nil, err
						}
					}
				case *types.ContentBlockDeltaMemberReasoningContent:
					if callback != nil {
						switch block := delta.Value.(type) {
						case *types.ReasoningContentBlockDeltaMemberText:
							reasoningContent.WriteString(block.Value)
							chunk := streaming.Chunk{
								Type:      streaming.ChunkTypeReasoning,
								Reasoning: &reasoning.ContentReasoning{Content: block.Value},
							}
							if err := callback(ctx, chunk); err != nil {
								return nil, err
							}
						case *types.ReasoningContentBlockDeltaMemberRedactedContent:
							if len(block.Value) > 0 {
								redactedContent.Write(block.Value)
							}
						case *types.ReasoningContentBlockDeltaMemberSignature:
							if len(block.Value) > 0 {
								signature.WriteString(block.Value)
							}
						}
					}
				case *types.ContentBlockDeltaMemberToolUse:
					// Handle tool use delta (accumulate input arguments)
					if delta.Value.Input != nil {
						// delta.Value.Input is already a partial JSON string
						inputStr := *delta.Value.Input

						// Find the active tool call and accumulate arguments
						for _, toolCall := range currentToolCalls {
							toolCall.Arguments += inputStr
							break // Only one active tool call at a time typically
						}
					}
				}
			}
		case *types.ConverseStreamOutputMemberContentBlockStart:
			if e.Value.Start != nil {
				if toolUse, ok := e.Value.Start.(*types.ContentBlockStartMemberToolUse); ok {
					// Create streaming tool call, arguments will come through delta events
					toolCall := &streaming.ToolCall{
						ID:        *toolUse.Value.ToolUseId,
						Name:      *toolUse.Value.Name,
						Arguments: "",
					}

					currentToolCalls[*toolUse.Value.ToolUseId] = toolCall

					// Don't send chunk here, will send complete one in ContentBlockStop
				}
			}
		case *types.ConverseStreamOutputMemberContentBlockStop:
			// Handle tool call completion
			for _, toolCall := range currentToolCalls {
				// Send complete tool call through streaming
				if callback != nil {
					streamChunk := streaming.Chunk{
						Type:     streaming.ChunkTypeToolCall,
						ToolCall: *toolCall,
					}
					if err := callback(ctx, streamChunk); err != nil {
						return nil, err
					}
				}

				// Convert streaming tool call to final llms.ToolCall
				finalToolCall := llms.ToolCall{
					ID:   toolCall.ID,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
					},
				}
				toolCalls = append(toolCalls, finalToolCall)
			}
			// Clear current tool calls after completion
			currentToolCalls = make(map[string]*streaming.ToolCall)
		case *types.ConverseStreamOutputMemberMessageStop:
			// Stream completed - ensure any remaining tool calls are added
			for _, toolCall := range currentToolCalls {
				finalToolCall := llms.ToolCall{
					ID:   toolCall.ID,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
					},
				}
				toolCalls = append(toolCalls, finalToolCall)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	var sig []byte
	if signature.Len() > 0 {
		sig = signature.Bytes()
	}
	var redacted []byte
	if redactedContent.Len() > 0 {
		redacted = redactedContent.Bytes()
	}

	choice := &llms.ContentChoice{
		Content:        fullContent.String(),
		ToolCalls:      toolCalls,
		GenerationInfo: make(map[string]any),
		Reasoning:      c.processReasoning(reasoningContent.String(), sig, redacted),
	}

	result := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{choice},
	}

	return result, nil
}

func (c *ConverseClient) processReasoning(reasoningContent string, signature []byte, redactedContent []byte) *reasoning.ContentReasoning {
	if reasoningContent == "" && len(signature) == 0 && len(redactedContent) == 0 {
		return nil
	}

	return &reasoning.ContentReasoning{
		Content:         reasoningContent,
		Signature:       signature,
		RedactedContent: redactedContent,
	}
}

// convertConverseResponse converts Converse response to ContentResponse
func (c *ConverseClient) convertConverseResponse(response *bedrockruntime.ConverseOutput) (*llms.ContentResponse, error) {
	if response.Output == nil {
		return &llms.ContentResponse{}, nil
	}

	choice := &llms.ContentChoice{
		GenerationInfo: make(map[string]any),
	}

	// Handle different output types
	switch output := response.Output.(type) {
	case *types.ConverseOutputMemberMessage:
		// Process message content
		for _, contentBlock := range output.Value.Content {
			switch block := contentBlock.(type) {
			case *types.ContentBlockMemberText:
				choice.Content += block.Value
			case *types.ContentBlockMemberToolUse:
				// Convert tool use to ToolCall
				// Extract input from document.LazyDocument
				var inputData any
				if block.Value.Input != nil {
					if err := block.Value.Input.UnmarshalSmithyDocument(&inputData); err != nil {
						return nil, fmt.Errorf("failed to unmarshal tool input document: %w", err)
					}
				}
				argsJSON, err := json.Marshal(inputData)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool input to JSON: %w", err)
				}
				toolCall := llms.ToolCall{
					ID:   *block.Value.ToolUseId,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      *block.Value.Name,
						Arguments: string(argsJSON),
					},
				}
				choice.ToolCalls = append(choice.ToolCalls, toolCall)
			case *types.ContentBlockMemberReasoningContent:
				// The block.Value is of type ReasoningContentBlock
				switch content := block.Value.(type) {
				case *types.ReasoningContentBlockMemberReasoningText:
					reasoningText := ""
					if content.Value.Text != nil {
						reasoningText = *content.Value.Text
					}
					var sig []byte
					if content.Value.Signature != nil {
						sig = []byte(*content.Value.Signature)
					}
					choice.Reasoning = c.processReasoning(reasoningText, sig, nil)
				case *types.ReasoningContentBlockMemberRedactedContent:
					var redacted []byte
					if len(content.Value) > 0 {
						redacted = content.Value
					}
					if choice.Reasoning == nil {
						choice.Reasoning = c.processReasoning("", nil, redacted)
					} else {
						choice.Reasoning.RedactedContent = redacted
					}
				}
			}
		}

		// Note: types.Message doesn't have StopReason field
		// StopReason might be available in a different part of the response
	}

	// Add usage information
	if response.Usage != nil {
		if response.Usage.InputTokens != nil {
			choice.GenerationInfo["input_tokens"] = *response.Usage.InputTokens
			choice.GenerationInfo["PromptTokens"] = *response.Usage.InputTokens
		}
		if response.Usage.OutputTokens != nil {
			choice.GenerationInfo["output_tokens"] = *response.Usage.OutputTokens
			choice.GenerationInfo["CompletionTokens"] = *response.Usage.OutputTokens
		}
		if response.Usage.TotalTokens != nil {
			choice.GenerationInfo["total_tokens"] = *response.Usage.TotalTokens
			choice.GenerationInfo["TotalTokens"] = *response.Usage.TotalTokens
		}
		// Add cache metrics if available
		if response.Usage.CacheReadInputTokens != nil {
			choice.GenerationInfo["cacheReadInputTokens"] = *response.Usage.CacheReadInputTokens
			choice.GenerationInfo["CacheReadInputTokens"] = *response.Usage.CacheReadInputTokens
			choice.GenerationInfo["PromptCachedTokens"] = *response.Usage.CacheReadInputTokens
		}
		if response.Usage.CacheWriteInputTokens != nil {
			choice.GenerationInfo["cacheWriteInputTokens"] = *response.Usage.CacheWriteInputTokens
			choice.GenerationInfo["CacheCreationInputTokens"] = *response.Usage.CacheWriteInputTokens
		}
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{choice},
	}, nil
}

// Helper methods for model detection

// supportsReasoning checks if the model supports reasoning
func (c *ConverseClient) supportsReasoning(modelID string) bool {
	// Only Claude 3.7+ models support reasoning
	reasoningModels := []string{
		"anthropic.claude-opus-4-",
		"anthropic.claude-sonnet-4-",
		"anthropic.claude-haiku-4-",
		"anthropic.claude-3-7-",
		"openai.gpt-oss-120b",
		"openai.gpt-oss-20b",
		"moonshot.kimi-k2-thinking",
	}

	for _, model := range reasoningModels {
		if strings.Contains(modelID, model) {
			return true
		}
	}
	return false
}

func ptrStringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
