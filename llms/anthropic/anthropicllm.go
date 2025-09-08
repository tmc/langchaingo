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

	betaHeaders, thinking := extractThinkingOptions(o, opts)

	result, err := o.client.CreateMessage(ctx, &anthropicclient.MessageRequest{
		Model:         opts.Model,
		Messages:      chatMessages,
		System:        systemPrompt,
		MaxTokens:     opts.MaxTokens,
		StopWords:     opts.StopWords,
		Temperature:   opts.Temperature,
		TopP:          opts.TopP,
		Tools:         tools,
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

	choices := make([]*llms.ContentChoice, len(result.Content))
	for i, content := range result.Content {
		switch content.GetType() {
		case "text":
			if textContent, ok := content.(*anthropicclient.TextContent); ok {
				// Extract thinking content from the response text
				thinkingContent, outputContent := extractThinkingFromText(textContent.Text)

				choices[i] = &llms.ContentChoice{
					Content:    textContent.Text,
					StopReason: result.StopReason,
					GenerationInfo: map[string]any{
						"InputTokens":              result.Usage.InputTokens,
						"OutputTokens":             result.Usage.OutputTokens,
						"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
						"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
						// Standardized fields for cross-provider compatibility
						"ThinkingContent": thinkingContent, // Standardized field
						"OutputContent":   outputContent,   // Standardized field
					},
				}
			} else {
				return nil, fmt.Errorf("anthropic: %w for text message", ErrInvalidContentType)
			}
		case "tool_use":
			if toolUseContent, ok := content.(*anthropicclient.ToolUseContent); ok {
				argumentsJSON, err := json.Marshal(toolUseContent.Input)
				if err != nil {
					return nil, fmt.Errorf("anthropic: failed to marshal tool use arguments: %w", err)
				}
				choices[i] = &llms.ContentChoice{
					ToolCalls: []llms.ToolCall{
						{
							ID: toolUseContent.ID,
							FunctionCall: &llms.FunctionCall{
								Name:      toolUseContent.Name,
								Arguments: string(argumentsJSON),
							},
						},
					},
					StopReason: result.StopReason,
					GenerationInfo: map[string]any{
						"InputTokens":              result.Usage.InputTokens,
						"OutputTokens":             result.Usage.OutputTokens,
						"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
						"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
					},
				}
			} else {
				return nil, fmt.Errorf("anthropic: %w for tool use message %T", ErrInvalidContentType, content)
			}
		case "thinking":
			if thinkingContent, ok := content.(*anthropicclient.ThinkingContent); ok {
				choices[i] = &llms.ContentChoice{
					Content:    "", // Thinking content is not included in output
					StopReason: result.StopReason,
					GenerationInfo: map[string]any{
						"ThinkingContent":          thinkingContent.Thinking,
						"ThinkingSignature":        thinkingContent.Signature,
						"InputTokens":              result.Usage.InputTokens,
						"OutputTokens":             result.Usage.OutputTokens,
						"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
						"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
					},
				}
			} else {
				return nil, fmt.Errorf("anthropic: %w for thinking message %T", ErrInvalidContentType, content)
			}
		default:
			return nil, fmt.Errorf("anthropic: %w: %v", ErrUnsupportedContentType, content.GetType())
		}
	}

	return &llms.ContentResponse{
		Choices: choices,
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

	// Extract thinking configuration
	var budgetTokens int
	if opts.Metadata != nil {
		if config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig); ok {
			// Only set budget_tokens for models that support extended thinking
			// Claude 3.7+ and Claude 4+ support this feature
			currentModel := opts.Model
			if currentModel == "" {
				currentModel = o.model
			}
			if supportsReasoningForModel(currentModel) {
				if config.BudgetTokens > 0 {
					budgetTokens = config.BudgetTokens
				} else if config.Mode != llms.ThinkingModeNone {
					// Calculate budget based on mode
					budgetTokens = llms.CalculateThinkingBudget(config.Mode, opts.MaxTokens)
				}

				// Ensure budget is within valid range for Claude 3.7+
				if budgetTokens > 0 {
					if budgetTokens < 1024 {
						budgetTokens = 1024 // Minimum for Claude
					} else if budgetTokens > 128000 {
						budgetTokens = 128000 // Maximum for Claude (128K)
					}
				}
			}

			// Add interleaved thinking header if requested (Claude 4+)
			if config.InterleaveThinking && supportsReasoningForModel(currentModel) {
				betaHeaders = append(betaHeaders, "interleaved-thinking-2025-05-14")
			}
		}
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
