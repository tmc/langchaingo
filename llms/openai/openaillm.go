package openai

import (
	"context"
	"fmt"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai/internal/openaiclient"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

type ChatMessage = openaiclient.ChatMessage

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *openaiclient.Client
}

const (
	RoleSystem    = "system"
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleFunction  = "function"
	RoleTool      = "tool"
)

var _ llms.Model = (*LLM)(nil)

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	opt, c, err := newClient(opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client:           c,
		CallbacksHandler: opt.callbackHandler,
	}, err
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// Create Text to Speech.
func (o *LLM) GenerateTTS(ctx context.Context, input string, options ...llms.CallOption) ([]byte, error) {
	if input == "" {
		return nil, fmt.Errorf("input is empty")
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	req := &openaiclient.TTSRequest{
		Input:          input,
		Model:          opts.GetModel(),
		Voice:          opts.GetVoice(),
		ResponseFormat: opts.GetResponseFormat(),
		Speed:          opts.GetSpeed(),
	}

	if req.Model != string(openaiclient.TTS1) && req.Model != string(openaiclient.TTS1HD) {
		req.Model = string(openaiclient.TTS1)
	}

	result, err := o.client.CreateTTS(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint:lll
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	chatMsgs, err := o.convertMessages(messages)
	if err != nil {
		return nil, err
	}

	req, err := o.createChatRequest(chatMsgs, opts)
	if err != nil {
		return nil, err
	}

	result, err := o.client.CreateChat(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	response := o.processResponse(result)

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}
	return response, nil
}

// convertMessages converts LangChain messages to OpenAI chat messages.
func (o *LLM) convertMessages(messages []llms.MessageContent) ([]*ChatMessage, error) {
	chatMsgs := make([]*ChatMessage, 0, len(messages))
	for _, mc := range messages {
		msg := &ChatMessage{MultiContent: mc.Parts}

		if err := o.setMessageRole(msg, mc); err != nil {
			return nil, err
		}

		newParts, toolCalls, toolCallResponses := ExtractToolParts(msg)
		msg.MultiContent = newParts
		msg.ToolCalls = toolCallsFromToolCalls(toolCalls)

		// Preserve reasoning content for multi-turn conversations with tool calls
		if o.client != nil && o.client.PreserveReasoningContent && msg.Role == RoleAssistant && len(toolCalls) > 0 {
			msg.ReasoningContent = extractReasoningContent(mc.Parts)
		}

		if len(msg.MultiContent) != 0 || len(msg.ToolCalls) != 0 {
			if msg.Role == RoleTool {
				msg.Role = RoleAssistant
			}
			chatMsgs = append(chatMsgs, msg)
		}

		for _, toolCallResponse := range toolCallResponses {
			chatMsgs = append(chatMsgs, &ChatMessage{
				Role:       RoleTool,
				Content:    toolCallResponse.Content,
				Name:       toolCallResponse.Name,
				ToolCallID: toolCallResponse.ToolCallID,
			})
		}
	}

	return chatMsgs, nil
}

// setMessageRole sets the appropriate role for a message and handles special cases.
func (o *LLM) setMessageRole(msg *ChatMessage, mc llms.MessageContent) error {
	switch mc.Role {
	case llms.ChatMessageTypeSystem:
		msg.Role = RoleSystem
	case llms.ChatMessageTypeAI:
		msg.Role = RoleAssistant
	case llms.ChatMessageTypeHuman:
		msg.Role = RoleUser
	case llms.ChatMessageTypeGeneric:
		msg.Role = RoleUser
	case llms.ChatMessageTypeFunction:
		msg.Role = RoleFunction
		return o.handleFunctionMessage(msg, mc)
	case llms.ChatMessageTypeTool:
		msg.Role = RoleTool
		return o.handleToolMessage(mc)
	default:
		return fmt.Errorf("role %v not supported", mc.Role)
	}
	return nil
}

// handleFunctionMessage handles function messages.
func (o *LLM) handleFunctionMessage(msg *ChatMessage, mc llms.MessageContent) error {
	if len(mc.Parts) != 1 {
		return fmt.Errorf("expected exactly one part for role %v, got %v", mc.Role, len(mc.Parts))
	}

	switch p := mc.Parts[0].(type) {
	case llms.ToolCallResponse:
		msg.ToolCallID = p.ToolCallID
		msg.Name = p.Name
		msg.Content = p.Content
	default:
		return fmt.Errorf("expected part of type ToolCallResponse for role %v, got %T",
			mc.Role, mc.Parts[0])
	}

	return nil
}

// handleToolMessage handles tool messages and returns complete tool response messages.
func (o *LLM) handleToolMessage(mc llms.MessageContent) error {
	for _, p := range mc.Parts {
		switch tr := p.(type) {
		case llms.ToolCallResponse:
			if tr.ToolCallID == "" || tr.Name == "" {
				return fmt.Errorf("tool call ID or name is empty for part %v", tr)
			}
		case llms.TextContent:
			// ignore text content, it should be handled on ExtractToolParts call
		default:
			return fmt.Errorf("expected part of type ToolCallResponse for role %v, got %T", mc.Role, tr)
		}
	}

	return nil
}

// createChatRequest creates an OpenAI chat request with the given parameters.
func (o *LLM) createChatRequest(chatMsgs []*ChatMessage, opts llms.CallOptions) (*openaiclient.ChatRequest, error) {
	req := &openaiclient.ChatRequest{
		Model:                opts.GetModel(),
		StopWords:            opts.StopWords,
		Messages:             chatMsgs,
		StreamingFunc:        opts.StreamingFunc,
		Temperature:          opts.Temperature,
		TopK:                 opts.TopK,
		TopP:                 opts.TopP,
		MinP:                 opts.MinP,
		N:                    opts.N,
		FrequencyPenalty:     opts.FrequencyPenalty,
		PresencePenalty:      opts.PresencePenalty,
		RepetitionPenalty:    opts.RepetitionPenalty,
		ToolChoice:           opts.ToolChoice,
		FunctionCallBehavior: openaiclient.FunctionCallBehavior(opts.FunctionCallBehavior),
		Seed:                 opts.Seed,
		Metadata:             opts.Metadata,
		WebSearchOptions:     webSearchOptionsFromCallOptions(opts.WebSearchOptions),
		ExtraBody:            getExtraBody(&opts),
	}

	if isLegacyMaxTokensField(&opts) {
		req.MaxTokens = opts.MaxTokens
	} else {
		req.MaxCompletionTokens = opts.MaxTokens
	}

	if opts.GetJSONMode() {
		req.ResponseFormat = ResponseFormatJSON
	}

	// set temperature to 1.0 for reasoning models
	if reasoning.IsReasoningModel(opts.GetModel()) {
		temperature := 1.0
		req.Temperature = &temperature
	}

	// add tools from functions and tool definitions
	if err := o.addToolsToRequest(req, opts); err != nil {
		return nil, err
	}

	// set response format from client if available
	if o.client.ResponseFormat != nil {
		req.ResponseFormat = o.client.ResponseFormat
	}

	// set reasoning options, depends on the client and request options
	o.setReasoning(req, opts)

	return req, nil
}

// setReasoning sets reasoning options, depends on the client and request options.
func (o *LLM) setReasoning(req *openaiclient.ChatRequest, opts llms.CallOptions) {
	if !opts.Reasoning.IsEnabled() {
		return
	}

	defer func() {
		if req.Reasoning != nil || req.ReasoningEffort != nil {
			// must of all reasoning models can't use temperature and top_p with reasoning at the same time
			req.Temperature, req.TopP = nil, nil
		}
	}()

	reasoningEffort := opts.Reasoning.GetEffort(opts.GetMaxTokens())
	reasoningTokens := opts.Reasoning.GetTokens(opts.GetMaxTokens())
	if !o.client.ModernReasoningFormat {
		if reasoningEffort != llms.ReasoningNone {
			req.ReasoningEffort = &reasoningEffort
		}
		return
	}

	// using modern reasoning format
	if o.client.UseReasoningMaxTokens && opts.Reasoning.Tokens != 0 && reasoningTokens != 0 {
		req.Reasoning = &openaiclient.ReasoningOptions{
			MaxTokens: reasoningTokens,
		}
	} else if reasoningEffort != llms.ReasoningNone {
		req.Reasoning = &openaiclient.ReasoningOptions{
			Effort: reasoningEffort,
		}
	}
}

// addToolsToRequest adds tools to the request from functions and tool definitions.
func (o *LLM) addToolsToRequest(req *openaiclient.ChatRequest, opts llms.CallOptions) error {
	// add function-based tools (deprecated approach)
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
			return fmt.Errorf("failed to convert llms tool to openai tool: %w", err)
		}
		req.Tools = append(req.Tools, t)
	}

	return nil
}

// processResponse processes the OpenAI API response into a ContentResponse.
func (o *LLM) processResponse(result *openaiclient.ChatCompletionResponse) *llms.ContentResponse {
	choices := make([]*llms.ContentChoice, len(result.Choices))

	for i, c := range result.Choices {
		choices[i] = &llms.ContentChoice{
			Content:        c.Message.Content,
			Reasoning:      o.processReasoning(c.Message.ReasoningContent),
			StopReason:     fmt.Sprint(c.FinishReason),
			GenerationInfo: o.processUsage(&result.Usage),
		}

		o.processToolCalls(choices[i], c)
	}

	return &llms.ContentResponse{Choices: choices}
}

func (o *LLM) processUsage(usage *openaiclient.ChatUsage) map[string]any {
	return map[string]any{
		"CompletionTokens":  usage.CompletionTokens,
		"PromptTokens":      usage.PromptTokens,
		"TotalTokens":       usage.TotalTokens,
		"ReasoningTokens":   usage.CompletionTokensDetails.ReasoningTokens,
		"PromptAudioTokens": usage.PromptTokensDetails.AudioTokens,
		// Standardized fields for cross-provider compatibility
		"PromptCachedTokens":                 usage.PromptTokensDetails.CachedTokens,
		"CacheReadInputTokens":               usage.PromptTokensDetails.CachedTokens,
		"CacheCreationInputTokens":           usage.PromptTokensDetails.CacheWriteTokens,
		"CompletionAudioTokens":              usage.CompletionTokensDetails.AudioTokens,
		"CompletionReasoningTokens":          usage.CompletionTokensDetails.ReasoningTokens,
		"CompletionAcceptedPredictionTokens": usage.CompletionTokensDetails.AcceptedPredictionTokens,
		"CompletionRejectedPredictionTokens": usage.CompletionTokensDetails.RejectedPredictionTokens,
		// Special fields for OpenRouter provider
		"UpstreamInferencePromptCost":      usage.CostDetails.UpstreamInferencePromptCost,
		"UpstreamInferenceCompletionsCost": usage.CostDetails.UpstreamInferenceCompletionsCost,
	}
}

// processReasoning processes reasoning content in the response.
func (o *LLM) processReasoning(reasoningContent string) *reasoning.ContentReasoning {
	if reasoningContent == "" {
		return nil
	}

	return &reasoning.ContentReasoning{
		Content:   reasoningContent,
		Signature: nil, // not supported yet for OpenAI compatible providers
	}
}

// processToolCalls processes tool calls in the response.
func (o *LLM) processToolCalls(choice *llms.ContentChoice, c *openaiclient.ChatCompletionChoice) {
	// legacy function call handling
	if c.FinishReason == "function_call" {
		choice.FuncCall = &llms.FunctionCall{
			Name:      c.Message.FunctionCall.Name,
			Arguments: c.Message.FunctionCall.Arguments,
		}
	}

	for _, tool := range c.Message.ToolCalls {
		choice.ToolCalls = append(choice.ToolCalls, llms.ToolCall{
			ID:   tool.ID,
			Type: string(tool.Type),
			FunctionCall: &llms.FunctionCall{
				Name:      tool.Function.Name,
				Arguments: tool.Function.Arguments,
			},
		})
	}

	// populate legacy single-function call field for backwards compatibility
	if len(choice.ToolCalls) > 0 {
		choice.FuncCall = choice.ToolCalls[0].FunctionCall
	}
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
func ExtractToolParts(msg *ChatMessage) ([]llms.ContentPart, []llms.ToolCall, []llms.ToolCallResponse) {
	var content []llms.ContentPart
	var toolCalls []llms.ToolCall
	var toolCallResponses []llms.ToolCallResponse
	for _, part := range msg.MultiContent {
		switch p := part.(type) {
		case llms.ToolCall:
			toolCalls = append(toolCalls, p)
		case llms.ToolCallResponse:
			toolCallResponses = append(toolCallResponses, p)
		case llms.TextContent:
			if p.Text == "" {
				continue
			}
			content = append(content, p)
		case llms.ImageURLContent, llms.BinaryContent:
			content = append(content, p)
		default:
			// ignore other parts
		}
	}
	return content, toolCalls, toolCallResponses
}

// extractReasoningContent extracts reasoning content from message parts.
// It returns the first non-empty reasoning content found in TextContent parts.
func extractReasoningContent(parts []llms.ContentPart) string {
	for _, part := range parts {
		if tc, ok := part.(llms.TextContent); ok {
			if tc.Reasoning != nil && tc.Reasoning.Content != "" {
				return tc.Reasoning.Content
			}
		}
	}
	return ""
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
	for idx, tc := range tcs {
		toolCalls[idx] = toolCallFromToolCall(tc)
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

// webSearchOptionsFromCallOptions converts llms.WebSearchOptions to openaiclient.WebSearchOptions.
func webSearchOptionsFromCallOptions(opts *llms.WebSearchOptions) *openaiclient.WebSearchOptions {
	if opts == nil {
		return nil
	}
	result := &openaiclient.WebSearchOptions{
		SearchContextSize: opts.SearchContextSize,
	}
	if opts.UserLocation != nil {
		result.UserLocation = &openaiclient.UserLocation{
			Type: opts.UserLocation.Type,
		}
		if opts.UserLocation.Approximate != nil {
			result.UserLocation.Approximate = &openaiclient.ApproximateLocation{
				Country: opts.UserLocation.Approximate.Country,
				City:    opts.UserLocation.Approximate.City,
				Region:  opts.UserLocation.Approximate.Region,
			}
		}
	}
	return result
}
