package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrInvalidContentType       = errors.New("invalid content type")
	ErrUnsupportedMessageType   = errors.New("unsupported message type")
	ErrUnsupportedContentType   = errors.New("unsupported content type")
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *anthropicclient.Client
	model            string // Track current model for reasoning detection
}

var (
	_ llms.Model          = (*LLM)(nil)
	_ llms.ReasoningModel = (*LLM)(nil)
)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to create client: %w", err)
	}
	return &LLM{
		client: c,
		model:  c.Model, // Store the model for reasoning detection
	}, nil
}

func newClient(opts ...Option) (*anthropicclient.Client, error) {
	options := &options{
		token:      os.Getenv(tokenEnvVarName),
		baseURL:    anthropicclient.DefaultBaseURL,
		httpClient: httputil.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return anthropicclient.New(options.token, options.model, options.baseURL,
		anthropicclient.WithHTTPClient(options.httpClient),
		anthropicclient.WithLegacyTextCompletionsAPI(options.useLegacyTextCompletionsAPI),
		anthropicclient.WithAnthropicBetaHeader(options.anthropicBetaHeader),
	)
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if o.client.UseLegacyTextCompletionsAPI {
		return generateCompletionsContent(ctx, o, messages, opts)
	}
	return generateMessagesContent(ctx, o, messages, opts)
}

func generateCompletionsContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
	if len(messages) == 0 || len(messages[0].Parts) == 0 {
		return nil, ErrEmptyResponse
	}

	msg0 := messages[0]
	part := msg0.Parts[0]
	partText, ok := part.(llms.TextContent)
	if !ok {
		return nil, fmt.Errorf("anthropic: unexpected message type: %T", part)
	}
	prompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", partText.Text)
	result, err := o.client.CreateCompletion(ctx, &anthropicclient.CompletionRequest{
		Model:         opts.Model,
		Prompt:        prompt,
		MaxTokens:     opts.MaxTokens,
		StopWords:     opts.StopWords,
		Temperature:   opts.Temperature,
		TopP:          opts.TopP,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("anthropic: failed to create completion: %w", err)
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}
	return resp, nil
}

func generateMessagesContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
	chatMessages, systemPrompt, err := processMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to process messages: %w", err)
	}

	tools := toolsToTools(opts.Tools)
	var toolChoice any

	// Handle structured output by simulating with tool calling
	var structuredOutputToolName string
	if opts.StructuredOutput != nil {
		// Convert schema to tool definition
		structuredTool := schemaToTool(opts.StructuredOutput)
		tools = append(tools, structuredTool)
		structuredOutputToolName = structuredTool.Name

		// Force the model to use this tool
		toolChoice = map[string]any{
			"type": "tool",
			"name": structuredTool.Name,
		}
	} else if opts.ToolChoice != nil {
		toolChoice = opts.ToolChoice
	}

	betaHeaders, thinking := extractThinkingOptions(o, opts)

	// Enforce temperature = 1.0 when thinking is enabled (Anthropic requirement)
	temperature := opts.Temperature
	if thinking != nil && temperature != 1.0 {
		temperature = 1.0
	}

	result, err := o.client.CreateMessage(ctx, &anthropicclient.MessageRequest{
		Model:                  opts.Model,
		Messages:               chatMessages,
		System:                 systemPrompt,
		MaxTokens:              opts.MaxTokens,
		StopWords:              opts.StopWords,
		Temperature:            temperature,
		TopP:                   opts.TopP,
		Tools:                  tools,
		ToolChoice:             toolChoice,
		Thinking:               thinking,
		BetaHeaders:            betaHeaders,
		StreamingFunc:          opts.StreamingFunc,
		StreamingReasoningFunc: opts.StreamingReasoningFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("anthropic: failed to create message: %w", err)
	}

	// If structured output was requested, extract JSON from tool call
	if structuredOutputToolName != "" {
		return extractStructuredOutput(result, structuredOutputToolName)
	}

	return processAnthropicResponse(result, thinking)
}

// processAnthropicResponse converts Anthropic API response to standard ContentResponse
func processAnthropicResponse(result *anthropicclient.MessageResponsePayload, thinkingConfig *anthropicclient.ThinkingConfig) (*llms.ContentResponse, error) {
	if result == nil || len(result.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	// Collect all content blocks and separate thinking from output
	var textParts []string
	var thinkingParts []string
	var toolCalls []llms.ToolCall
	var thinkingTokens int

	for _, content := range result.Content {
		switch content.GetType() {
		case "text":
			if textContent, ok := content.(*anthropicclient.TextContent); ok {
				textParts = append(textParts, textContent.Text)
			}
		case "tool_use":
			if toolUseContent, ok := content.(*anthropicclient.ToolUseContent); ok {
				argumentsJSON, err := json.Marshal(toolUseContent.Input)
				if err != nil {
					return nil, fmt.Errorf("anthropic: failed to marshal tool use arguments: %w", err)
				}
				toolCalls = append(toolCalls, llms.ToolCall{
					ID: toolUseContent.ID,
					FunctionCall: &llms.FunctionCall{
						Name:      toolUseContent.Name,
						Arguments: string(argumentsJSON),
					},
				})
			}
		case "thinking":
			if thinkingContent, ok := content.(*anthropicclient.ThinkingContent); ok {
				thinkingParts = append(thinkingParts, thinkingContent.Thinking)
				// Estimate thinking tokens (rough approximation: 1 token ~= 4 chars)
				thinkingTokens += len(thinkingContent.Thinking) / 4
			}
		}
	}

	// Combine all text parts
	combinedText := strings.Join(textParts, "\n")
	combinedThinking := strings.Join(thinkingParts, "\n")

	// Build generation info with thinking token details
	genInfo := map[string]any{
		"InputTokens":              result.Usage.InputTokens,
		"OutputTokens":             result.Usage.OutputTokens,
		"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
		"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
	}

	// Add thinking-specific fields if thinking was enabled
	if len(thinkingParts) > 0 {
		genInfo["ThinkingContent"] = combinedThinking
		genInfo["ThinkingTokens"] = thinkingTokens
		if thinkingConfig != nil {
			genInfo["ThinkingBudgetAllocated"] = thinkingConfig.BudgetTokens
			genInfo["ThinkingBudgetUsed"] = thinkingTokens
		}
	}

	// Create the choice
	choice := &llms.ContentChoice{
		Content:        combinedText,
		ToolCalls:      toolCalls,
		StopReason:     result.StopReason,
		GenerationInfo: genInfo,
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{choice},
	}, nil
}

func toolsToTools(tools []llms.Tool) []anthropicclient.Tool {
	toolReq := make([]anthropicclient.Tool, len(tools))
	for i, tool := range tools {
		toolReq[i] = anthropicclient.Tool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters,
		}
	}
	return toolReq
}

func processMessages(messages []llms.MessageContent) ([]anthropicclient.ChatMessage, string, error) {
	chatMessages := make([]anthropicclient.ChatMessage, 0, len(messages))
	systemPrompt := ""
	for _, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			content, err := handleSystemMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle system message: %w", err)
			}
			systemPrompt += content
		case llms.ChatMessageTypeHuman:
			chatMessage, err := handleHumanMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle human message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeAI:
			chatMessage, err := handleAIMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle AI message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeTool:
			chatMessage, err := handleToolMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle tool message: %w", err)
			}
			chatMessages = append(chatMessages, chatMessage)
		case llms.ChatMessageTypeGeneric, llms.ChatMessageTypeFunction:
			return nil, "", fmt.Errorf("anthropic: %w: %v", ErrUnsupportedMessageType, msg.Role)
		default:
			return nil, "", fmt.Errorf("anthropic: %w: %v", ErrUnsupportedMessageType, msg.Role)
		}
	}
	return chatMessages, systemPrompt, nil
}

func handleSystemMessage(msg llms.MessageContent) (string, error) {
	// Handle both direct TextContent and CachedContent wrapper
	part := msg.Parts[0]

	// If it's cached content, unwrap it
	if cached, ok := part.(llms.CachedContent); ok {
		part = cached.ContentPart
	}

	// Extract text from the part
	if textContent, ok := part.(llms.TextContent); ok {
		return textContent.Text, nil
	}

	return "", fmt.Errorf("anthropic: %w for system message", ErrInvalidContentType)
}

func handleHumanMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	var contents []anthropicclient.Content

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.CachedContent:
			// Handle cached content with cache control
			var cacheControl *anthropicclient.CacheControl
			if p.CacheControl != nil {
				cacheControl = &anthropicclient.CacheControl{
					Type: p.CacheControl.Type,
				}
			}

			// Process the wrapped content
			switch wrapped := p.ContentPart.(type) {
			case llms.TextContent:
				contents = append(contents, &anthropicclient.TextContent{
					Type:         "text",
					Text:         wrapped.Text,
					CacheControl: cacheControl,
				})
			case llms.BinaryContent:
				contents = append(contents, &anthropicclient.ImageContent{
					Type: "image",
					Source: anthropicclient.ImageSource{
						Type:      "base64",
						MediaType: wrapped.MIMEType,
						Data:      base64.StdEncoding.EncodeToString(wrapped.Data),
					},
					CacheControl: cacheControl,
				})
			default:
				return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: unsupported cached content part type: %T", wrapped)
			}
		case llms.TextContent:
			contents = append(contents, &anthropicclient.TextContent{
				Type: "text",
				Text: p.Text,
			})
		case llms.BinaryContent:
			contents = append(contents, &anthropicclient.ImageContent{
				Type: "image",
				Source: anthropicclient.ImageSource{
					Type:      "base64",
					MediaType: p.MIMEType,
					Data:      base64.StdEncoding.EncodeToString(p.Data),
				},
			})
		default:
			return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: unsupported human message part type: %T", part)
		}
	}

	if len(contents) == 0 {
		return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: no valid content in human message")
	}

	return anthropicclient.ChatMessage{
		Role:    RoleUser,
		Content: contents,
	}, nil
}

func handleAIMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	if toolCall, ok := msg.Parts[0].(llms.ToolCall); ok {
		var inputStruct map[string]interface{}
		err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &inputStruct)
		if err != nil {
			return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: failed to unmarshal tool call arguments: %w", err)
		}
		toolUse := anthropicclient.ToolUseContent{
			Type:  "tool_use",
			ID:    toolCall.ID,
			Name:  toolCall.FunctionCall.Name,
			Input: inputStruct,
		}

		return anthropicclient.ChatMessage{
			Role:    RoleAssistant,
			Content: []anthropicclient.Content{toolUse},
		}, nil
	}
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return anthropicclient.ChatMessage{
			Role: RoleAssistant,
			Content: []anthropicclient.Content{&anthropicclient.TextContent{
				Type: "text",
				Text: textContent.Text,
			}},
		}, nil
	}
	return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for AI message", ErrInvalidContentType)
}

type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

func handleToolMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	if toolCallResponse, ok := msg.Parts[0].(llms.ToolCallResponse); ok {
		toolContent := anthropicclient.ToolResultContent{
			Type:      "tool_result",
			ToolUseID: toolCallResponse.ToolCallID,
			Content:   toolCallResponse.Content,
		}

		return anthropicclient.ChatMessage{
			Role:    RoleUser,
			Content: []anthropicclient.Content{toolContent},
		}, nil
	}
	return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for tool message", ErrInvalidContentType)
}

// SupportsReasoning implements the ReasoningModel interface.
// Returns true if the current model supports extended thinking capabilities.
func (o *LLM) SupportsReasoning() bool {
	return supportsReasoningForModel(o.model)
}

// supportsReasoningForModel checks if a specific model supports reasoning.
// This is a separate function to avoid race conditions when checking capabilities.
func supportsReasoningForModel(model string) bool {
	if model == "" {
		return false
	}

	modelLower := strings.ToLower(model)

	// Claude 3.7+ supports extended thinking
	if strings.Contains(modelLower, "claude-3-7") ||
		strings.Contains(modelLower, "claude-3.7") ||
		strings.Contains(modelLower, "claude-3-7-sonnet") {
		return true
	}

	// Claude 4+ supports extended thinking with interleaving
	if strings.Contains(modelLower, "claude-4") ||
		strings.Contains(modelLower, "claude-opus-4") ||
		strings.Contains(modelLower, "claude-sonnet-4") {
		return true
	}

	// Future Claude 5+ expected to support reasoning
	if strings.Contains(modelLower, "claude-5") ||
		strings.Contains(modelLower, "claude-opus-5") ||
		strings.Contains(modelLower, "claude-sonnet-5") {
		return true
	}

	return false
}

// extractThinkingOptions extracts thinking configuration and beta headers from call options
func extractThinkingOptions(o *LLM, opts *llms.CallOptions) ([]string, *anthropicclient.ThinkingConfig) {
	// Extract beta headers for prompt caching support
	var betaHeaders []string
	if opts.Metadata != nil {
		if headers, ok := opts.Metadata["anthropic:beta_headers"].([]string); ok {
			betaHeaders = headers
		}
	}

	// Check if reasoning is enabled via new unified API
	if opts.Reasoning == nil {
		return betaHeaders, nil
	}

	// Determine current model
	currentModel := opts.Model
	if currentModel == "" {
		currentModel = o.model
	}

	// Only configure thinking for models that support extended thinking
	if !supportsReasoningForModel(currentModel) {
		return betaHeaders, nil
	}

	// Calculate budget tokens
	var budgetTokens int
	if opts.Reasoning.BudgetTokens != nil && *opts.Reasoning.BudgetTokens > 0 {
		budgetTokens = *opts.Reasoning.BudgetTokens
	} else if opts.Reasoning.Mode != "" {
		// Map ThinkingMode to token budget
		switch opts.Reasoning.Mode {
		case llms.ThinkingModeLow:
			budgetTokens = 2000
		case llms.ThinkingModeMedium:
			budgetTokens = 8000
		case llms.ThinkingModeHigh:
			budgetTokens = 16000
		}
	}

	// Ensure budget is within valid range for Claude 3.7+
	if budgetTokens > 0 {
		if budgetTokens < 1024 {
			budgetTokens = 1024 // Minimum for Claude
		} else if budgetTokens > 128000 {
			budgetTokens = 128000 // Maximum for Claude (128K)
		}
	}

	// Add interleaved thinking header if requested (Claude 4+)
	if opts.Reasoning.Interleaved {
		betaHeaders = append(betaHeaders, "interleaved-thinking-2025-05-14")
	}

	// Add extended thinking header
	if budgetTokens > 0 {
		betaHeaders = append(betaHeaders, "extended-thinking-2025-01-01")
	}

	// Create thinking configuration if we have a budget
	var thinking *anthropicclient.ThinkingConfig
	if budgetTokens > 0 {
		thinking = &anthropicclient.ThinkingConfig{
			Type:         "enabled",
			BudgetTokens: budgetTokens,
		}
	}

	return betaHeaders, thinking
}

// extractThinkingFromText extracts thinking content from Anthropic responses
// Anthropic models often embed thinking in <thinking> tags
func extractThinkingFromText(fullText string) (thinkingContent, outputContent string) {
	// Look for <thinking> tags in the text
	if strings.Contains(fullText, "<thinking>") {
		start := strings.Index(fullText, "<thinking>")
		end := strings.Index(fullText, "</thinking>")
		if start >= 0 && end > start {
			// Extract thinking content between tags
			thinkingContent = fullText[start+10 : end] // +10 for "<thinking>"

			// Extract output content (everything before and after thinking tags)
			beforeThinking := strings.TrimSpace(fullText[:start])
			afterThinking := ""
			if end+12 < len(fullText) { // +12 for "</thinking>"
				afterThinking = strings.TrimSpace(fullText[end+12:])
			}

			// Combine non-thinking content
			if beforeThinking != "" && afterThinking != "" {
				outputContent = beforeThinking + "\n\n" + afterThinking
			} else if beforeThinking != "" {
				outputContent = beforeThinking
			} else {
				outputContent = afterThinking
			}

			return strings.TrimSpace(thinkingContent), strings.TrimSpace(outputContent)
		}
	}

	// If no thinking tags found, treat entire text as output
	return "", fullText
}

// schemaToTool converts a StructuredOutputSchema to an Anthropic tool definition.
// This is used to simulate structured output via tool calling.
func schemaToTool(def *llms.StructuredOutputDefinition) anthropicclient.Tool {
	return anthropicclient.Tool{
		Name:        def.Name,
		Description: def.Description,
		InputSchema: convertSchemaToJSONSchema(def.Schema),
	}
}

// convertSchemaToJSONSchema converts llms.StructuredOutputSchema to JSON Schema format
func convertSchemaToJSONSchema(schema *llms.StructuredOutputSchema) map[string]any {
	if schema == nil {
		return nil
	}

	result := map[string]any{
		"type": string(schema.Type),
	}

	if schema.Description != "" {
		result["description"] = schema.Description
	}

	// Handle object properties
	if schema.Type == llms.SchemaTypeObject && len(schema.Properties) > 0 {
		properties := make(map[string]any)
		for name, prop := range schema.Properties {
			properties[name] = convertSchemaToJSONSchema(prop)
		}
		result["properties"] = properties

		if len(schema.Required) > 0 {
			result["required"] = schema.Required
		}

		if !schema.AdditionalProperties {
			result["additionalProperties"] = false
		}
	}

	// Handle array items
	if schema.Type == llms.SchemaTypeArray && schema.Items != nil {
		result["items"] = convertSchemaToJSONSchema(schema.Items)

		if schema.MinItems != nil {
			result["minItems"] = *schema.MinItems
		}
		if schema.MaxItems != nil {
			result["maxItems"] = *schema.MaxItems
		}
	}

	// Handle string constraints
	if schema.Type == llms.SchemaTypeString {
		if len(schema.Enum) > 0 {
			result["enum"] = schema.Enum
		}
		if schema.MinLength != nil {
			result["minLength"] = *schema.MinLength
		}
		if schema.MaxLength != nil {
			result["maxLength"] = *schema.MaxLength
		}
		if schema.Pattern != "" {
			result["pattern"] = schema.Pattern
		}
	}

	// Handle number/integer constraints
	if schema.Type == llms.SchemaTypeNumber || schema.Type == llms.SchemaTypeInteger {
		if schema.Minimum != nil {
			result["minimum"] = *schema.Minimum
		}
		if schema.Maximum != nil {
			result["maximum"] = *schema.Maximum
		}
	}

	return result
}

// extractStructuredOutput extracts JSON from a tool call response when using structured output
func extractStructuredOutput(result *anthropicclient.MessageResponsePayload, toolName string) (*llms.ContentResponse, error) {
	if result == nil || len(result.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	// Look for the tool_use content block with our structured output tool
	for _, content := range result.Content {
		if content.GetType() == "tool_use" {
			if toolUseContent, ok := content.(*anthropicclient.ToolUseContent); ok {
				if toolUseContent.Name == toolName {
					// Convert the tool input to JSON string
					jsonBytes, err := json.Marshal(toolUseContent.Input)
					if err != nil {
						return nil, fmt.Errorf("anthropic: failed to marshal structured output: %w", err)
					}

					return &llms.ContentResponse{
						Choices: []*llms.ContentChoice{
							{
								Content:    string(jsonBytes),
								StopReason: result.StopReason,
								GenerationInfo: map[string]any{
									"InputTokens":              result.Usage.InputTokens,
									"OutputTokens":             result.Usage.OutputTokens,
									"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
									"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
									"StructuredOutput":         true,
								},
							},
						},
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("anthropic: structured output tool call not found in response")
}
