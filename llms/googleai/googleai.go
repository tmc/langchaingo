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

	// Build generation config
	temperature := float32(opts.Temperature)
	topP := float32(opts.TopP)
	topK := float32(opts.TopK)

	config := &genai.GenerateContentConfig{
		CandidateCount:  int32(opts.CandidateCount),
		MaxOutputTokens: int32(opts.MaxTokens),
		Temperature:     &temperature,
		TopP:            &topP,
		TopK:            &topK,
		StopSequences:   opts.StopWords,
	}

	// Handle response MIME type and JSON mode
	switch {
	case opts.ResponseMIMEType != "" && opts.JSONMode:
		return nil, fmt.Errorf("conflicting options, can't use JSONMode and ResponseMIMEType together")
	case opts.ResponseMIMEType != "" && !opts.JSONMode:
		config.ResponseMIMEType = opts.ResponseMIMEType
	case opts.ResponseMIMEType == "" && opts.JSONMode:
		config.ResponseMIMEType = ResponseMIMETypeJson
	}

	// Handle thinking configuration for reasoning models
	if opts.Reasoning != nil && opts.Reasoning.IsEnabled() {
		thinkingBudget := int32(opts.Reasoning.GetTokens(opts.MaxTokens))
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
		response, err = g.generateFromSingleMessage(ctx, opts.Model, theMessage.Parts, config, &opts)
	} else {
		response, err = g.generateFromMessages(ctx, opts.Model, messages, config, &opts)
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
		if len(chunk.Candidates) == 0 {
			continue
		}

		candidate := chunk.Candidates[0]
		if candidate.Content == nil {
			continue
		}

		for _, part := range candidate.Content.Parts {
			if len(part.Text) > 0 {
				if part.Thought {
					accumulatedReasoningContent.WriteString(part.Text)
					chunk := streaming.Chunk{
						Type:    streaming.ChunkTypeReasoning,
						Content: part.Text,
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
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content:          accumulatedContent.String(),
			ReasoningContent: accumulatedReasoningContent.String(),
			ToolCalls:        accumulatedToolCalls,
		}},
	}, nil
}

func convertResponse(resp *genai.GenerateContentResponse) (*llms.ContentResponse, error) {
	if len(resp.Candidates) == 0 {
		return nil, ErrNoContentInResponse
	}

	var choices []*llms.ContentChoice

	for _, candidate := range resp.Candidates {
		var buf strings.Builder
		var toolCalls []llms.ToolCall
		var thinkingContent string

		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if len(part.Text) > 0 {
					// Check if this is thinking content
					if part.Thought {
						thinkingContent += part.Text
					} else {
						buf.WriteString(part.Text)
					}
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
					toolCalls = append(toolCalls, toolCall)
				}
			}
		}

		metadata := make(map[string]any)
		metadata[CITATIONS] = candidate.CitationMetadata
		metadata[SAFETY] = candidate.SafetyRatings

		if resp.UsageMetadata != nil {
			metadata["input_tokens"] = resp.UsageMetadata.PromptTokenCount
			metadata["output_tokens"] = resp.UsageMetadata.CandidatesTokenCount
			metadata["total_tokens"] = resp.UsageMetadata.TotalTokenCount
		}

		choices = append(choices, &llms.ContentChoice{
			Content:          buf.String(),
			ReasoningContent: thinkingContent,
			StopReason:       string(candidate.FinishReason),
			GenerationInfo:   metadata,
			ToolCalls:        toolCalls,
		})
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func convertParts(parts []llms.ContentPart) ([]*genai.Part, error) {
	convertedParts := make([]*genai.Part, 0, len(parts))

	for _, part := range parts {
		var genaiPart *genai.Part

		switch p := part.(type) {
		case llms.TextContent:
			genaiPart = &genai.Part{Text: p.Text}

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

		schema, err := convertToSchema(tool.Function.Parameters, true)
		if err != nil {
			return nil, fmt.Errorf("tool [%d]: %w", i, err)
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

func convertToSchema(e any, topLevel bool) (*genai.Schema, error) {
	e, err := convertMaps(e)
	if err != nil {
		return nil, err
	}
	schema := &genai.Schema{}

	eMap, ok := e.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("tool: unsupported type %T of Parameters", e)
	}

	if ty, ok := eMap["type"]; ok {
		tyString, ok := ty.(string)
		if !ok {
			return nil, fmt.Errorf("tool: expected string for type")
		}
		schema.Type = convertToolSchemaType(tyString)

		if topLevel && schema.Type != genai.TypeObject {
			return nil, fmt.Errorf("tool: top-level schema must be an object")
		}
	}

	if properties, ok := eMap["properties"]; ok {
		paramProperties, ok := properties.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool: expected map[string]any for properties")
		}
		schema.Properties = make(map[string]*genai.Schema)
		for propName, propValue := range paramProperties {
			recSchema, err := convertToSchema(propValue, false)
			if err != nil {
				return nil, fmt.Errorf("tool, property [%v]: %w", propName, err)
			}
			schema.Properties[propName] = recSchema
		}
	} else if schema.Type == genai.TypeObject {
		return nil, fmt.Errorf("tool: object schema must have properties")
	}

	if items, ok := eMap["items"]; ok {
		itemsSchema, err := convertToSchema(items, false)
		if err != nil {
			return nil, fmt.Errorf("tool: %w", err)
		}
		schema.Items = itemsSchema
	} else if schema.Type == genai.TypeArray {
		return nil, fmt.Errorf("tool: array schema must have items")
	}

	if description, ok := eMap["description"]; ok {
		descString, ok := description.(string)
		if !ok {
			return nil, fmt.Errorf("tool: expected string for description")
		}
		schema.Description = descString
	}

	if nullable, ok := eMap["nullable"]; ok {
		nullableBool, ok := nullable.(bool)
		if !ok {
			return nil, fmt.Errorf("tool: expected bool for nullable")
		}
		schema.Nullable = &nullableBool
	}

	if enum, ok := eMap["enum"]; ok {
		enumSlice, err := convertToSliceOfStrings(enum)
		if err != nil {
			return nil, fmt.Errorf("tool: %w", err)
		}
		schema.Enum = enumSlice
	}

	if required, ok := eMap["required"]; ok {
		requiredSlice, err := convertToSliceOfStrings(required)
		if err != nil {
			return nil, fmt.Errorf("tool: %w", err)
		}
		schema.Required = requiredSlice
	}

	return schema, nil
}

func convertToSliceOfStrings(e any) ([]string, error) {
	if rs, ok := e.([]string); ok {
		return rs, nil
	}

	ri, ok := e.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tool: expected []interface{} for required")
	}
	rs := make([]string, 0, len(ri))
	for _, r := range ri {
		rString, ok := r.(string)
		if !ok {
			return nil, fmt.Errorf("tool: expected string for required")
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
