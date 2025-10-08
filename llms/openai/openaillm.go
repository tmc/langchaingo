package openai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

type ChatMessage = openaiclient.ChatMessage

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *openaiclient.Client
	model            string // Track current model for reasoning detection
}

const (
	RoleSystem    = "system"
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleFunction  = "function"
	RoleTool      = "tool"
)

// ModelCapability defines what a model supports
type ModelCapability struct {
	Pattern          string // Regex pattern to match model names
	SupportsSystem   bool   // If true, supports system messages
	SupportsThinking bool   // If true, supports reasoning/thinking
	SupportsCaching  bool   // If true, supports prompt caching
	// Add more capabilities as needed
}

// modelCapabilities defines capabilities for different model patterns
var modelCapabilities = []ModelCapability{
	// OpenAI reasoning models (o1, o3 series) - no system message support
	{
		Pattern:          `(?i)^o[13](-mini|-preview)?$`, // Matches o1, o1-mini, o1-preview, o3, o3-mini
		SupportsSystem:   false,                          // O1 models don't support system messages
		SupportsThinking: true,
		SupportsCaching:  false,
	},
	// GPT-4 models
	{
		Pattern:          `(?i)^gpt-4`, // Matches gpt-4, gpt-4-turbo, etc.
		SupportsSystem:   true,
		SupportsThinking: false,
		SupportsCaching:  false, // OpenAI caching coming soon
	},
	// GPT-3.5 models
	{
		Pattern:          `(?i)^gpt-3\.5`,
		SupportsSystem:   true,
		SupportsThinking: false,
		SupportsCaching:  false,
	},
	// Future models can be added here
}

// getModelCapabilities returns the capabilities for a given model
func getModelCapabilities(model string) ModelCapability {
	for _, cap := range modelCapabilities {
		if matched, _ := regexp.MatchString(cap.Pattern, model); matched {
			return cap
		}
	}
	// Default capabilities - assume standard model
	return ModelCapability{
		SupportsSystem:   true,
		SupportsThinking: false,
		SupportsCaching:  false,
	}
}

var (
	_ llms.Model          = (*LLM)(nil)
	_ llms.ReasoningModel = (*LLM)(nil)
)

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	opt, c, err := newClient(opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client:           c,
		CallbacksHandler: opt.callbackHandler,
		model:            c.Model, // Store the model for reasoning detection
	}, err
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, funlen
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Determine the effective model for this request (don't mutate o.model to avoid races)
	effectiveModel := opts.Model
	if effectiveModel == "" {
		effectiveModel = o.model
	}

	// Get capabilities for this model
	modelCaps := getModelCapabilities(effectiveModel)

	// For models that don't support system messages, we need to merge them into user messages
	var systemContent string
	if !modelCaps.SupportsSystem {
		for _, mc := range messages {
			if mc.Role == llms.ChatMessageTypeSystem {
				// Extract system message content
				for _, part := range mc.Parts {
					if textPart, ok := part.(llms.TextContent); ok {
						if systemContent != "" {
							systemContent += "\n\n"
						}
						systemContent += textPart.Text
					}
				}
			}
		}
	}

	chatMsgs := make([]*ChatMessage, 0, len(messages))
	for _, mc := range messages {
		// Skip system messages for models that don't support them
		if mc.Role == llms.ChatMessageTypeSystem && !modelCaps.SupportsSystem {
			continue
		}

		msg := &ChatMessage{MultiContent: mc.Parts}
		switch mc.Role {
		case llms.ChatMessageTypeSystem:
			msg.Role = RoleSystem
		case llms.ChatMessageTypeAI:
			msg.Role = RoleAssistant
		case llms.ChatMessageTypeHuman:
			msg.Role = RoleUser
			// For models without system support, prepend system content to first user message
			if systemContent != "" && !modelCaps.SupportsSystem {
				// Prepend system content to the user message
				newParts := []llms.ContentPart{}
				if systemContent != "" {
					newParts = append(newParts, llms.TextContent{Text: systemContent + "\n\n"})
				}
				newParts = append(newParts, mc.Parts...)
				msg.MultiContent = newParts
				systemContent = "" // Clear after using
			}
		case llms.ChatMessageTypeGeneric:
			msg.Role = RoleUser
		case llms.ChatMessageTypeFunction:
			msg.Role = RoleFunction
			// Extract name and content from ToolCallResponse for function messages
			if len(mc.Parts) == 1 {
				if p, ok := mc.Parts[0].(llms.ToolCallResponse); ok {
					msg.Name = p.Name
					msg.Content = p.Content
				}
			}
		case llms.ChatMessageTypeTool:
			msg.Role = RoleTool
			// Here we extract tool calls from the message and populate the ToolCalls field.

			// parse mc.Parts (which should have one entry of type ToolCallResponse) and populate msg.Content and msg.ToolCallID
			if len(mc.Parts) != 1 {
				return nil, fmt.Errorf("expected exactly one part for role %v, got %v", mc.Role, len(mc.Parts))
			}
			switch p := mc.Parts[0].(type) {
			case llms.ToolCallResponse:
				msg.ToolCallID = p.ToolCallID
				msg.Content = p.Content
			default:
				return nil, fmt.Errorf("expected part of type ToolCallResponse for role %v, got %T", mc.Role, mc.Parts[0])
			}

		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}

		// Here we extract tool calls from the message and populate the ToolCalls field.
		newParts, toolCalls := ExtractToolParts(msg)
		msg.MultiContent = newParts
		msg.ToolCalls = toolCallsFromToolCalls(toolCalls)

		chatMsgs = append(chatMsgs, msg)
	}
	// Check if we should use the legacy max_tokens field
	useLegacyMaxTokens := false
	if opts.Metadata != nil {
		if v, ok := opts.Metadata["openai:use_legacy_max_tokens"].(bool); ok {
			useLegacyMaxTokens = v
		}
	}

	// Extract reasoning effort for thinking models
	// Note: OpenAI o1/o3 models have built-in reasoning and don't support reasoning_effort parameter
	// This is kept for future models that might support it (like GPT-5)
	var reasoningEffort string
	// Commented out for now since current o1 models don't support this parameter
	/*
		if opts.Metadata != nil {
			if config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig); ok {
				// Map thinking mode to reasoning effort
				switch config.Mode {
				case llms.ThinkingModeLow:
					reasoningEffort = "low"
				case llms.ThinkingModeMedium:
					reasoningEffort = "medium"
				case llms.ThinkingModeHigh:
					reasoningEffort = "high"
				}

				// Handle streaming for thinking
				if config.StreamThinking && opts.StreamingReasoningFunc == nil && opts.StreamingFunc != nil {
					// Set up default reasoning streaming if requested but not provided
					// Wrap the single-param streaming func into a reasoning func
					opts.StreamingReasoningFunc = func(ctx context.Context, reasoningChunk []byte, chunk []byte) error {
						// For default behavior, we might want to stream both or just the main content
						// Here we'll just stream the main content chunk
						if len(chunk) > 0 {
							return opts.StreamingFunc(ctx, chunk)
						}
						return nil
					}
				}
			}
		}
	*/

	// Filter out internal metadata that shouldn't be sent to API
	apiMetadata := make(map[string]any)
	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			// Skip internal metadata keys
			if k == "thinking_config" || strings.HasPrefix(k, "openai:") {
				continue
			}
			apiMetadata[k] = v
		}
	}
	// Only include metadata if there are actual values to send
	if len(apiMetadata) == 0 {
		apiMetadata = nil
	}

	req := &openaiclient.ChatRequest{
		Model:                  opts.Model,
		StopWords:              opts.StopWords,
		Messages:               chatMsgs,
		StreamingFunc:          opts.StreamingFunc,
		StreamingReasoningFunc: opts.StreamingReasoningFunc,
		Temperature:            opts.Temperature,
		N:                      opts.N,
		FrequencyPenalty:       opts.FrequencyPenalty,
		PresencePenalty:        opts.PresencePenalty,
		ReasoningEffort:        reasoningEffort,

		// Token handling: check metadata flag for legacy behavior
		// By default use max_completion_tokens (modern field)
		// If WithLegacyMaxTokensField() is used, use max_tokens instead
		MaxCompletionTokens: func() int {
			if useLegacyMaxTokens {
				return 0 // Don't set max_completion_tokens
			}
			return opts.MaxTokens
		}(),
		MaxTokens: func() int {
			if useLegacyMaxTokens {
				return opts.MaxTokens // Set the legacy field
			}
			return 0 // Don't set max_tokens
		}(),

		ToolChoice:           opts.ToolChoice,
		FunctionCallBehavior: openaiclient.FunctionCallBehavior(opts.FunctionCallBehavior),
		Seed:                 opts.Seed,
		Metadata:             apiMetadata,
	}
	if opts.JSONMode {
		req.ResponseFormat = ResponseFormatJSON
	}

	// since req.Functions is deprecated, we need to use the new Tools API.
	for _, fn := range opts.Functions {
		req.Tools = append(req.Tools, openaiclient.Tool{
			Type: "function",
			Function: openaiclient.FunctionDefinition{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.Parameters,
				Strict:      fn.Strict,
			},
		})
	}
	// if opts.Tools is not empty, append them to req.Tools
	for _, tool := range opts.Tools {
		t, err := toolFromTool(tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert llms tool to openai tool: %w", err)
		}
		req.Tools = append(req.Tools, t)
	}

	// if o.client.ResponseFormat is set, use it for the request
	if o.client.ResponseFormat != nil {
		req.ResponseFormat = o.client.ResponseFormat
	}

	result, err := o.client.CreateChat(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	choices := make([]*llms.ContentChoice, len(result.Choices))
	for i, c := range result.Choices {
		choices[i] = &llms.ContentChoice{
			Content:          c.Message.Content,
			ReasoningContent: c.Message.ReasoningContent,
			StopReason:       fmt.Sprint(c.FinishReason),
			GenerationInfo: map[string]any{
				"CompletionTokens":  result.Usage.CompletionTokens,
				"PromptTokens":      result.Usage.PromptTokens,
				"TotalTokens":       result.Usage.TotalTokens,
				"ReasoningTokens":   result.Usage.CompletionTokensDetails.ReasoningTokens,
				"PromptAudioTokens": result.Usage.PromptTokensDetails.AudioTokens,
				// Standardized fields for cross-provider compatibility
				"ThinkingContent":                    c.Message.ReasoningContent,                           // Standardized field
				"ThinkingTokens":                     result.Usage.CompletionTokensDetails.ReasoningTokens, // Standardized field
				"PromptCachedTokens":                 result.Usage.PromptTokensDetails.CachedTokens,
				"CompletionAudioTokens":              result.Usage.CompletionTokensDetails.AudioTokens,
				"CompletionReasoningTokens":          result.Usage.CompletionTokensDetails.ReasoningTokens,
				"CompletionAcceptedPredictionTokens": result.Usage.CompletionTokensDetails.AcceptedPredictionTokens,
				"CompletionRejectedPredictionTokens": result.Usage.CompletionTokensDetails.RejectedPredictionTokens,
			},
		}

		// Legacy function call handling
		if c.FinishReason == "function_call" {
			choices[i].FuncCall = &llms.FunctionCall{
				Name:      c.Message.FunctionCall.Name,
				Arguments: c.Message.FunctionCall.Arguments,
			}
		}
		for _, tool := range c.Message.ToolCalls {
			choices[i].ToolCalls = append(choices[i].ToolCalls, llms.ToolCall{
				ID:   tool.ID,
				Type: string(tool.Type),
				FunctionCall: &llms.FunctionCall{
					Name:      tool.Function.Name,
					Arguments: tool.Function.Arguments,
				},
			})
		}
		// populate legacy single-function call field for backwards compatibility
		if len(choices[i].ToolCalls) > 0 {
			choices[i].FuncCall = choices[i].ToolCalls[0].FunctionCall
		}
	}
	response := &llms.ContentResponse{Choices: choices}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}
	return response, nil
}

// SupportsReasoning implements the ReasoningModel interface.
// Returns true if the current model supports reasoning/thinking tokens.
func (o *LLM) SupportsReasoning() bool {
	// Check the current model (may have been overridden by WithModel option)
	model := o.model
	if model == "" {
		model = o.client.Model
	}

	modelLower := strings.ToLower(model)

	// OpenAI o1 series (reasoning models)
	if strings.HasPrefix(modelLower, "o1-") ||
		strings.Contains(modelLower, "o1-preview") ||
		strings.Contains(modelLower, "o1-mini") {
		return true
	}

	// OpenAI o3 series
	if strings.HasPrefix(modelLower, "o3-") ||
		strings.Contains(modelLower, "o3-mini") {
		return true
	}

	// Future o4+ series
	if strings.HasPrefix(modelLower, "o4-") ||
		strings.HasPrefix(modelLower, "o5-") {
		return true
	}

	// GPT-5 series (expected to have reasoning capabilities)
	if strings.HasPrefix(modelLower, "gpt-5") {
		return true
	}

	return false
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
		Input: inputTexts,
		Model: o.client.EmbeddingModel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create openai embeddings: %w", err)
	}
	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
}

// ExtractToolParts extracts the tool parts from a message.
func ExtractToolParts(msg *ChatMessage) ([]llms.ContentPart, []llms.ToolCall) {
	var content []llms.ContentPart
	var toolCalls []llms.ToolCall
	for _, part := range msg.MultiContent {
		switch p := part.(type) {
		case llms.TextContent:
			content = append(content, p)
		case llms.ImageURLContent:
			content = append(content, p)
		case llms.BinaryContent:
			content = append(content, p)
		case llms.ToolCall:
			toolCalls = append(toolCalls, p)
		}
	}
	return content, toolCalls
}

// toolFromTool converts an llms.Tool to a Tool.
func toolFromTool(t llms.Tool) (openaiclient.Tool, error) {
	tool := openaiclient.Tool{
		Type: openaiclient.ToolType(t.Type),
	}
	switch t.Type {
	case string(openaiclient.ToolTypeFunction):
		tool.Function = openaiclient.FunctionDefinition{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  t.Function.Parameters,
			Strict:      t.Function.Strict,
		}
	default:
		return openaiclient.Tool{}, fmt.Errorf("tool type %v not supported", t.Type)
	}
	return tool, nil
}

// toolCallsFromToolCalls converts a slice of llms.ToolCall to a slice of ToolCall.
func toolCallsFromToolCalls(tcs []llms.ToolCall) []openaiclient.ToolCall {
	toolCalls := make([]openaiclient.ToolCall, len(tcs))
	for i, tc := range tcs {
		toolCalls[i] = toolCallFromToolCall(tc)
	}
	return toolCalls
}

// toolCallFromToolCall converts an llms.ToolCall to a ToolCall.
func toolCallFromToolCall(tc llms.ToolCall) openaiclient.ToolCall {
	return openaiclient.ToolCall{
		ID:   tc.ID,
		Type: openaiclient.ToolType(tc.Type),
		Function: openaiclient.ToolFunction{
			Name:      tc.FunctionCall.Name,
			Arguments: tc.FunctionCall.Arguments,
		},
	}
}
