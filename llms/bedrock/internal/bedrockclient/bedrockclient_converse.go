package bedrockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/vxcontrol/langchaingo/internal/numutil"
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

	var resp *llms.ContentResponse
	if input.StreamingFunc != nil {
		resp, err = c.handleStreamingResponse(ctx, converseInput, input.StreamingFunc)
	} else {
		resp, err = c.handleNonStreamingResponse(ctx, converseInput)
	}
	if err != nil {
		return nil, err
	}
	if err := validateConverseStructuredOutput(input, resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// ConverseInput represents input for the Converse API
type ConverseInput struct {
	Messages         []Message
	ModelID          string
	MaxTokens        *int
	Temperature      *float64
	TopP             *float64
	StopSequences    []string
	Tools            []llms.Tool
	ToolChoice       any
	StreamingFunc    streaming.Callback
	ReasoningConfig  *llms.ReasoningConfig
	EnableCaching    bool
	StructuredOutput *llms.StructuredOutputConfig
}

type converseThinkingPayload struct {
	Type         string `json:"type" document:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty" document:"budget_tokens,omitempty"`
	// Display keeps adaptive reasoning text visible ("summarized"); Opus 4.7+
	// otherwise omits it by default.
	Display string `json:"display,omitempty" document:"display,omitempty"`
}

// converseOutputConfig carries the adaptive-thinking effort. It is a sibling of
// "thinking" inside additionalModelRequestFields.
type converseOutputConfig struct {
	Effort string `json:"effort,omitempty" document:"effort,omitempty"`
}

type converseAdditionalModelRequestFields struct {
	Thinking     *converseThinkingPayload `json:"thinking,omitempty" document:"thinking,omitempty"`
	OutputConfig *converseOutputConfig    `json:"output_config,omitempty" document:"output_config,omitempty"`
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
		inferenceConfig.MaxTokens = aws.Int32(numutil.SaturateInt32(*input.MaxTokens))
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
		toolConfig, err := c.convertToolsToToolConfig(input.Tools, input.ToolChoice)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		converseInput.ToolConfig = toolConfig
	}

	// Add additional model fields
	switch input.ReasoningConfig.ResolveMode() {
	case llms.ReasoningOn:
		additionalModelFields := converseAdditionalModelRequestFields{}
		maxTokens := 0 // Use 0 to let it use default maxTokens
		if input.MaxTokens != nil {
			maxTokens = *input.MaxTokens
		}
		setAdaptive := func() {
			effort := reasoning.ClaudeClampEffort(input.ModelID, string(input.ReasoningConfig.GetEffort(maxTokens)))
			additionalModelFields.Thinking = &converseThinkingPayload{Type: "adaptive", Display: "summarized"}
			additionalModelFields.OutputConfig = &converseOutputConfig{Effort: effort}
			// Adaptive models reject sampling params.
			inferenceConfig.Temperature = nil
			inferenceConfig.TopP = nil
		}
		setBudget := func() {
			if tokens := input.ReasoningConfig.GetTokens(maxTokens); tokens > 0 {
				additionalModelFields.Thinking = &converseThinkingPayload{Type: "enabled", BudgetTokens: tokens}
				if reasoning.ClaudeSupportsEffortWithBudget(input.ModelID, reasoning.ProviderBedrock) {
					effort := reasoning.ClaudeClampEffort(input.ModelID, string(input.ReasoningConfig.GetEffort(maxTokens)))
					additionalModelFields.OutputConfig = &converseOutputConfig{Effort: effort}
				}
				// Budget thinking requires temperature=1.0 and rejects top_p.
				if isAnthropicModelID(input.ModelID) {
					inferenceConfig.Temperature = aws.Float32(1.0)
					inferenceConfig.TopP = nil
				}
			}
		}
		switch reasoning.ResolveMechanism(input.ModelID, input.ReasoningConfig.Adaptive,
			isAnthropicModelID(input.ModelID), c.supportsReasoning(input.ModelID)) {
		case reasoning.MechanismAdaptive:
			setAdaptive()
		case reasoning.MechanismBudget:
			setBudget()
		}
		if additionalModelFields.Thinking != nil {
			converseInput.AdditionalModelRequestFields = document.NewLazyDocument(additionalModelFields)
		}
	case llms.ReasoningOff:
		switch reasoning.ResolveOff(input.ModelID, reasoning.ProviderBedrock) {
		case reasoning.OffDisableClaude:
			fields := converseAdditionalModelRequestFields{Thinking: &converseThinkingPayload{Type: "disabled"}}
			converseInput.AdditionalModelRequestFields = document.NewLazyDocument(fields)
		case reasoning.OffUnsupported:
			return nil, &reasoning.ErrReasoningOffUnsupported{Model: input.ModelID}
		}
	}

	// Adaptive-only models reject sampling params even without thinking.
	if reasoning.ClaudeRejectsSampling(input.ModelID) {
		inferenceConfig.Temperature = nil
		inferenceConfig.TopP = nil
	}
	if reasoning.ClaudeMutuallyExclusiveSampling(input.ModelID) &&
		inferenceConfig.Temperature != nil && inferenceConfig.TopP != nil {
		inferenceConfig.TopP = nil
	}

	// Native AWS structured output rides on the top-level OutputConfig.TextFormat,
	// independent of the provider-specific reasoning fields above.
	if err := applyConverseStructuredOutput(input, converseInput); err != nil {
		return nil, err
	}

	return converseInput, nil
}

// aiMessageAccumulator accumulates consecutive AI messages into a single assistant message
type aiMessageAccumulator struct {
	textContent   string
	reasoning     *reasoning.ContentReasoning
	toolUseBlocks []types.ContentBlock
	cacheControl  *CacheControl
	hasAnyContent bool
}

// addTextContent adds text content to the accumulator
func (a *aiMessageAccumulator) addTextContent(content string, reasoningContent *reasoning.ContentReasoning) {
	if a.textContent == "" {
		a.textContent = content
	}

	if a.reasoning == nil && reasoningContent != nil {
		a.reasoning = reasoningContent
	}

	a.hasAnyContent = true
}

// addToolUse adds a tool use block to the accumulator
func (a *aiMessageAccumulator) addToolUse(toolCall *ToolCall) error {
	if toolCall == nil {
		return nil
	}

	a.toolUseBlocks = append(a.toolUseBlocks, &types.ContentBlockMemberToolUse{
		Value: types.ToolUseBlock{
			ToolUseId: aws.String(toolCall.ID),
			Name:      aws.String(toolCall.Name),
			Input:     document.NewLazyDocument(toolCall.Arguments),
		},
	})

	a.hasAnyContent = true
	return nil
}

// setCacheControl sets cache control if not already set
func (a *aiMessageAccumulator) setCacheControl(cacheControl *CacheControl) {
	if a.cacheControl == nil {
		a.cacheControl = cacheControl
	}
}

// build creates a Converse Message from accumulated data
func (a *aiMessageAccumulator) build() types.Message {
	content := make([]types.ContentBlock, 0)

	// Add reasoning content if present
	if a.reasoning != nil && (a.reasoning.Content != "" || len(a.reasoning.Signature) > 0) {
		reasoningBlock := types.ReasoningContentBlockMemberReasoningText{
			Value: types.ReasoningTextBlock{
				Text:      aws.String(a.reasoning.Content),
				Signature: ptrStringOrNil(string(a.reasoning.Signature)),
			},
		}
		content = append(content, &types.ContentBlockMemberReasoningContent{
			Value: &reasoningBlock,
		})
	}

	// Add text content if present
	if a.textContent != "" {
		content = append(content, &types.ContentBlockMemberText{Value: a.textContent})
	}

	// Add all tool use blocks
	content = append(content, a.toolUseBlocks...)

	// Add cache point if needed
	if a.cacheControl != nil {
		cacheBlock := types.CachePointBlock{
			Type: types.CachePointTypeDefault,
		}
		if a.cacheControl.TTL == "1h" {
			cacheBlock.Ttl = types.CacheTTLOneHour
		} else {
			cacheBlock.Ttl = types.CacheTTLFiveMinutes
		}
		content = append(content, &types.ContentBlockMemberCachePoint{
			Value: cacheBlock,
		})
	}

	return types.Message{
		Role:    types.ConversationRoleAssistant,
		Content: content,
	}
}

// reset clears the accumulator
func (a *aiMessageAccumulator) reset() {
	a.textContent = ""
	a.reasoning = nil
	a.toolUseBlocks = nil
	a.cacheControl = nil
	a.hasAnyContent = false
}

// isEmpty returns true if no content has been accumulated
func (a *aiMessageAccumulator) isEmpty() bool {
	return !a.hasAnyContent
}

// toolResultAccumulator accumulates consecutive tool result messages into a single user message
type toolResultAccumulator struct {
	resultBlocks []types.ContentBlock
}

// addToolResult adds a tool result block to the accumulator
func (t *toolResultAccumulator) addToolResult(toolResult *ToolResult) error {
	if toolResult == nil {
		return fmt.Errorf("tool message missing tool result")
	}

	t.resultBlocks = append(t.resultBlocks, &types.ContentBlockMemberToolResult{
		Value: types.ToolResultBlock{
			ToolUseId: aws.String(toolResult.ToolCallID),
			Content: []types.ToolResultContentBlock{
				&types.ToolResultContentBlockMemberText{
					Value: toolResult.Content,
				},
			},
		},
	})

	return nil
}

// build creates a Converse Message from accumulated tool results
func (t *toolResultAccumulator) build() types.Message {
	return types.Message{
		Role:    types.ConversationRoleUser,
		Content: t.resultBlocks,
	}
}

// reset clears the accumulator
func (t *toolResultAccumulator) reset() {
	t.resultBlocks = nil
}

// isEmpty returns true if no tool results have been accumulated
func (t *toolResultAccumulator) isEmpty() bool {
	return len(t.resultBlocks) == 0
}

// convertMessages converts our messages to Converse format
// All consecutive AI messages (with or without tool calls) are combined into a single assistant message
// All consecutive tool result messages are combined into a single user message
func (c *ConverseClient) convertMessages(messages []Message) ([]types.Message, []types.SystemContentBlock, error) {
	var converseMessages []types.Message
	var systemPrompts []types.SystemContentBlock

	aiAccum := &aiMessageAccumulator{}
	toolAccum := &toolResultAccumulator{}

	// Helper to flush AI accumulator
	flushAI := func() error {
		if !aiAccum.isEmpty() {
			converseMessages = append(converseMessages, aiAccum.build())
			aiAccum.reset()
		}
		return nil
	}

	// Helper to flush tool result accumulator
	flushToolResults := func() error {
		if !toolAccum.isEmpty() {
			converseMessages = append(converseMessages, toolAccum.build())
			toolAccum.reset()
		}
		return nil
	}

	for i, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			// System messages don't need flushing accumulators
			systemPrompts = append(systemPrompts, &types.SystemContentBlockMemberText{
				Value: msg.Content,
			})

		case llms.ChatMessageTypeHuman:
			// Flush all pending accumulators before user message
			if err := flushAI(); err != nil {
				return nil, nil, err
			}
			if err := flushToolResults(); err != nil {
				return nil, nil, err
			}

			// Convert and add user message directly
			converseMsg, err := c.convertUserOrAssistantMessage(msg)
			if err != nil {
				return nil, nil, err
			}
			if msg.CacheControl != nil {
				converseMsg.Content = append(converseMsg.Content, c.createCachePointBlock(msg.CacheControl))
			}
			converseMessages = append(converseMessages, converseMsg)

		case llms.ChatMessageTypeAI:
			// Flush tool results if any before processing AI message
			if err := flushToolResults(); err != nil {
				return nil, nil, err
			}

			// Accumulate AI message content and tool calls
			if msg.ToolCall != nil {
				// Add tool call to accumulator
				if err := aiAccum.addToolUse(msg.ToolCall); err != nil {
					return nil, nil, err
				}
			} else {
				// Add text content to accumulator
				aiAccum.addTextContent(msg.Content, msg.Reasoning)
			}

			// Set cache control if present
			aiAccum.setCacheControl(msg.CacheControl)

			// Check if next message is also AI - if not, flush
			isLast := i == len(messages)-1
			nextIsNotAI := !isLast && messages[i+1].Role != llms.ChatMessageTypeAI

			if isLast || nextIsNotAI {
				if err := flushAI(); err != nil {
					return nil, nil, err
				}
			}

		case llms.ChatMessageTypeTool:
			// Flush AI if any before processing tool results
			if err := flushAI(); err != nil {
				return nil, nil, err
			}

			// Accumulate tool result
			if err := toolAccum.addToolResult(msg.ToolResult); err != nil {
				return nil, nil, err
			}

			// Check if next message is also tool result - if not, flush
			isLast := i == len(messages)-1
			nextIsNotTool := !isLast && messages[i+1].Role != llms.ChatMessageTypeTool

			if isLast || nextIsNotTool {
				if err := flushToolResults(); err != nil {
					return nil, nil, err
				}
			}
		}
	}

	// Flush any remaining accumulated messages
	if err := flushAI(); err != nil {
		return nil, nil, err
	}
	if err := flushToolResults(); err != nil {
		return nil, nil, err
	}

	return converseMessages, systemPrompts, nil
}

// createCachePointBlock creates a cache point block from cache control
func (c *ConverseClient) createCachePointBlock(cacheControl *CacheControl) *types.ContentBlockMemberCachePoint {
	cacheBlock := types.CachePointBlock{
		Type: types.CachePointTypeDefault,
	}
	if cacheControl.TTL == "1h" {
		cacheBlock.Ttl = types.CacheTTLOneHour
	} else {
		cacheBlock.Ttl = types.CacheTTLFiveMinutes
	}
	return &types.ContentBlockMemberCachePoint{
		Value: cacheBlock,
	}
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
					Text:      aws.String(msg.Reasoning.Content),
					Signature: ptrStringOrNil(string(msg.Reasoning.Signature)),
				},
			}
			contentBlocks = append(contentBlocks, &types.ContentBlockMemberReasoningContent{
				Value: &reasoningBlock,
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

// convertToolsToToolConfig converts llms.Tool to Converse ToolConfiguration
func (c *ConverseClient) convertToolsToToolConfig(tools []llms.Tool, choice any) (*types.ToolConfiguration, error) {
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
		ToolChoice: converseToolChoice(choice),
	}, nil
}

// converseToolChoice carries the caller's choice to the wire. An unset or
// unrecognized choice leaves the decision to the model.
func converseToolChoice(choice any) types.ToolChoice {
	name, forced := llms.ForcedToolName(choice)
	switch {
	case forced && name != "":
		return &types.ToolChoiceMemberTool{Value: types.SpecificToolChoice{Name: &name}}
	case forced:
		return &types.ToolChoiceMemberAny{}
	default:
		return &types.ToolChoiceMemberAuto{}
	}
}

type converseToolCallBuilder struct {
	id        string
	name      string
	arguments strings.Builder
}

func (b *converseToolCallBuilder) streamingCall() streaming.ToolCall {
	return streaming.ToolCall{ID: b.id, Name: b.name, Arguments: b.arguments.String()}
}

func (b *converseToolCallBuilder) toolCall() llms.ToolCall {
	return llms.ToolCall{
		ID:   b.id,
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      b.name,
			Arguments: b.arguments.String(),
		},
	}
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
		// Without this the schema silently vanishes on ConverseStream while
		// non-streaming works — structured output must ride both paths.
		OutputConfig: input.OutputConfig,
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
	var toolCalls []llms.ToolCall
	var stopReason string
	currentToolCalls := make(map[int32]*converseToolCallBuilder) // Track streaming tool calls by content block index

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
						case *types.ReasoningContentBlockDeltaMemberSignature:
							if len(block.Value) > 0 {
								signature.WriteString(block.Value)
							}
						}
					}
				case *types.ContentBlockDeltaMemberToolUse:
					if delta.Value.Input != nil {
						if builder, ok := currentToolCalls[aws.ToInt32(e.Value.ContentBlockIndex)]; ok {
							builder.arguments.WriteString(*delta.Value.Input)
						}
					}
				}
			}
		case *types.ConverseStreamOutputMemberContentBlockStart:
			if e.Value.Start != nil {
				if toolUse, ok := e.Value.Start.(*types.ContentBlockStartMemberToolUse); ok {
					currentToolCalls[aws.ToInt32(e.Value.ContentBlockIndex)] = &converseToolCallBuilder{
						id:   aws.ToString(toolUse.Value.ToolUseId),
						name: aws.ToString(toolUse.Value.Name),
					}
				}
			}
		case *types.ConverseStreamOutputMemberContentBlockStop:
			index := aws.ToInt32(e.Value.ContentBlockIndex)
			if builder, ok := currentToolCalls[index]; ok {
				delete(currentToolCalls, index)
				if callback != nil {
					chunk := streaming.Chunk{
						Type:     streaming.ChunkTypeToolCall,
						ToolCall: builder.streamingCall(),
					}
					if err := callback(ctx, chunk); err != nil {
						return nil, err
					}
				}
				toolCalls = append(toolCalls, builder.toolCall())
			}
		case *types.ConverseStreamOutputMemberMessageStop:
			// The terminal event carries the stop reason (end_turn, tool_use,
			// max_tokens, guardrail_intervened, content_filtered, ...).
			stopReason = string(e.Value.StopReason)
			// Stream completed - ensure any remaining tool calls are added
			for _, index := range slices.Sorted(maps.Keys(currentToolCalls)) {
				toolCalls = append(toolCalls, currentToolCalls[index].toolCall())
			}
			currentToolCalls = nil
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	var sig []byte
	if signature.Len() > 0 {
		sig = signature.Bytes()
	}

	choice := &llms.ContentChoice{
		Content:        fullContent.String(),
		ToolCalls:      toolCalls,
		GenerationInfo: make(map[string]any),
		Reasoning:      c.processReasoning(reasoningContent.String(), sig),
		StopReason:     stopReason,
		Truncated:      llms.IsTruncated(stopReason),
	}

	result := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{choice},
	}

	return result, nil
}

func (c *ConverseClient) processReasoning(reasoningContent string, signature []byte) *reasoning.ContentReasoning {
	if reasoningContent == "" && len(signature) == 0 {
		return nil
	}

	return &reasoning.ContentReasoning{
		Content:   reasoningContent,
		Signature: signature,
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
				var argsJSON []byte
				var err error
				if block.Value.Input != nil {
					argsJSON, err = block.Value.Input.MarshalSmithyDocument()
					if err != nil {
						return nil, fmt.Errorf("failed to marshal tool input document: %w", err)
					}
				} else {
					argsJSON = []byte("{}")
				}
				if argsJSON == nil {
					argsJSON = []byte("{}")
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
					choice.Reasoning = c.processReasoning(reasoningText, sig)
				}
			}
		}

	}

	// The stop reason lives on the response, not inside types.Message. Surfacing it
	// distinguishes end_turn from tool_use, max_tokens, guardrail_intervened,
	// content_filtered and malformed-output states.
	choice.StopReason = string(response.StopReason)
	choice.Truncated = llms.IsTruncated(string(response.StopReason))

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
	// Claude 4+ and select non-Claude models support reasoning.
	reasoningModels := []string{
		"anthropic.claude-fable-5",
		"anthropic.claude-sonnet-5",
		"anthropic.claude-opus-5",
		"anthropic.claude-opus-4-",
		"anthropic.claude-sonnet-4-",
		"anthropic.claude-haiku-4-",
		"openai.gpt-oss-120b",
		"openai.gpt-oss-20b",
		"moonshot.kimi-k2-thinking",
		"minimax.minimax-m2.5",
		"minimax.minimax-m2.1",
	}

	for _, model := range reasoningModels {
		if strings.Contains(modelID, model) {
			return true
		}
	}
	return false
}

// isAnthropicModelID reports whether the Bedrock model ID belongs to the Claude
// family (with or without a region prefix such as "us.").
func isAnthropicModelID(modelID string) bool {
	return strings.Contains(modelID, "anthropic.claude")
}

func ptrStringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
