package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
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

	betaHeaders, thinking, outputConfig := extractThinkingOptions(o, opts)

	// Models that reject the temperature and top_p sampling parameters
	// require them omitted from the request.
	temperature := &opts.Temperature
	topP := opts.TopP
	if !modelCapabilities(o.resolvedModel(opts)).sampling {
		temperature = nil
		topP = 0
	}

	result, err := o.client.CreateMessage(ctx, &anthropicclient.MessageRequest{
		Model:                  opts.Model,
		Messages:               chatMessages,
		System:                 systemPrompt,
		MaxTokens:              opts.MaxTokens,
		StopWords:              opts.StopWords,
		Temperature:            temperature,
		TopP:                   topP,
		Tools:                  tools,
		Thinking:               thinking,
		OutputConfig:           outputConfig,
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
	return processAnthropicResponse(result)
}

// processAnthropicResponse converts Anthropic API response to standard ContentResponse
func processAnthropicResponse(result *anthropicclient.MessageResponsePayload) (*llms.ContentResponse, error) {
	if result == nil || len(result.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	choices := make([]*llms.ContentChoice, len(result.Content))
	for i, content := range result.Content {
		choice, err := contentBlockToChoice(result, content)
		if err != nil {
			return nil, err
		}
		choices[i] = choice
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil
}

// generationInfo builds the per-choice GenerationInfo map: the usage
// counters shared by every choice plus any block-specific extras.
func generationInfo(result *anthropicclient.MessageResponsePayload, extra map[string]any) map[string]any {
	info := map[string]any{
		"InputTokens":              result.Usage.InputTokens,
		"OutputTokens":             result.Usage.OutputTokens,
		"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
		"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
	}
	maps.Copy(info, extra)
	return info
}

func contentBlockToChoice(result *anthropicclient.MessageResponsePayload, content anthropicclient.Content) (*llms.ContentChoice, error) {
	switch block := content.(type) {
	case *anthropicclient.TextContent:
		// Extract thinking content from the response text
		thinkingContent, outputContent := extractThinkingFromText(block.Text)
		return &llms.ContentChoice{
			Content:    block.Text,
			StopReason: result.StopReason,
			Parts:      []llms.ContentPart{llms.TextContent{Text: block.Text}},
			GenerationInfo: generationInfo(result, map[string]any{
				// Standardized fields for cross-provider compatibility
				"ThinkingContent": thinkingContent,
				"OutputContent":   outputContent,
			}),
		}, nil
	case *anthropicclient.ToolUseContent:
		argumentsJSON, err := json.Marshal(block.Input)
		if err != nil {
			return nil, fmt.Errorf("anthropic: failed to marshal tool use arguments: %w", err)
		}
		toolCall := llms.ToolCall{
			ID: block.ID,
			FunctionCall: &llms.FunctionCall{
				Name:      block.Name,
				Arguments: string(argumentsJSON),
			},
		}
		return &llms.ContentChoice{
			ToolCalls:      []llms.ToolCall{toolCall},
			StopReason:     result.StopReason,
			Parts:          []llms.ContentPart{toolCall},
			GenerationInfo: generationInfo(result, nil),
		}, nil
	case *anthropicclient.ThinkingContent:
		return &llms.ContentChoice{
			Content:          "", // Thinking content is not included in output
			StopReason:       result.StopReason,
			ReasoningContent: block.Thinking,
			Parts: []llms.ContentPart{llms.ThinkingContent{
				Thinking:  block.Thinking,
				Signature: block.Signature,
			}},
			GenerationInfo: generationInfo(result, map[string]any{
				"ThinkingContent":   block.Thinking,
				"ThinkingSignature": block.Signature,
			}),
		}, nil
	case *anthropicclient.RedactedThinkingContent:
		return &llms.ContentChoice{
			Content:    "", // Redacted thinking content is encrypted and not included in output
			StopReason: result.StopReason,
			Parts:      []llms.ContentPart{llms.RedactedThinkingContent{Data: block.Data}},
			GenerationInfo: generationInfo(result, map[string]any{
				"RedactedThinkingData": block.Data,
			}),
		}, nil
	default:
		return nil, fmt.Errorf("anthropic: %w: %v", ErrUnsupportedContentType, content.GetType())
	}
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

// handleAIMessage converts an assistant message, preserving all parts in
// order. Thinking blocks keep their signatures and redacted thinking its
// data, both of which the API requires sent back verbatim in tool loops.
func handleAIMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	var contents []anthropicclient.Content
	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.ThinkingContent:
			contents = append(contents, &anthropicclient.ThinkingContent{
				Type:      "thinking",
				Thinking:  p.Thinking,
				Signature: p.Signature,
			})
		case llms.RedactedThinkingContent:
			contents = append(contents, &anthropicclient.RedactedThinkingContent{
				Type: "redacted_thinking",
				Data: p.Data,
			})
		case llms.TextContent:
			contents = append(contents, &anthropicclient.TextContent{
				Type: "text",
				Text: p.Text,
			})
		case llms.ToolCall:
			var inputStruct map[string]interface{}
			if err := json.Unmarshal([]byte(p.FunctionCall.Arguments), &inputStruct); err != nil {
				return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: failed to unmarshal tool call arguments: %w", err)
			}
			contents = append(contents, anthropicclient.ToolUseContent{
				Type:  "tool_use",
				ID:    p.ID,
				Name:  p.FunctionCall.Name,
				Input: inputStruct,
			})
		default:
			return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for AI message: %T", ErrInvalidContentType, part)
		}
	}
	if len(contents) == 0 {
		return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for AI message", ErrInvalidContentType)
	}
	return anthropicclient.ChatMessage{
		Role:    RoleAssistant,
		Content: contents,
	}, nil
}

type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

// handleToolMessage converts a tool message, preserving all parts: the
// results of parallel tool calls belong in a single user message.
func handleToolMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	var contents []anthropicclient.Content
	for _, part := range msg.Parts {
		toolCallResponse, ok := part.(llms.ToolCallResponse)
		if !ok {
			return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for tool message: %T", ErrInvalidContentType, part)
		}
		contents = append(contents, anthropicclient.ToolResultContent{
			Type:      "tool_result",
			ToolUseID: toolCallResponse.ToolCallID,
			Content:   toolCallResponse.Content,
		})
	}
	if len(contents) == 0 {
		return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for tool message", ErrInvalidContentType)
	}
	return anthropicclient.ChatMessage{
		Role:    RoleUser,
		Content: contents,
	}, nil
}

// SupportsReasoning implements the ReasoningModel interface.
// Returns true if the current model supports extended thinking capabilities.
func (o *LLM) SupportsReasoning() bool {
	return supportsReasoningForModel(o.model)
}

// resolvedModel returns the model used for a call: the per-call override
// if set, otherwise the client's configured model.
func (o *LLM) resolvedModel(opts *llms.CallOptions) string {
	if opts.Model != "" {
		return opts.Model
	}
	return o.model
}

// capabilities describes what a model accepts on the wire.
type capabilities struct {
	thinking         bool // extended or adaptive thinking supported
	adaptiveThinking bool // thinking is {"type": "adaptive"} only; budgets and explicit "disabled" are rejected
	sampling         bool // temperature and top_p accepted
	effort           bool // output_config.effort accepted
}

var (
	// adaptiveCaps is the newest model generation: adaptive thinking
	// only, sampling parameters rejected with a 400 error, effort
	// levels accepted. Claude Fable 5 additionally rejects an explicit
	// {"type": "disabled"}, so when thinking is off the parameter must
	// be omitted entirely.
	adaptiveCaps = capabilities{thinking: true, adaptiveThinking: true, effort: true}

	// budgetCaps are budget-era thinking models (Claude 3.7 to 4.x).
	budgetCaps = capabilities{thinking: true, sampling: true}

	// legacyCaps are models without extended thinking.
	legacyCaps = capabilities{sampling: true}
)

// capabilityRules maps model-name substrings to capabilities; the first
// match wins, so more specific names come first.
var capabilityRules = []struct {
	match string
	caps  capabilities
}{
	{"claude-fable", adaptiveCaps},
	{"claude-opus-4-7", adaptiveCaps},
	{"claude-opus-4-8", adaptiveCaps},
	{"claude-opus-4-6", capabilities{thinking: true, sampling: true, effort: true}},
	{"claude-sonnet-4-6", capabilities{thinking: true, sampling: true, effort: true}},
	{"claude-opus-4", budgetCaps},
	{"claude-sonnet-4", budgetCaps},
	{"claude-haiku-4", budgetCaps},
	{"claude-4", budgetCaps},
	{"claude-3-7", budgetCaps},
	{"claude-3.7", budgetCaps},
	{"claude-3", legacyCaps},
	{"claude-2", legacyCaps},
	{"claude-instant", legacyCaps},
}

// modelCapabilities returns the capabilities of a model. Unknown model
// names get newest-generation defaults: an unrecognized name is more
// likely a new model than an old one, omitting sampling parameters is
// accepted by every model, and sending them is rejected by new ones.
// The empty model name keeps legacy behavior, since the effective
// default model is resolved later, at the client layer.
func modelCapabilities(model string) capabilities {
	if model == "" {
		return legacyCaps
	}
	modelLower := strings.ToLower(model)
	for _, rule := range capabilityRules {
		if strings.Contains(modelLower, rule.match) {
			return rule.caps
		}
	}
	return adaptiveCaps
}

// supportsReasoningForModel checks if a specific model supports reasoning.
func supportsReasoningForModel(model string) bool {
	return modelCapabilities(model).thinking
}

// adaptiveThinkingOnly reports whether the model accepts only adaptive
// thinking ({"type": "adaptive"}).
func adaptiveThinkingOnly(model string) bool {
	return modelCapabilities(model).adaptiveThinking
}

// extractThinkingOptions extracts the thinking configuration, output
// controls, and beta headers from call options.
func extractThinkingOptions(o *LLM, opts *llms.CallOptions) ([]string, *anthropicclient.ThinkingConfig, *anthropicclient.OutputConfig) {
	// Extract beta headers for prompt caching support
	var betaHeaders []string
	if opts.Metadata != nil {
		if headers, ok := opts.Metadata["anthropic:beta_headers"].([]string); ok {
			betaHeaders = headers
		}
	}

	// Extract thinking configuration. Only models that support extended
	// thinking (Claude 3.7+) honor it.
	config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		return betaHeaders, nil, nil
	}
	currentModel := o.resolvedModel(opts)
	caps := modelCapabilities(currentModel)
	if !caps.thinking {
		return betaHeaders, nil, nil
	}

	var outputConfig *anthropicclient.OutputConfig
	if caps.effort && config.Effort != "" {
		outputConfig = &anthropicclient.OutputConfig{Effort: string(config.Effort)}
	}

	// Models with adaptive thinking choose their own budget, so
	// budget_tokens is never sent. Adaptive thinking interleaves
	// automatically, so the interleaved-thinking beta header is
	// unnecessary.
	if caps.adaptiveThinking {
		if config.BudgetTokens > 0 || config.Mode != llms.ThinkingModeNone {
			return betaHeaders, &anthropicclient.ThinkingConfig{
				Type:    "adaptive",
				Display: string(config.Display),
			}, outputConfig
		}
		return betaHeaders, nil, outputConfig
	}

	var budgetTokens int
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

	// Add interleaved thinking header if requested (Claude 4+)
	if config.InterleaveThinking {
		betaHeaders = append(betaHeaders, "interleaved-thinking-2025-05-14")
	}

	// Create thinking configuration if we have a budget
	var thinking *anthropicclient.ThinkingConfig
	if budgetTokens > 0 {
		thinking = &anthropicclient.ThinkingConfig{
			Type:         "enabled",
			BudgetTokens: budgetTokens,
		}
	}

	return betaHeaders, thinking, outputConfig
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
