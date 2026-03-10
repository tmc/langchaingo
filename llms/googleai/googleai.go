//nolint:all
package googleai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/vxcontrol/langchaingo/internal/imageutil"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"google.golang.org/genai"
)

var (
	ErrNoContentInResponse   = errors.New("no content in generation response")
	ErrUnknownPartInResponse = errors.New("unknown part type in generation response")
	ErrInvalidMimeType       = errors.New("invalid mime type on content")
)

const (
	CITATIONS                         = "citations"
	SAFETY                            = "safety"
	RoleSystem                        = "system"
	RoleModel                         = "model"
	RoleUser                          = "user"
	RoleTool                          = "tool"
	ResponseMIMETypeJson              = "application/json"
	GENERATED_FUNCTION_CALL_ID_PREFIX = "fcall_"
)

// ensureFunctionCallID generates a unique ID if the provided ID is empty.
// Generated IDs use the format "fcall_{16_hex_chars}" to distinguish them from backend-provided IDs.
func ensureFunctionCallID(id string) string {
	if id != "" {
		return id
	}

	// Generate 8 random bytes = 16 hex characters
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to deterministic ID if random generation fails
		return GENERATED_FUNCTION_CALL_ID_PREFIX + "00000000"
	}

	return GENERATED_FUNCTION_CALL_ID_PREFIX + hex.EncodeToString(bytes)
}

// cleanFunctionCallID removes generated ID prefix when sending to LLM backend.
// If ID has the generated prefix, returns empty string; otherwise returns the original ID.
func cleanFunctionCallID(id string) string {
	if strings.HasPrefix(id, GENERATED_FUNCTION_CALL_ID_PREFIX) {
		return ""
	}
	return id
}

// extractThoughtSignature extracts the thought signature from ContentReasoning if present.
func extractThoughtSignature(r *reasoning.ContentReasoning) []byte {
	if r == nil {
		return nil
	}
	return r.Signature
}

// Call implements the [llms.Model] interface.
func (g *GoogleAI) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, g, prompt, options...)
}

// GenerateContent implements the [llms.Model] interface.
func (g *GoogleAI) GenerateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{
		Model:          getStringPointer(g.opts.DefaultModel),
		CandidateCount: getIntPointer(g.opts.DefaultCandidateCount),
		MaxTokens:      getIntPointer(g.opts.DefaultMaxTokens),
		Temperature:    getFloatPointer(g.opts.DefaultTemperature),
		TopP:           getFloatPointer(g.opts.DefaultTopP),
		TopK:           getIntPointer(g.opts.DefaultTopK),
	}
	for _, opt := range options {
		opt(&opts)
	}

	// Build generation config
	temperature := convertToFloat32Pointer(opts.Temperature)
	topP := convertToFloat32Pointer(opts.TopP)
	topK := convertIntToFloat32Pointer(opts.TopK)

	config := &genai.GenerateContentConfig{
		CandidateCount:  convertToInt32(opts.CandidateCount),
		MaxOutputTokens: convertToInt32(opts.MaxTokens),
		Temperature:     temperature,
		TopP:            topP,
		TopK:            topK,
		StopSequences:   opts.StopWords,
	}

	// Check for cached content
	if opts.Metadata != nil {
		if cachedContentName, ok := opts.Metadata["CachedContentName"].(string); ok && cachedContentName != "" {
			config.CachedContent = cachedContentName
		}
	}

	// Handle response MIME type and JSON mode
	switch {
	case opts.ResponseMIMEType != nil && opts.JSONMode:
		return nil, fmt.Errorf("conflicting options, can't use JSONMode and ResponseMIMEType together")
	case opts.ResponseMIMEType != nil && !opts.JSONMode:
		config.ResponseMIMEType = opts.GetResponseMIMEType()
	case opts.GetResponseMIMEType() == "" && opts.JSONMode:
		config.ResponseMIMEType = ResponseMIMETypeJson
	}

	// Handle thinking configuration for reasoning models
	if opts.Reasoning != nil && opts.Reasoning.IsEnabled() {
		thinkingBudget := int32(opts.Reasoning.GetTokens(opts.GetMaxTokens()))
		if thinkingBudget > 0 {
			config.ThinkingConfig = &genai.ThinkingConfig{
				ThinkingBudget:  &thinkingBudget,
				IncludeThoughts: true, // Include thought summaries by default
			}
		}
	}

	// Convert tools
	if len(opts.Tools) > 0 {
		tools, err := convertTools(opts.Tools)
		if err != nil {
			return nil, err
		}
		config.Tools = tools
	}

	// Add safety settings
	config.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: convertHarmBlockThreshold(g.opts.HarmThreshold),
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: convertHarmBlockThreshold(g.opts.HarmThreshold),
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: convertHarmBlockThreshold(g.opts.HarmThreshold),
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: convertHarmBlockThreshold(g.opts.HarmThreshold),
		},
	}

	var response *llms.ContentResponse
	var err error

	if len(messages) == 1 {
		theMessage := messages[0]
		if theMessage.Role != llms.ChatMessageTypeHuman {
			return nil, fmt.Errorf("got %v message role, want human", theMessage.Role)
		}
		response, err = g.generateFromSingleMessage(ctx, opts.GetModel(), theMessage.Parts, config, &opts)
	} else {
		response, err = g.generateFromMessages(ctx, opts.GetModel(), messages, config, &opts)
	}
	if err != nil {
		return nil, err
	}

	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

func (g *GoogleAI) generateFromSingleMessage(
	ctx context.Context,
	model string,
	parts []llms.ContentPart,
	config *genai.GenerateContentConfig,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	convertedParts, err := convertParts(parts)
	if err != nil {
		return nil, err
	}

	content := []*genai.Content{{
		Parts: convertedParts,
		Role:  RoleUser,
	}}

	if opts.StreamingFunc == nil {
		resp, err := g.client.Models.GenerateContent(ctx, model, content, config)
		if err != nil {
			return nil, err
		}
		return convertResponse(resp)
	}

	return g.generateStreamingContent(ctx, model, content, config, opts)
}

func (g *GoogleAI) generateFromMessages(
	ctx context.Context,
	model string,
	messages []llms.MessageContent,
	config *genai.GenerateContentConfig,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	var systemInstruction *genai.Content
	var contents []*genai.Content

	for _, msg := range messages {
		content, err := convertContent(msg)
		if err != nil {
			return nil, err
		}

		if msg.Role == llms.ChatMessageTypeSystem {
			systemInstruction = content
		} else {
			contents = append(contents, content)
		}
	}

	if systemInstruction != nil {
		config.SystemInstruction = systemInstruction
	}

	if opts.StreamingFunc == nil {
		resp, err := g.client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			return nil, err
		}
		return convertResponse(resp)
	}

	return g.generateStreamingContent(ctx, model, contents, config, opts)
}

func (g *GoogleAI) generateStreamingContent(
	ctx context.Context,
	model string,
	contents []*genai.Content,
	config *genai.GenerateContentConfig,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	iter := g.client.Models.GenerateContentStream(ctx, model, contents, config)

	defer streaming.CallWithDone(ctx, opts.StreamingFunc)

	var accumulatedContent strings.Builder
	var accumulatedReasoningContent strings.Builder
	var accumulatedToolCalls []llms.ToolCall
	var thoughtSignature []byte
	var lastUsageMetadata *genai.GenerateContentResponseUsageMetadata
	var lastCandidate *genai.Candidate

	// Trying to keep the same ID for the same tool call name
	toolCallIDs := make(map[string]string)
	ensureStreamFunctionCallID := func(name, id string) string {
		if rid, ok := toolCallIDs[name]; id == "" && ok {
			return rid
		}
		toolCallIDs[name] = ensureFunctionCallID(id)
		return toolCallIDs[name]
	}

	for chunk, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("error generating content: %w", err)
		}
		if chunk == nil {
			return nil, fmt.Errorf("unexpected case: chunk is nil")
		}

		// Capture usage metadata from each chunk (last one will be the final)
		if chunk.UsageMetadata != nil {
			lastUsageMetadata = chunk.UsageMetadata
		}

		if len(chunk.Candidates) == 0 {
			continue
		}

		candidate := chunk.Candidates[0]
		lastCandidate = candidate
		if candidate.Content == nil {
			continue
		}

		for _, part := range candidate.Content.Parts {
			// Accumulate thought signature from any part
			if len(part.ThoughtSignature) > 0 {
				thoughtSignature = part.ThoughtSignature
			}

			if len(part.Text) > 0 {
				if part.Thought {
					accumulatedReasoningContent.WriteString(part.Text)
					chunk := streaming.Chunk{
						Type: streaming.ChunkTypeReasoning,
						Reasoning: &reasoning.ContentReasoning{
							Content:   part.Text,
							Signature: thoughtSignature, // Include signature if available
						},
					}
					if err := opts.StreamingFunc(ctx, chunk); err != nil {
						goto StreamEnd
					}
				} else {
					accumulatedContent.WriteString(part.Text)
					if err := streaming.CallWithText(ctx, opts.StreamingFunc, part.Text); err != nil {
						goto StreamEnd
					}
				}
			}

			if part.FunctionCall != nil {
				b, _ := json.Marshal(part.FunctionCall.Args)
				toolCall := llms.ToolCall{
					ID: ensureStreamFunctionCallID(part.FunctionCall.Name, part.FunctionCall.ID),
					FunctionCall: &llms.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(b),
					},
				}
				accumulatedToolCalls = append(accumulatedToolCalls, toolCall)
			}
		}
	}

StreamEnd:
	// Distribute Reasoning according to the rules:
	// - If there are ToolCalls, attach Reasoning ONLY to the first one
	// - If there are no ToolCalls, attach Reasoning to ContentChoice
	var choiceReasoning *reasoning.ContentReasoning
	if len(accumulatedToolCalls) > 0 {
		// Add reasoning to the first tool call
		if accumulatedReasoningContent.Len() > 0 || len(thoughtSignature) > 0 {
			accumulatedToolCalls[0].Reasoning = &reasoning.ContentReasoning{
				Content:   accumulatedReasoningContent.String(),
				Signature: thoughtSignature,
			}
		}
	} else {
		// Add reasoning to the content choice
		if accumulatedReasoningContent.Len() > 0 || len(thoughtSignature) > 0 {
			choiceReasoning = &reasoning.ContentReasoning{
				Content:   accumulatedReasoningContent.String(),
				Signature: thoughtSignature,
			}
		}
	}

	// Build metadata from accumulated usage information
	metadata := make(map[string]any)
	if lastCandidate != nil {
		metadata[CITATIONS] = lastCandidate.CitationMetadata
		metadata[SAFETY] = lastCandidate.SafetyRatings
	}

	if lastUsageMetadata != nil {
		metadata["input_tokens"] = int(lastUsageMetadata.PromptTokenCount)
		metadata["output_tokens"] = int(lastUsageMetadata.CandidatesTokenCount)
		metadata["total_tokens"] = int(lastUsageMetadata.TotalTokenCount)

		// Standardized field names for cross-provider compatibility
		metadata["PromptTokens"] = int(lastUsageMetadata.PromptTokenCount)
		metadata["CompletionTokens"] = int(lastUsageMetadata.CandidatesTokenCount)
		metadata["TotalTokens"] = int(lastUsageMetadata.TotalTokenCount)
		metadata["ReasoningTokens"] = int(lastUsageMetadata.ThoughtsTokenCount)
		metadata["PromptCachedTokens"] = int(lastUsageMetadata.CachedContentTokenCount)
		metadata["CacheReadInputTokens"] = int(lastUsageMetadata.CachedContentTokenCount)

		// Cache-related token information (if available)
		if lastUsageMetadata.CachedContentTokenCount > 0 {
			metadata["CacheCreationInputTokens"] = max(int(lastUsageMetadata.PromptTokenCount-lastUsageMetadata.CachedContentTokenCount), 0)
			metadata["PromptTokens"] = metadata["CacheCreationInputTokens"] // Google AI includes cached tokens in the prompt count
		} else {
			metadata["CacheCreationInputTokens"] = 0
		}
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content:        accumulatedContent.String(),
			Reasoning:      choiceReasoning,
			ToolCalls:      accumulatedToolCalls,
			GenerationInfo: metadata,
		}},
	}, nil
}

func convertResponse(resp *genai.GenerateContentResponse) (*llms.ContentResponse, error) {
	if len(resp.Candidates) == 0 {
		return nil, ErrNoContentInResponse
	}

	var choices []*llms.ContentChoice

	for _, candidate := range resp.Candidates {
		var content, thinking strings.Builder
		var toolCalls []llms.ToolCall
		var thoughtSignature []byte
		var firstFunctionCallIdx = -1

		if candidate.Content != nil {
			for idx, part := range candidate.Content.Parts {
				if len(part.Text) > 0 {
					// Check if this is thinking content
					if part.Thought {
						thinking.WriteString(part.Text)
					} else {
						content.WriteString(part.Text)
					}
				}

				// Collect thought signature from any part (typically from first function call or last text part)
				if len(part.ThoughtSignature) > 0 {
					thoughtSignature = part.ThoughtSignature
				}

				if part.FunctionCall != nil {
					b, err := json.Marshal(part.FunctionCall.Args)
					if err != nil {
						return nil, err
					}
					toolCall := llms.ToolCall{
						ID: ensureFunctionCallID(part.FunctionCall.ID),
						FunctionCall: &llms.FunctionCall{
							Name:      part.FunctionCall.Name,
							Arguments: string(b),
						},
					}

					// Track index of first function call
					if firstFunctionCallIdx == -1 {
						firstFunctionCallIdx = idx
					}

					toolCalls = append(toolCalls, toolCall)
				}
			}
		}

		// Distribute Reasoning according to the rules:
		// - If there are ToolCalls, attach Reasoning ONLY to the first one
		// - If there are no ToolCalls, attach Reasoning to ContentChoice
		var choiceReasoning *reasoning.ContentReasoning
		if len(toolCalls) > 0 {
			// Add reasoning to the first tool call
			if thinking.Len() > 0 || len(thoughtSignature) > 0 {
				toolCalls[0].Reasoning = &reasoning.ContentReasoning{
					Content:   thinking.String(),
					Signature: thoughtSignature,
				}
			}
		} else {
			// Add reasoning to the content choice
			if thinking.Len() > 0 || len(thoughtSignature) > 0 {
				choiceReasoning = &reasoning.ContentReasoning{
					Content:   thinking.String(),
					Signature: thoughtSignature,
				}
			}
		}

		metadata := make(map[string]any)
		metadata[CITATIONS] = candidate.CitationMetadata
		metadata[SAFETY] = candidate.SafetyRatings

		if usage := resp.UsageMetadata; usage != nil {
			metadata["input_tokens"] = usage.PromptTokenCount
			metadata["output_tokens"] = usage.CandidatesTokenCount
			metadata["total_tokens"] = usage.TotalTokenCount

			// Standardized field names for cross-provider compatibility
			metadata["PromptTokens"] = int(usage.PromptTokenCount)
			metadata["CompletionTokens"] = int(usage.CandidatesTokenCount)
			metadata["TotalTokens"] = int(usage.TotalTokenCount)
			metadata["ReasoningTokens"] = int(usage.ThoughtsTokenCount)
			metadata["PromptCachedTokens"] = int(usage.CachedContentTokenCount)
			metadata["CacheReadInputTokens"] = int(usage.CachedContentTokenCount)
			// Google AI does not provide cache creation information, always 0
			metadata["CacheCreationInputTokens"] = 0
		}

		choices = append(choices, &llms.ContentChoice{
			Content:        content.String(),
			Reasoning:      choiceReasoning,
			StopReason:     string(candidate.FinishReason),
			GenerationInfo: metadata,
			ToolCalls:      toolCalls,
		})
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func convertParts(parts []llms.ContentPart) ([]*genai.Part, error) {
	convertedParts := make([]*genai.Part, 0, len(parts))

	// Pre-scan to detect signature placement strategy:
	// If we have tool calls with signatures, we should NOT add signature to empty TextContent
	// This prevents duplicate signatures in a single message.
	hasToolCallWithSignature := false
	for _, part := range parts {
		if tc, ok := part.(llms.ToolCall); ok {
			if tc.Reasoning != nil && len(tc.Reasoning.Signature) > 0 {
				hasToolCallWithSignature = true
				break
			}
		}
	}

	for _, part := range parts {
		var genaiPart *genai.Part

		switch p := part.(type) {
		case llms.TextContent:
			// Skip completely empty text parts without reasoning.
			// Empty parts serve no purpose and cause Gemini API errors:
			// "required oneof field 'data' must have one initialized field"
			if p.Text == "" && (p.Reasoning == nil || p.Reasoning.IsEmpty()) {
				continue
			}

			// Optimization: skip empty text parts that only carry signature
			// when we already have tool call with signature.
			// This prevents duplicate signatures in the same message.
			if p.Text == "" && hasToolCallWithSignature && p.Reasoning != nil {
				// Skip this part - signature will be in tool call
				continue
			}

			genaiPart = &genai.Part{
				Text:             p.Text,
				ThoughtSignature: extractThoughtSignature(p.Reasoning),
			}

		case llms.BinaryContent:
			genaiPart = &genai.Part{
				InlineData: &genai.Blob{
					MIMEType: p.MIMEType,
					Data:     p.Data,
				},
			}

		case llms.ImageURLContent:
			typ, data, err := imageutil.DownloadImageData(p.URL)
			if err != nil {
				return nil, err
			}
			genaiPart = &genai.Part{
				InlineData: &genai.Blob{
					MIMEType: typ,
					Data:     data,
				},
			}

		case llms.ToolCall:
			fc := p.FunctionCall
			var argsMap map[string]any
			if err := json.Unmarshal([]byte(fc.Arguments), &argsMap); err != nil {
				return nil, err
			}
			genaiPart = &genai.Part{
				FunctionCall: &genai.FunctionCall{
					ID:   cleanFunctionCallID(p.ID),
					Name: fc.Name,
					Args: argsMap,
				},
				ThoughtSignature: extractThoughtSignature(p.Reasoning),
			}

		case llms.ToolCallResponse:
			genaiPart = &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:   cleanFunctionCallID(p.ToolCallID),
					Name: p.Name,
					Response: map[string]any{
						"response": p.Content,
					},
				},
			}

		default:
			return nil, fmt.Errorf("unsupported content part type: %T", part)
		}

		convertedParts = append(convertedParts, genaiPart)
	}

	return convertedParts, nil
}

func convertContent(content llms.MessageContent) (*genai.Content, error) {
	parts, err := convertParts(content.Parts)
	if err != nil {
		return nil, err
	}

	var role string
	switch content.Role {
	case llms.ChatMessageTypeSystem:
		role = RoleSystem
	case llms.ChatMessageTypeAI:
		role = RoleModel
	case llms.ChatMessageTypeHuman:
		role = RoleUser
	case llms.ChatMessageTypeGeneric:
		role = RoleUser
	case llms.ChatMessageTypeTool:
		role = RoleUser
	case llms.ChatMessageTypeFunction:
		fallthrough
	default:
		return nil, fmt.Errorf("role %v not supported", content.Role)
	}

	return &genai.Content{
		Parts: parts,
		Role:  role,
	}, nil
}

func convertTools(tools []llms.Tool) ([]*genai.Tool, error) {
	if len(tools) == 0 {
		return nil, nil
	}

	var functionDeclarations []*genai.FunctionDeclaration

	for i, tool := range tools {
		if tool.Type != "function" {
			return nil, fmt.Errorf("tool [%d]: unsupported type %q, want 'function'", i, tool.Type)
		}

		genaiFuncDecl := &genai.FunctionDeclaration{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
		}

		schema, err := convertToSchema(tool.Function.Parameters, true, i, "")
		if err != nil {
			return nil, err
		}
		genaiFuncDecl.Parameters = schema

		functionDeclarations = append(functionDeclarations, genaiFuncDecl)
	}

	return []*genai.Tool{{
		FunctionDeclarations: functionDeclarations,
	}}, nil
}

func convertMaps(i any) (any, error) {
	var err error
	switch v := i.(type) {
	case map[any]any:
		m := make(map[string]any)
		for key, val := range v {
			sKey, ok := key.(string)
			if !ok {
				return v, nil
			}
			m[sKey], err = convertMaps(val)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case []any:
		s := make([]any, len(v))
		for idx, val := range v {
			s[idx], err = convertMaps(val)
			if err != nil {
				s[idx] = val
			}
		}
		return s, nil
	default:
		d, err := json.Marshal(i)
		if err != nil {
			return i, err
		}
		var m any
		if err := json.Unmarshal(d, &m); err != nil {
			return i, err
		}
		return m, nil
	}
}

func convertToSchema(e any, topLevel bool, toolIndex int, propertyPath string) (*genai.Schema, error) {
	e, err := convertMaps(e)
	if err != nil {
		return nil, err
	}
	schema := &genai.Schema{}

	eMap, ok := e.(map[string]any)
	if !ok {
		if propertyPath != "" {
			return nil, fmt.Errorf("tool [%d], property [%s]: unsupported type %T of Parameters", toolIndex, propertyPath, e)
		}
		return nil, fmt.Errorf("tool [%d]: unsupported type %T of Parameters", toolIndex, e)
	}

	if ty, ok := eMap["type"]; ok {
		tyString, ok := ty.(string)
		if !ok {
			if propertyPath != "" {
				return nil, fmt.Errorf("tool [%d], property [%s]: expected string for type", toolIndex, propertyPath)
			}
			return nil, fmt.Errorf("tool [%d]: expected string for type", toolIndex)
		}
		schema.Type = convertToolSchemaType(tyString)

		if topLevel && schema.Type != genai.TypeObject {
			return nil, fmt.Errorf("tool [%d]: top-level schema must be an object", toolIndex)
		}
	}

	if properties, ok := eMap["properties"]; ok {
		paramProperties, ok := properties.(map[string]any)
		if !ok {
			if propertyPath != "" {
				return nil, fmt.Errorf("tool [%d], property [%s]: expected map[string]any for properties", toolIndex, propertyPath)
			}
			return nil, fmt.Errorf("tool [%d]: expected map[string]any for properties", toolIndex)
		}
		schema.Properties = make(map[string]*genai.Schema)
		for propName, propValue := range paramProperties {
			// Build nested path for better error messages
			nestedPath := propName
			if propertyPath != "" {
				nestedPath = propertyPath + "." + propName
			}

			recSchema, err := convertToSchema(propValue, false, toolIndex, nestedPath)
			if err != nil {
				return nil, err
			}
			schema.Properties[propName] = recSchema
		}
	} else if schema.Type == genai.TypeObject && propertyPath == "" {
		// For top-level object schemas without properties, this is an error
		return nil, fmt.Errorf("tool [%d]: expected to find a map of properties", toolIndex)
	}

	if items, ok := eMap["items"]; ok {
		if schema.Type == genai.TypeArray {
			// Build items path for better error messages
			itemsPath := propertyPath + "[]"
			itemsSchema, err := convertToSchema(items, false, toolIndex, itemsPath)
			if err != nil {
				return nil, err
			}
			schema.Items = itemsSchema
		} else {
			// items field present but type is not array
			itemsSchema, err := convertToSchema(items, false, toolIndex, propertyPath)
			if err != nil {
				return nil, err
			}
			schema.Items = itemsSchema
		}
	} else if schema.Type == genai.TypeArray {
		if propertyPath != "" {
			return nil, fmt.Errorf("tool [%d], property [%s]: array schema must have items", toolIndex, propertyPath)
		}
		return nil, fmt.Errorf("tool [%d]: array schema must have items", toolIndex)
	}

	if description, ok := eMap["description"]; ok {
		descString, ok := description.(string)
		if !ok {
			if propertyPath != "" {
				return nil, fmt.Errorf("tool [%d], property [%s]: expected string for description", toolIndex, propertyPath)
			}
			return nil, fmt.Errorf("tool [%d]: expected string for description", toolIndex)
		}
		schema.Description = descString
	}

	if nullable, ok := eMap["nullable"]; ok {
		nullableBool, ok := nullable.(bool)
		if !ok {
			if propertyPath != "" {
				return nil, fmt.Errorf("tool [%d], property [%s]: expected bool for nullable", toolIndex, propertyPath)
			}
			return nil, fmt.Errorf("tool [%d]: expected bool for nullable", toolIndex)
		}
		schema.Nullable = &nullableBool
	}

	if enum, ok := eMap["enum"]; ok {
		enumSlice, err := convertToSliceOfStrings(enum, toolIndex, propertyPath)
		if err != nil {
			return nil, err
		}
		schema.Enum = enumSlice
	}

	if required, ok := eMap["required"]; ok {
		requiredSlice, err := convertToSliceOfStrings(required, toolIndex, propertyPath)
		if err != nil {
			return nil, err
		}
		schema.Required = requiredSlice
	}

	return schema, nil
}

func convertToSliceOfStrings(e any, toolIndex int, propertyPath string) ([]string, error) {
	if rs, ok := e.([]string); ok {
		return rs, nil
	}

	ri, ok := e.([]interface{})
	if !ok {
		if propertyPath != "" {
			return nil, fmt.Errorf("tool [%d], property [%s]: expected []string or []interface{} for slice", toolIndex, propertyPath)
		}
		return nil, fmt.Errorf("tool [%d]: expected []string or []interface{} for slice", toolIndex)
	}
	rs := make([]string, 0, len(ri))
	for _, r := range ri {
		rString, ok := r.(string)
		if !ok {
			if propertyPath != "" {
				return nil, fmt.Errorf("tool [%d], property [%s]: expected string element in slice", toolIndex, propertyPath)
			}
			return nil, fmt.Errorf("tool [%d]: expected string element in slice", toolIndex)
		}
		rs = append(rs, rString)
	}
	return rs, nil
}

func convertToolSchemaType(ty string) genai.Type {
	switch ty {
	case "object":
		return genai.TypeObject
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	default:
		return genai.TypeUnspecified
	}
}

func showContent(w io.Writer, cs []*genai.Content) {
	fmt.Fprintf(w, "Content (len=%v)\n", len(cs))
	for i, c := range cs {
		fmt.Fprintf(w, "[%d]: Role=%s\n", i, c.Role)
		for j, p := range c.Parts {
			fmt.Fprintf(w, "  Parts[%v]: ", j)
			switch {
			case len(p.Text) > 0:
				fmt.Fprintf(w, "Text %q\n", p.Text)
			case p.InlineData != nil:
				fmt.Fprintf(w, "Blob MIME=%q, size=%d\n", p.InlineData.MIMEType, len(p.InlineData.Data))
			case p.FunctionCall != nil:
				fmt.Fprintf(w, "FunctionCall ID=%v Name=%v, Args=%v\n",
					p.FunctionCall.ID, p.FunctionCall.Name, p.FunctionCall.Args)
			case p.FunctionResponse != nil:
				fmt.Fprintf(w, "FunctionResponse ID=%v Name=%v Response=%v\n",
					p.FunctionResponse.ID, p.FunctionResponse.Name, p.FunctionResponse.Response)
			default:
				fmt.Fprintf(w, "unknown part type\n")
			}
		}
	}
}

func convertHarmBlockThreshold(threshold HarmBlockThreshold) genai.HarmBlockThreshold {
	switch threshold {
	case HarmBlockUnspecified:
		return genai.HarmBlockThresholdUnspecified
	case HarmBlockLowAndAbove:
		return genai.HarmBlockThresholdBlockLowAndAbove
	case HarmBlockMediumAndAbove:
		return genai.HarmBlockThresholdBlockMediumAndAbove
	case HarmBlockOnlyHigh:
		return genai.HarmBlockThresholdBlockOnlyHigh
	case HarmBlockNone:
		return genai.HarmBlockThresholdBlockNone
	default:
		return genai.HarmBlockThresholdBlockOnlyHigh // Safe default
	}
}

func getStringPointer(s string) *string {
	return &s
}

func getIntPointer(i int) *int {
	return &i
}

func getFloatPointer(f float64) *float64 {
	return &f
}

func convertToFloat32Pointer(f *float64) *float32 {
	if f == nil {
		return nil
	}

	f32 := float32(*f)
	return &f32
}

func convertToInt32Pointer(i *int) *int32 {
	if i == nil {
		return nil
	}

	i32 := int32(*i)
	return &i32
}

func convertToInt32(i *int) int32 {
	if i == nil {
		return 0
	}

	return int32(*i)
}

func convertIntToFloat32Pointer(i *int) *float32 {
	if i == nil {
		return nil
	}

	f32 := float32(*i)
	return &f32
}
