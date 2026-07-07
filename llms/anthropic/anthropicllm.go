package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"
	"time"

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
	// require them omitted from the request. Budget-era extended
	// thinking additionally rejects any temperature other than the
	// default, so sampling parameters are omitted whenever a thinking
	// payload is sent.
	temperature := &opts.Temperature
	topP := opts.TopP
	if !modelCapabilities(o.resolvedModel(opts)).sampling || thinking != nil {
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

// processMessages converts messages to the wire format. The second return
// value is the system prompt: a plain string when no system part carries
// cache control, otherwise a []anthropicclient.TextContent block list.
func processMessages(messages []llms.MessageContent) ([]anthropicclient.ChatMessage, any, error) {
	chatMessages := make([]anthropicclient.ChatMessage, 0, len(messages))
	var systemBlocks []anthropicclient.TextContent
	for _, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			blocks, err := handleSystemMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle system message: %w", err)
			}
			systemBlocks = append(systemBlocks, blocks...)
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
	return chatMessages, encodeSystem(systemBlocks), nil
}

// encodeSystem picks the wire encoding for the system prompt. The plain
// string form is preserved when no block carries cache control, keeping
// requests byte-identical with earlier releases.
func encodeSystem(blocks []anthropicclient.TextContent) any {
	for _, b := range blocks {
		if b.CacheControl != nil {
			return blocks
		}
	}
	var sb strings.Builder
	for _, b := range blocks {
		sb.WriteString(b.Text)
	}
	return sb.String()
}

// handleSystemMessage converts a system message to text blocks, preserving
// cache control (#1456).
func handleSystemMessage(msg llms.MessageContent) ([]anthropicclient.TextContent, error) {
	var blocks []anthropicclient.TextContent
	for _, part := range msg.Parts {
		var cacheControl *anthropicclient.CacheControl
		if cached, ok := part.(llms.CachedContent); ok {
			cacheControl = cacheControlToClient(cached.CacheControl)
			part = cached.ContentPart
		}
		textContent, ok := part.(llms.TextContent)
		if !ok {
			return nil, fmt.Errorf("anthropic: %w for system message: %T", ErrInvalidContentType, part)
		}
		blocks = append(blocks, anthropicclient.TextContent{
			Type:         "text",
			Text:         textContent.Text,
			CacheControl: cacheControl,
		})
	}
	return blocks, nil
}

// cacheControlToClient maps cache control to the wire form. The type
// defaults to "ephemeral" and Duration maps to the ttl values the API
// accepts: at most five minutes (or zero, the default) omits ttl, longer
// durations request the one-hour cache.
func cacheControlToClient(cc *llms.CacheControl) *anthropicclient.CacheControl {
	if cc == nil {
		return nil
	}
	out := &anthropicclient.CacheControl{Type: cc.Type}
	if out.Type == "" {
		out.Type = "ephemeral"
	}
	if cc.Duration > 5*time.Minute {
		out.TTL = "1h"
	}
	return out
}

func handleHumanMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	var contents []anthropicclient.Content

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.CachedContent:
			// Handle cached content with cache control
			cacheControl := cacheControlToClient(p.CacheControl)

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
	return supportsReasoningForModel(o.resolvedModel(nil))
}

// resolvedModel returns the model a call targets: the per-call override
// if set, then the client's configured model, then the client default.
// Capability lookups must use this, never the raw option value, so they
// agree with the model the request is actually sent to.
func (o *LLM) resolvedModel(opts *llms.CallOptions) string {
	if opts != nil && opts.Model != "" {
		return opts.Model
	}
	if o.model != "" {
		return o.model
	}
	return anthropicclient.DefaultModel
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
	{"claude-4", budgetCaps},
	{"claude-3-7", budgetCaps},
	{"claude-3.7", budgetCaps},
	{"claude-3", legacyCaps},
	{"claude-2", legacyCaps},
	{"claude-instant", legacyCaps},
}

// familyRules covers members of the 4-x model families that
// capabilityRules does not list individually. lastListed is the
// highest minor version whose capabilities are known; later minors
// get newest-generation defaults, since claude-opus-4-9 is more
// likely next month's adaptive-only model than one missing from the
// table, and omitting sampling parameters is accepted by every model
// while sending them is a 400 on new ones. Bump lastListed when
// adding a family member to capabilityRules.
var familyRules = []struct {
	prefix     string
	lastListed int
}{
	{"claude-opus-4", 8},
	{"claude-sonnet-4", 6},
	{"claude-haiku-4", 5},
}

// modelCapabilities returns the capabilities of a model. Unknown model
// names get newest-generation defaults: an unrecognized name is more
// likely a new model than an old one, omitting sampling parameters is
// accepted by every model, and sending them is rejected by new ones.
func modelCapabilities(model string) capabilities {
	modelLower := strings.ToLower(model)
	for _, rule := range capabilityRules {
		if strings.Contains(modelLower, rule.match) {
			return rule.caps
		}
	}
	for _, f := range familyRules {
		n, ok := familyMinor(modelLower, f.prefix)
		if !ok {
			continue
		}
		if n > f.lastListed {
			return adaptiveCaps
		}
		return budgetCaps
	}
	// A name that is not a Claude model at all — a gateway alias, a
	// fine-tune, an Anthropic-compatible endpoint — says nothing about
	// which parameters it rejects, so assume the long-standing wire
	// contract. Caller-supplied sampling parameters then reach it
	// instead of being dropped silently, and an endpoint that rejects
	// them reports a clear error. An empty name means "use the client
	// default", which is a Claude model, so it falls through.
	if modelLower != "" && !strings.Contains(modelLower, "claude") {
		return budgetCaps
	}
	// An unrecognized Claude model is assumed newer than every listed
	// one, and newer models are adaptive-thinking only.
	return adaptiveCaps
}

// familyMinor extracts the minor version from model names like
// "claude-opus-4-N..." for the given family prefix ("claude-opus-4").
// Bare family names and dated snapshots ("claude-opus-4-20250514")
// name the .0 model and report minor 0. The second result is false
// when the name does not belong to the family.
func familyMinor(model, prefix string) (int, bool) {
	i := strings.Index(model, prefix)
	if i < 0 {
		return 0, false
	}
	rest := model[i+len(prefix):]
	if !strings.HasPrefix(rest, "-") {
		return 0, true
	}
	rest = rest[1:]
	j := 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	if j == 0 || j > 2 {
		return 0, true
	}
	n, err := strconv.Atoi(rest[:j])
	if err != nil {
		return 0, true
	}
	return n, true
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
