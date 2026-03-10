package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic/internal/anthropicclient"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
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
	CallbacksHandler     callbacks.Handler
	client               *anthropicclient.Client
	defaultCacheStrategy *CacheStrategy
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	options := &options{
		token:      os.Getenv(tokenEnvVarName),
		baseURL:    anthropicclient.DefaultBaseURL,
		httpClient: httputil.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	c, err := newClientFromOptions(options)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to create client: %w", err)
	}

	return &LLM{
		client:               c,
		defaultCacheStrategy: options.defaultCacheStrategy,
	}, nil
}

func newClientFromOptions(options *options) (*anthropicclient.Client, error) {
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
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint:lll
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if o.client.UseLegacyTextCompletionsAPI {
		resp, err := generateCompletionsContent(ctx, o, messages, opts)
		if err != nil {
			return nil, err
		}

		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
		}

		return resp, nil
	}

	resp, err := generateMessagesContent(ctx, o, messages, opts)
	if err != nil {
		return nil, err
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}

	return resp, nil
}

func generateCompletionsContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) { //nolint:lll
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
		Model:         opts.GetModel(),
		Prompt:        prompt,
		MaxTokens:     opts.GetMaxTokens(),
		StopWords:     opts.GetStopWords(),
		Temperature:   opts.GetTemperature(),
		TopP:          opts.GetTopP(),
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

func generateMessagesContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) { //nolint:lll,funlen,cyclop
	chatMessages, systemPrompt, err := processMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to process messages: %w", err)
	}

	var thinking *anthropicclient.ThinkingPayload
	if opts.Reasoning.IsEnabled() {
		thinking = &anthropicclient.ThinkingPayload{
			Type:   "enabled",
			Budget: opts.Reasoning.GetTokens(opts.GetMaxTokens()),
		}
	}

	tools := toolsToTools(opts.Tools)

	// Merge client-level and call-level cache strategies
	if mergedStrategy := mergeCacheStrategies(o.defaultCacheStrategy, opts); mergedStrategy != nil {
		applyCacheStrategy(&tools, &systemPrompt, &chatMessages, *mergedStrategy)
	}

	betaHeaders := extractBetaHeaders(opts)

	// Convert system prompt to appropriate type
	system := systemPrompt
	if systemPrompt != nil {
		switch sp := systemPrompt.(type) {
		case string:
			if sp == "" {
				system = nil
			} else {
				system = sp
			}
		case []anthropicclient.Content:
			if len(sp) == 0 {
				system = nil
			} else {
				system = sp
			}
		}
	}

	// Set temperature to 1.0 for thinking models
	temperature, maxTokens := opts.Temperature, opts.GetMaxTokens()
	if thinking != nil && thinking.Type == "enabled" && thinking.Budget > 0 {
		temperature = getFloatPointer(1.0)
		maxTokens = max(thinking.Budget*2, maxTokens) // 2x the budget for thinking
	}

	result, err := o.client.CreateMessage(ctx, &anthropicclient.MessageRequest{
		Model:         opts.GetModel(),
		Messages:      chatMessages,
		System:        system,
		MaxTokens:     &maxTokens,
		StopWords:     opts.StopWords,
		Temperature:   temperature,
		TopP:          opts.TopP,
		Tools:         tools,
		ToolChoice:    opts.ToolChoice,
		Thinking:      thinking,
		BetaHeaders:   betaHeaders,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("anthropic: failed to create message: %w", err)
	}
	return processAnthropicResponse(result)
}

// processAnthropicResponse converts Anthropic API response to standard ContentResponse
func processAnthropicResponse(result *anthropicclient.MessageResponsePayload) (*llms.ContentResponse, error) {
	if result == nil || len(result.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	// Extract ALL thinking content, signature
	// According to Anthropic docs, there's ONE thinking block per response
	var reasoningContent strings.Builder
	var signature []byte

	for _, content := range result.Content {
		switch cv := content.(type) {
		case *anthropicclient.ThinkingContent:
			reasoningContent.WriteString(cv.Thinking)
			if len(cv.Signature) > 0 {
				signature = []byte(cv.Signature)
			}
		}
	}

	// Create reasoning object
	var contentReasoning *reasoning.ContentReasoning
	if reasoningContent.Len() > 0 || len(signature) > 0 {
		contentReasoning = &reasoning.ContentReasoning{
			Content:   reasoningContent.String(),
			Signature: signature,
		}
	}

	// Process content blocks to collect text and tool calls
	var toolCalls []llms.ToolCall
	var textContent string

	for _, content := range result.Content {
		switch cv := content.(type) {
		case *anthropicclient.TextContent:
			textContent = cv.Text
		case *anthropicclient.ToolUseContent:
			argumentsJSON, err := json.Marshal(cv.Input)
			if err != nil {
				return nil, fmt.Errorf("anthropic: failed to marshal tool use arguments: %w", err)
			}
			toolCall := llms.ToolCall{
				ID:   cv.ID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      cv.Name,
					Arguments: string(argumentsJSON),
				},
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	// Build response choice - reasoning ALWAYS goes to choice, not tool calls
	choice := &llms.ContentChoice{
		Content:    textContent,
		Reasoning:  contentReasoning, // Always in choice for Anthropic
		ToolCalls:  toolCalls,
		StopReason: result.StopReason,
		GenerationInfo: map[string]any{
			// Standardized field names for cross-provider compatibility
			"PromptTokens":             result.Usage.InputTokens + result.Usage.CacheCreationInputTokens + result.Usage.CacheReadInputTokens,
			"CompletionTokens":         result.Usage.OutputTokens,
			"TotalTokens":              result.Usage.InputTokens + result.Usage.CacheCreationInputTokens + result.Usage.CacheReadInputTokens + result.Usage.OutputTokens,
			"ReasoningTokens":          0, // Reasoning tokens are not included in the usage metrics
			"PromptCachedTokens":       result.Usage.CacheReadInputTokens,
			"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
			"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
			// Special fields for Anthropic cache creation
			"CacheCreationEphemeral5mInputTokens": result.Usage.CacheCreation.Ephemeral5mInputTokens,
			"CacheCreationEphemeral1hInputTokens": result.Usage.CacheCreation.Ephemeral1hInputTokens,
			// Special fields for Anthropic
			"InputTokens":  result.Usage.InputTokens,
			"OutputTokens": result.Usage.OutputTokens,
			"ServiceTier":  result.Usage.ServiceTier,
		},
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

// parseBase64URI returns values data, media type from a base64 URI and error if invalid.
func parseBase64URI(uri string) (string, string, error) {
	re := regexp.MustCompile(`^data:(.*?);base64,(.*)$`)
	matches := re.FindStringSubmatch(uri)
	if len(matches) != 3 {
		return "", "", errors.New("invalid base64 URI")
	}

	return matches[2], matches[1], nil
}

// mergeCacheStrategies merges client-level and call-level cache strategies.
// Call-level strategy takes precedence on a per-field basis.
//
// Merge logic:
// - If call-level has a field set to true, use it (overrides client-level)
// - If call-level field is false, check client-level field
// - TTL: call-level takes precedence if set, otherwise use client-level
//
// Returns nil if no strategy is defined at either level.
func mergeCacheStrategies(clientStrategy *CacheStrategy, opts *llms.CallOptions) *CacheStrategy {
	// Extract call-level strategy from metadata
	var callStrategy *CacheStrategy
	if opts.Metadata != nil {
		if cs, ok := opts.Metadata["anthropic:cache_strategy"].(CacheStrategy); ok {
			callStrategy = &cs
		}
	}

	// No strategies defined at any level
	if clientStrategy == nil && callStrategy == nil {
		return nil
	}

	// Only call-level defined
	if clientStrategy == nil {
		return callStrategy
	}

	// Only client-level defined
	if callStrategy == nil {
		return clientStrategy
	}

	// Both defined - merge with call-level priority
	merged := CacheStrategy{
		CacheTools:    callStrategy.CacheTools || clientStrategy.CacheTools,
		CacheSystem:   callStrategy.CacheSystem || clientStrategy.CacheSystem,
		CacheMessages: callStrategy.CacheMessages || clientStrategy.CacheMessages,
	}

	// TTL: call-level takes precedence
	if callStrategy.TTL != "" {
		merged.TTL = callStrategy.TTL
	} else {
		merged.TTL = clientStrategy.TTL
	}

	return &merged
}

// convertCacheControl converts shared llms.CacheControl to Anthropic-specific format
func convertCacheControl(llmCache *llms.CacheControl) *anthropicclient.CacheControl {
	if llmCache == nil {
		return nil
	}

	anthropicCache := &anthropicclient.CacheControl{
		Type: llmCache.Type,
	}

	// Convert duration to TTL string
	if llmCache.Duration > 0 {
		if llmCache.Duration >= time.Hour {
			anthropicCache.TTL = "1h"
		} else {
			anthropicCache.TTL = "5m" // Default
		}
	}

	return anthropicCache
}

// applyCacheStrategy applies automatic cache control based on strategy.
// Anthropic cache hierarchy: tools → system → messages
// Each breakpoint caches the entire prefix up to that point.
func applyCacheStrategy(
	tools *[]anthropicclient.Tool,
	systemPrompt *any,
	chatMessages *[]anthropicclient.ChatMessage,
	strategy CacheStrategy,
) {
	// Determine TTL
	ttl := strategy.TTL
	if ttl == "" {
		ttl = "5m" // Default to 5-minute cache
	}

	cacheControl := &anthropicclient.CacheControl{
		Type: "ephemeral",
		TTL:  ttl,
	}

	// 1. Cache tools (if requested and tools exist)
	if strategy.CacheTools && tools != nil && len(*tools) > 0 {
		// Per Anthropic docs: mark the LAST tool in the array
		(*tools)[len(*tools)-1].CacheControl = cacheControl
	}

	// 2. Cache system (if requested and system exists)
	if strategy.CacheSystem && systemPrompt != nil {
		switch sp := (*systemPrompt).(type) {
		case []anthropicclient.Content:
			// System is in array format - mark last content block
			if len(sp) > 0 {
				markLastContentBlockForCaching(sp, cacheControl)
				*systemPrompt = sp
			}
		case string:
			// System is a string - convert to array format to add cache control
			if sp != "" {
				systemBlocks := []anthropicclient.Content{
					&anthropicclient.TextContent{
						Type:         "text",
						Text:         sp,
						CacheControl: cacheControl,
					},
				}
				*systemPrompt = systemBlocks
			}
		}
	}

	// 3. Cache messages (if requested and messages exist)
	if strategy.CacheMessages && chatMessages != nil && len(*chatMessages) > 0 {
		// Per Anthropic docs: mark the LAST content block of the LAST message
		// This creates a cache breakpoint that includes ALL conversation history up to this point:
		// - First turn: caches the initial prompt
		// - Subsequent turns: caches prompt + all previous assistant responses + tool results
		// This enables incremental caching where each turn reads the previous cache + writes new content
		lastMessage := &(*chatMessages)[len(*chatMessages)-1]
		if len(lastMessage.Content) > 0 {
			markLastContentBlockForCaching(lastMessage.Content, cacheControl)
		}
	}
}

// markLastContentBlockForCaching marks the last content block with cache control.
func markLastContentBlockForCaching(contents []anthropicclient.Content, cacheControl *anthropicclient.CacheControl) {
	if len(contents) == 0 {
		return
	}

	lastIdx := len(contents) - 1
	switch c := contents[lastIdx].(type) {
	case *anthropicclient.TextContent:
		// Only set if not already set (explicit cache control takes precedence)
		if c.CacheControl == nil {
			c.CacheControl = cacheControl
		}
	case *anthropicclient.ToolResultContent:
		if c.CacheControl == nil {
			c.CacheControl = cacheControl
		}
	case *anthropicclient.ImageContent:
		if c.CacheControl == nil {
			c.CacheControl = cacheControl
		}
	case *anthropicclient.ToolUseContent:
		if c.CacheControl == nil {
			c.CacheControl = cacheControl
		}
	}
}

func processMessages(messages []llms.MessageContent) ([]anthropicclient.ChatMessage, any, error) {
	chatMessages := make([]anthropicclient.ChatMessage, 0, len(messages))
	var systemPrompt any = ""
	var systemBlocks []anthropicclient.Content

	for _, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			var err error
			systemPrompt, systemBlocks, err = processSystemParts(msg, systemPrompt, systemBlocks)
			if err != nil {
				return nil, "", err
			}
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

	// If we collected system blocks, use them instead of string
	if len(systemBlocks) > 0 {
		systemPrompt = systemBlocks
	}

	return chatMessages, systemPrompt, nil
}

func processSystemParts(msg llms.MessageContent, systemPrompt any, systemBlocks []anthropicclient.Content) (any, []anthropicclient.Content, error) {
	hasCacheControl := false
	for _, part := range msg.Parts {
		if _, ok := part.(CachedContent); ok {
			hasCacheControl = true
			break
		}
	}

	if hasCacheControl {
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case CachedContent:
				cacheControl := convertCacheControl(p.CacheControl)
				if textContent, ok := p.ContentPart.(llms.TextContent); ok {
					systemBlocks = append(systemBlocks, &anthropicclient.TextContent{
						Type:         "text",
						Text:         textContent.Text,
						CacheControl: cacheControl,
					})
				}
			case llms.TextContent:
				systemBlocks = append(systemBlocks, &anthropicclient.TextContent{
					Type: "text",
					Text: p.Text,
				})
			}
		}
	} else {
		content, err := handleSystemMessage(msg)
		if err != nil {
			return nil, nil, fmt.Errorf("anthropic: failed to handle system message: %w", err)
		}
		if sysStr, ok := systemPrompt.(string); ok {
			systemPrompt = sysStr + content
		}
	}

	return systemPrompt, systemBlocks, nil
}

func handleSystemMessage(msg llms.MessageContent) (string, error) {
	// System message in Anthropic doesn't support cache_control directly
	// Cache control for system messages is handled via system parameter
	// For now, just extract text and ignore cache control
	// TODO: Handle system message caching via array format if needed

	part := msg.Parts[0]

	// If it's cached content, unwrap it
	if cached, ok := part.(CachedContent); ok {
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
		case CachedContent:
			// Handle cached content with cache control
			cacheControl := convertCacheControl(p.CacheControl)

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
		case llms.ImageURLContent:
			data, mediaType, err := parseBase64URI(p.URL)
			if err != nil {
				return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for human message", err)
			}
			contents = append(contents, anthropicclient.ImageContent{
				Type: "image",
				Source: anthropicclient.ImageSource{
					Type:      "base64",
					MediaType: mediaType,
					Data:      data,
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
	message := anthropicclient.ChatMessage{
		Role:    RoleAssistant,
		Content: []anthropicclient.Content{},
	}

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			// If reasoning is present, add blocks in order: thinking → text
			if p.Reasoning != nil {
				// Add thinking block if present
				if len(p.Reasoning.Content) > 0 || len(p.Reasoning.Signature) > 0 {
					thinkingBlock := &anthropicclient.ThinkingContent{
						Type:     "thinking",
						Thinking: p.Reasoning.Content,
					}
					if len(p.Reasoning.Signature) > 0 {
						thinkingBlock.Signature = string(p.Reasoning.Signature)
					}
					message.Content = append(message.Content, thinkingBlock)
				}
			}

			// Add text content only if not empty
			// In interleaved thinking, we might have thinking-only response
			if p.Text != "" {
				textContent := &anthropicclient.TextContent{
					Type: "text",
					Text: p.Text,
				}
				message.Content = append(message.Content, textContent)
			}
		case llms.ToolCall:
			if p.FunctionCall == nil {
				continue
			}

			var inputStruct map[string]interface{}
			if err := json.Unmarshal([]byte(p.FunctionCall.Arguments), &inputStruct); err != nil {
				err = fmt.Errorf("anthropic: failed to unmarshal tool call arguments: %w", err)
				return anthropicclient.ChatMessage{}, err
			}

			toolUse := &anthropicclient.ToolUseContent{
				Type:  "tool_use",
				ID:    p.ID,
				Name:  p.FunctionCall.Name,
				Input: inputStruct,
			}
			message.Content = append(message.Content, toolUse)
		default:
			return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for AI message", ErrInvalidContentType)
		}
	}

	return message, nil
}

type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

func handleToolMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	if toolCallResponse, ok := msg.Parts[0].(llms.ToolCallResponse); ok {
		toolContent := &anthropicclient.ToolResultContent{
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

// extractBetaHeaders extracts beta headers from call options
func extractBetaHeaders(opts *llms.CallOptions) []string {
	// Extract beta headers for prompt caching support
	var betaHeaders []string
	if opts.Metadata != nil {
		if headers, ok := opts.Metadata["anthropic:beta_headers"].([]string); ok {
			// Filter out empty headers
			for _, h := range headers {
				if h != "" {
					betaHeaders = append(betaHeaders, h)
				}
			}
		}
	}

	// Auto-enable interleaved thinking when reasoning + tools are present
	if opts.Reasoning.IsEnabled() && len(opts.Tools) > 0 {
		betaHeaders = appendIfMissing(betaHeaders, "interleaved-thinking-2025-05-14")
	}

	return betaHeaders
}

// appendIfMissing appends val to slice if not already present
func appendIfMissing(slice []string, val string) []string {
	for _, item := range slice {
		if item == val {
			return slice // Already present
		}
	}
	return append(slice, val)
}

func getFloatPointer(f float64) *float64 {
	return &f
}
