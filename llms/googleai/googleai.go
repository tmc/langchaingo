//nolint:all
package googleai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/vendasta/langchaingo/internal/imageutil"
	"github.com/vendasta/langchaingo/llms"
	googleaierrors "github.com/vendasta/langchaingo/llms/googleai/errors"
	"google.golang.org/genai"
)

var (
	ErrNoContentInResponse   = errors.New("no content in generation response")
	ErrUnknownPartInResponse = errors.New("unknown part type in generation response")
	ErrInvalidMimeType       = errors.New("invalid mime type on content")
)

const (
	CITATIONS            = "citations"
	SAFETY               = "safety"
	RoleSystem           = "system"
	RoleModel            = "model"
	RoleUser             = "user"
	RoleTool             = "tool"
	ResponseMIMETypeJson = "application/json"
)

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
		Model:          g.opts.DefaultModel,
		CandidateCount: g.opts.DefaultCandidateCount,
		MaxTokens:      g.opts.DefaultMaxTokens,
		Temperature:    g.opts.DefaultTemperature,
		TopP:           g.opts.DefaultTopP,
		TopK:           g.opts.DefaultTopK,
	}
	for _, opt := range options {
		opt(&opts)
	}

	// Update the tracked model if it was overridden
	effectiveModel := opts.Model
	if effectiveModel != "" && effectiveModel != g.model {
		g.model = effectiveModel
	}

	var response *llms.ContentResponse

	var err error
	if len(messages) == 1 {
		theMessage := messages[0]
		if theMessage.Role != llms.ChatMessageTypeHuman {
			return nil, fmt.Errorf("got %v message role, want human", theMessage.Role)
		}
		response, err = generateFromSingleMessage(ctx, g.client, effectiveModel, theMessage.Parts, &opts, g.opts)
	} else {
		response, err = generateFromMessages(ctx, g.client, effectiveModel, messages, &opts, g.opts)
	}
	if err != nil {
		return nil, err
	}

	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

// buildGenerateContentConfig builds a GenerateContentConfig from CallOptions and client Options.
func buildGenerateContentConfig(opts *llms.CallOptions, clientOpts Options) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}

	// Set generation parameters
	if opts.CandidateCount > 0 {
		count := int32(opts.CandidateCount)
		config.CandidateCount = count
	}
	// Set MaxOutputTokens - this limits only the output/response tokens,
	// not thinking tokens. Thinking tokens are controlled separately via ThinkingBudget.
	if opts.MaxTokens > 0 {
		tokens := int32(opts.MaxTokens)
		config.MaxOutputTokens = tokens
	}
	if opts.Temperature > 0 {
		temp := float32(opts.Temperature)
		config.Temperature = &temp
	}
	if opts.TopP > 0 {
		topP := float32(opts.TopP)
		config.TopP = &topP
	}
	if opts.TopK > 0 {
		topK := float32(opts.TopK)
		config.TopK = &topK
	}
	if len(opts.StopWords) > 0 {
		config.StopSequences = opts.StopWords
	}

	// Set response MIME type
	switch {
	case opts.ResponseMIMEType != "" && opts.JSONMode:
		// Error handled in GenerateContent
	case opts.ResponseMIMEType != "" && !opts.JSONMode:
		config.ResponseMIMEType = opts.ResponseMIMEType
	case opts.ResponseMIMEType == "" && opts.JSONMode:
		config.ResponseMIMEType = ResponseMIMETypeJson
	}

	// Set safety settings
	// Convert our HarmBlockThreshold (int32) to the new SDK's HarmBlockThreshold (string)
	// The new SDK uses string-based enum constants, not int32 values
	threshold := convertHarmBlockThreshold(clientOpts.HarmThreshold)

	config.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: threshold,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: threshold,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: threshold,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: threshold,
		},
	}

	// Convert tool config and tools
	// When ToolChoice is "none", don't include tools or ToolConfig
	// as the behavior should be "same as when not passing any function declarations"
	if opts.ToolChoice != nil {
		// Check if ToolChoice is "none" - if so, skip both tools and ToolConfig
		if tc, ok := opts.ToolChoice.(string); ok && strings.ToLower(strings.TrimSpace(tc)) == "none" {
			// Explicitly ensure Tools and ToolConfig are not set when ToolChoice is "none"
			// This matches the API behavior: "same as when not passing any function declarations"
			config.Tools = nil
			config.ToolConfig = nil
		} else {
			// Set ToolConfig for other ToolChoice values
			config.ToolConfig = convertToolConfig(opts.ToolChoice)

			// Convert tools (only if ToolChoice is not "none")
			if len(opts.Tools) > 0 {
				if tools, err := convertTools(opts.Tools); err == nil && tools != nil {
					config.Tools = tools
				}
			}
		}
	} else {
		// No ToolChoice specified, include tools if provided
		if len(opts.Tools) > 0 {
			if tools, err := convertTools(opts.Tools); err == nil && tools != nil {
				config.Tools = tools
			}
		}
	}

	// Support for cached content
	// TODO: Update when new SDK supports cached content in GenerateContentConfig
	// For now, cached content support may need to be handled differently
	_ = opts.Metadata["CachedContentName"] // Placeholder for future implementation

	// Extract and set ThinkingConfig using the standard helper function
	if thinkingConfig := llms.GetThinkingConfig(opts); thinkingConfig != nil && thinkingConfig.Mode != llms.ThinkingModeNone {
		config.ThinkingConfig = convertThinkingConfig(thinkingConfig)
	}

	return config
}

// convertHarmBlockThreshold converts our int32-based HarmBlockThreshold to the new SDK's string-based enum.
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
		// Default to HarmBlockOnlyHigh if unknown
		return genai.HarmBlockThresholdBlockOnlyHigh
	}
}

// convertThinkingConfig converts llms.ThinkingConfig to genai.ThinkingConfig.
func convertThinkingConfig(config *llms.ThinkingConfig) *genai.ThinkingConfig {
	if config == nil {
		return nil
	}

	genaiConfig := &genai.ThinkingConfig{}

	// Map ThinkingMode to ThinkingLevel
	// Note: Google Gemini API only supports LOW and HIGH thinking levels.
	// MEDIUM is not supported, so we map it to LOW as a middle ground.
	var thinkingLevel genai.ThinkingLevel
	switch config.Mode {
	case llms.ThinkingModeLow:
		thinkingLevel = genai.ThinkingLevelLow
	case llms.ThinkingModeMedium:
		// Medium is not supported by Gemini API, map to LOW as a moderate option
		// This differentiates it from HIGH and provides a middle ground
		thinkingLevel = genai.ThinkingLevelLow
	case llms.ThinkingModeHigh:
		thinkingLevel = genai.ThinkingLevelHigh
	case llms.ThinkingModeAuto:
		// Auto defaults to HIGH for maximum reasoning capability
		thinkingLevel = genai.ThinkingLevelHigh
	default:
		return nil // ThinkingModeNone or unknown
	}
	genaiConfig.ThinkingLevel = thinkingLevel

	// Set thinking budget if provided
	if config.BudgetTokens > 0 {
		budget := int32(config.BudgetTokens)
		genaiConfig.ThinkingBudget = &budget
	}

	// Set IncludeThoughts based on ReturnThinking
	genaiConfig.IncludeThoughts = config.ReturnThinking

	return genaiConfig
}

// convertToolConfig converts a ToolChoice to a genai.ToolConfig.
func convertToolConfig(config any) *genai.ToolConfig {
	if config == nil {
		return nil
	}

	toolConfig := &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{},
	}

	// Handle string-based tool choice (mode only)
	if c, ok := config.(string); ok {
		switch strings.ToLower(c) {
		case "any":
			toolConfig.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAny
		case "none":
			toolConfig.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeNone
		case "auto":
			toolConfig.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAuto
		default:
			// Unknown mode, default to AUTO
			toolConfig.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAuto
		}
	} else if c, ok := config.([]string); ok && len(c) > 0 {
		// Array of function names provided - use ANY mode with allowed function names
		toolConfig.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAny
		toolConfig.FunctionCallingConfig.AllowedFunctionNames = c
	} else {
		// Unknown type, default to AUTO mode
		toolConfig.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAuto
	}

	return toolConfig
}

// convertCandidatesFromResponse converts a GenerateContentResponse to llms.ContentResponse.
// This uses the SDK's Text() method to extract text content, which properly handles
// reasoning models and thought parts.
func convertCandidatesFromResponse(resp *genai.GenerateContentResponse) (*llms.ContentResponse, error) {
	return convertCandidates(resp.Candidates, resp.UsageMetadata, resp)
}

// convertCandidates converts a sequence of genai.Candidate to a response.
// If response is provided, its Text() method is used for more reliable text extraction.
func convertCandidates(candidates []*genai.Candidate, usage *genai.GenerateContentResponseUsageMetadata, response *genai.GenerateContentResponse) (*llms.ContentResponse, error) {
	var contentResponse llms.ContentResponse

	for i, candidate := range candidates {
		var textContent string
		var toolCalls []llms.ToolCall
		
		// Use the response's Text() method if available (more reliable, handles thoughts correctly)
		// For multi-candidate responses, we need to extract text per candidate
		if response != nil && i == 0 {
			// For the first candidate, we can use the response's Text() method
			// which handles all the edge cases properly
			textContent = response.Text()
		} else {
			// Fallback to manual extraction for additional candidates or when response is nil
			buf := strings.Builder{}
			if candidate.Content != nil && candidate.Content.Parts != nil {
				for _, part := range candidate.Content.Parts {
					if part == nil {
						continue
					}
					// Skip thought parts (reasoning models mark internal thinking as thoughts)
					// Only include actual text content, matching the SDK's Text() method behavior
					if part.Text != "" && !part.Thought {
						_, err := buf.WriteString(part.Text)
						if err != nil {
							return nil, err
						}
					}
				}
			}
			textContent = buf.String()
		}

		// Extract tool calls from parts (per candidate)
		if candidate.Content != nil && candidate.Content.Parts != nil {
			for _, part := range candidate.Content.Parts {
				if part == nil {
					continue
				}
				if part.FunctionCall != nil {
					b, err := json.Marshal(part.FunctionCall.Args)
					if err != nil {
						return nil, err
					}
					toolCall := llms.ToolCall{
						FunctionCall: &llms.FunctionCall{
							Name:      part.FunctionCall.Name,
							Arguments: string(b),
						},
					}
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}

		metadata := make(map[string]any)
		if candidate.CitationMetadata != nil {
			metadata[CITATIONS] = candidate.CitationMetadata
		}
		if candidate.SafetyRatings != nil {
			metadata[SAFETY] = candidate.SafetyRatings
		}

		if usage != nil {
			// Extract token counts - field names may differ in new SDK
			promptTokens := int(usage.PromptTokenCount)
			outputTokens := int(usage.CandidatesTokenCount)
			totalTokens := int(usage.TotalTokenCount)

			metadata["input_tokens"] = promptTokens
			metadata["output_tokens"] = outputTokens
			metadata["total_tokens"] = totalTokens
			// Standardized field names for cross-provider compatibility
			metadata["PromptTokens"] = promptTokens
			metadata["CompletionTokens"] = outputTokens
			metadata["TotalTokens"] = totalTokens

			// Extract thinking tokens from the new SDK
			// ThoughtsTokenCount is an int32 field in GenerateContentResponseUsageMetadata
			thinkingTokens := int(usage.ThoughtsTokenCount)
			metadata["ThinkingTokens"] = thinkingTokens

			// Cache-related token information (if available)
			if usage.CachedContentTokenCount > 0 {
				cachedCount := int(usage.CachedContentTokenCount)
				metadata["CachedTokens"] = cachedCount
				metadata["CacheReadInputTokens"] = cachedCount // Anthropic compatibility
				// Google AI includes cached tokens in the prompt count, calculate non-cached
				metadata["NonCachedInputTokens"] = promptTokens - cachedCount
			}
		}

		// Google AI doesn't separate thinking content like OpenAI o1, but we provide empty standardized fields
		metadata["ThinkingContent"] = "" // Google models don't separate thinking content

		// Note: Google AI's CachedContent requires pre-created cached content via API,
		// not inline cache control like Anthropic. Use Client.CreateCachedContent() for caching.

		finishReason := string(candidate.FinishReason)

		contentResponse.Choices = append(contentResponse.Choices,
			&llms.ContentChoice{
				Content:        textContent,
				StopReason:     finishReason,
				GenerationInfo: metadata,
				ToolCalls:      toolCalls,
			})
	}
	return &contentResponse, nil
}

// convertParts converts between a sequence of langchain parts and genai parts.
func convertParts(parts []llms.ContentPart) ([]*genai.Part, error) {
	convertedParts := make([]*genai.Part, 0, len(parts))
	for _, part := range parts {
		var out *genai.Part

		switch p := part.(type) {
		case llms.TextContent:
			out = &genai.Part{Text: p.Text}
		case llms.BinaryContent:
			inlineData := &genai.Blob{
				MIMEType: p.MIMEType,
				Data:     p.Data,
			}
			out = &genai.Part{InlineData: inlineData}
		case llms.ImageURLContent:
			typ, data, err := imageutil.DownloadImageData(p.URL)
			if err != nil {
				return nil, err
			}
			inlineData := &genai.Blob{
				MIMEType: typ,
				Data:     data,
			}
			out = &genai.Part{InlineData: inlineData}
		case llms.ToolCall:
			fc := p.FunctionCall
			var argsMap map[string]any
			if err := json.Unmarshal([]byte(fc.Arguments), &argsMap); err != nil {
				return convertedParts, err
			}
			functionCall := &genai.FunctionCall{
				Name: fc.Name,
				Args: argsMap,
			}
			out = &genai.Part{FunctionCall: functionCall}
		case llms.ToolCallResponse:
			functionResponse := &genai.FunctionResponse{
				Name: p.Name,
				Response: map[string]any{
					"response": p.Content,
				},
			}
			out = &genai.Part{FunctionResponse: functionResponse}
		}

		if out != nil {
			convertedParts = append(convertedParts, out)
		}
	}
	return convertedParts, nil
}

// convertContent converts between a langchain MessageContent and genai content.
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
		role = RoleUser
	default:
		return nil, fmt.Errorf("role %v not supported", content.Role)
	}

	c := &genai.Content{
		Parts: parts,
		Role:  role,
	}

	return c, nil
}

// generateFromSingleMessage generates content from the parts of a single
// message.
func generateFromSingleMessage(
	ctx context.Context,
	client *genai.Client,
	model string,
	parts []llms.ContentPart,
	opts *llms.CallOptions,
	clientOpts Options,
) (*llms.ContentResponse, error) {
	convertedParts, err := convertParts(parts)
	if err != nil {
		return nil, err
	}

	contents := []*genai.Content{
		{Parts: convertedParts},
	}

	// Build GenerateContentConfig
	config := buildGenerateContentConfig(opts, clientOpts)

	if opts.StreamingFunc == nil {
		// When no streaming is requested, just call GenerateContent and return
		// the complete response with a list of candidates.
		resp, err := client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			return nil, googleaierrors.MapError(err)
		}

		if len(resp.Candidates) == 0 {
			return nil, ErrNoContentInResponse
		}
		response, err := convertCandidatesFromResponse(resp)
		if err != nil {
			return nil, googleaierrors.MapError(err)
		}
		return response, nil
	}

	// Streaming is requested - use GenerateContentStream
	return convertAndStreamFromIterator(ctx, client, model, contents, config, opts)
}

func generateFromMessages(
	ctx context.Context,
	client *genai.Client,
	model string,
	messages []llms.MessageContent,
	opts *llms.CallOptions,
	clientOpts Options,
) (*llms.ContentResponse, error) {
	contents := make([]*genai.Content, 0, len(messages))
	var systemInstruction *genai.Content

	for _, mc := range messages {
		content, err := convertContent(mc)
		if err != nil {
			return nil, err
		}
		if mc.Role == llms.ChatMessageTypeSystem {
			systemInstruction = content
			continue
		}
		contents = append(contents, content)
	}

	// Build GenerateContentConfig
	config := buildGenerateContentConfig(opts, clientOpts)
	if systemInstruction != nil {
		config.SystemInstruction = systemInstruction
	}

	if opts.StreamingFunc == nil {
		resp, err := client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			return nil, googleaierrors.MapError(err)
		}

		if len(resp.Candidates) == 0 {
			return nil, ErrNoContentInResponse
		}
		return convertCandidatesFromResponse(resp)
	}

	// Streaming is requested - use GenerateContentStream
	return convertAndStreamFromIterator(ctx, client, model, contents, config, opts)
}

// convertAndStreamFromIterator handles streaming responses from the new SDK.
// It iterates over the stream, calls the StreamingFunc for each chunk, and
// accumulates the final response.
func convertAndStreamFromIterator(
	ctx context.Context,
	client *genai.Client,
	model string,
	contents []*genai.Content,
	config *genai.GenerateContentConfig,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	// Get the streaming iterator
	stream := client.Models.GenerateContentStream(ctx, model, contents, config)

	// Track accumulated content and final response
	var finalResponse *genai.GenerateContentResponse
	var accumulatedTextLen int

	// Iterate over the stream
	for response, err := range stream {
		if err != nil {
			return nil, googleaierrors.MapError(err)
		}
		if response == nil {
			continue
		}

		// Store the final response (last one contains usage metadata)
		finalResponse = response

		// Extract text from this chunk
		// Note: response.Text() returns the full accumulated text from all parts
		// We need to extract only the new incremental text
		currentFullText := response.Text()
		currentLen := len(currentFullText)

		// If this response has more text than we've seen, extract the delta
		if currentLen > accumulatedTextLen {
			newText := currentFullText[accumulatedTextLen:]
			if len(newText) > 0 {
				// Call the streaming function with the new chunk
				if err := opts.StreamingFunc(ctx, []byte(newText)); err != nil {
					return nil, fmt.Errorf("streaming function error: %w", err)
				}
				accumulatedTextLen = currentLen
			}
		}
	}

	// Convert the final response to llms.ContentResponse
	if finalResponse == nil || len(finalResponse.Candidates) == 0 {
		return nil, ErrNoContentInResponse
	}

	return convertCandidatesFromResponse(finalResponse)
}

// convertSchemaRecursive recursively converts a schema map to a genai.Schema
func convertSchemaRecursive(schemaMap map[string]any, toolIndex int, propertyPath string) (*genai.Schema, error) {
	schema := &genai.Schema{}

	if ty, ok := schemaMap["type"]; ok {
		tyString, ok := ty.(string)
		if !ok {
			return nil, fmt.Errorf("tool [%d], property [%s]: expected string for type", toolIndex, propertyPath)
		}
		schema.Type = convertToolSchemaType(tyString)
	}

	if desc, ok := schemaMap["description"]; ok {
		descString, ok := desc.(string)
		if !ok {
			return nil, fmt.Errorf("tool [%d], property [%s]: expected string for description", toolIndex, propertyPath)
		}
		schema.Description = descString
	}

	// Handle object properties recursively
	if properties, ok := schemaMap["properties"]; ok {
		propMap, ok := properties.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool [%d], property [%s]: expected map for properties", toolIndex, propertyPath)
		}

		schema.Properties = make(map[string]*genai.Schema)
		for propName, propValue := range propMap {
			valueMap, ok := propValue.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("tool [%d], property [%s.%s]: expect to find a value map", toolIndex, propertyPath, propName)
			}

			nestedPath := propName
			if propertyPath != "" {
				nestedPath = propertyPath + "." + propName
			}

			nestedSchema, err := convertSchemaRecursive(valueMap, toolIndex, nestedPath)
			if err != nil {
				return nil, err
			}
			schema.Properties[propName] = nestedSchema
		}
	} else if schema.Type == genai.TypeObject && propertyPath == "" {
		// For top-level object schemas without properties, this is an error
		return nil, fmt.Errorf("tool [%d]: expected to find a map of properties", toolIndex)
	}

	// Handle array items recursively
	if items, ok := schemaMap["items"]; ok && schema.Type == genai.TypeArray {
		itemMap, ok := items.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool [%d], property [%s]: expect to find a map for array items", toolIndex, propertyPath)
		}

		itemsPath := propertyPath + "[]"
		itemsSchema, err := convertSchemaRecursive(itemMap, toolIndex, itemsPath)
		if err != nil {
			return nil, err
		}
		schema.Items = itemsSchema
	}

	// Handle required fields
	if required, ok := schemaMap["required"]; ok {
		if rs, ok := required.([]string); ok {
			schema.Required = rs
		} else if ri, ok := required.([]interface{}); ok {
			rs := make([]string, 0, len(ri))
			for _, r := range ri {
				rString, ok := r.(string)
				if !ok {
					return nil, fmt.Errorf("tool [%d], property [%s]: expected string for required", toolIndex, propertyPath)
				}
				rs = append(rs, rString)
			}
			schema.Required = rs
		} else {
			return nil, fmt.Errorf("tool [%d], property [%s]: expected array for required", toolIndex, propertyPath)
		}
	}

	return schema, nil
}

// convertTools converts from a list of langchaingo tools to a list of genai
// tools.
func convertTools(tools []llms.Tool) ([]*genai.Tool, error) {
	genaiFuncDecls := make([]*genai.FunctionDeclaration, 0, len(tools))
	for i, tool := range tools {
		if tool.Type != "function" {
			return nil, fmt.Errorf("tool [%d]: unsupported type %q, want 'function'", i, tool.Type)
		}

		// We have a llms.FunctionDefinition in tool.Function, and we have to
		// convert it to genai.FunctionDeclaration
		genaiFuncDecl := &genai.FunctionDeclaration{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
		}

		// Expect the Parameters field to be a map[string]any, from which we will
		// extract properties to populate the schema.
		params, ok := tool.Function.Parameters.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool [%d]: unsupported type %T of Parameters", i, tool.Function.Parameters)
		}

		schema, err := convertSchemaRecursive(params, i, "")
		if err != nil {
			return nil, err
		}
		genaiFuncDecl.Parameters = schema

		// google genai only support one tool, multiple tools must be embedded into function declarations:
		// https://github.com/GoogleCloudPlatform/generative-ai/issues/636
		// https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling#chat-samples
		genaiFuncDecls = append(genaiFuncDecls, genaiFuncDecl)
	}

	// Return nil if no tools are provided
	if len(genaiFuncDecls) == 0 {
		return nil, nil
	}

	genaiTools := []*genai.Tool{{FunctionDeclarations: genaiFuncDecls}}

	return genaiTools, nil
}

// convertToolSchemaType converts a tool's schema type from its langchaingo
// representation (string) to a genai enum.
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

// showContent is a debugging helper for genai.Content.
func showContent(w io.Writer, cs []*genai.Content) {
	fmt.Fprintf(w, "Content (len=%v)\n", len(cs))
	for i, c := range cs {
		role := c.Role
		fmt.Fprintf(w, "[%d]: Role=%s\n", i, role)
		for j, p := range c.Parts {
			fmt.Fprintf(w, "  Parts[%v]: ", j)
			if p.Text != "" {
				fmt.Fprintf(w, "Text %q\n", p.Text)
			} else if p.InlineData != nil {
				fmt.Fprintf(w, "Blob MIME=%q, size=%d\n", p.InlineData.MIMEType, len(p.InlineData.Data))
			} else if p.FunctionCall != nil {
				fmt.Fprintf(w, "FunctionCall Name=%v, Args=%v\n", p.FunctionCall.Name, p.FunctionCall.Args)
			} else if p.FunctionResponse != nil {
				fmt.Fprintf(w, "FunctionResponse Name=%v Response=%v\n", p.FunctionResponse.Name, p.FunctionResponse.Response)
			} else {
				fmt.Fprintf(w, "unknown type\n")
			}
		}
	}
}
