// This file contains the Vertex AI implementation for langchaingo.
// It uses google.golang.org/genai for API interactions.

//nolint:all
package vertex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/internal/imageutil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
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

// Ptr returns a pointer to t
func Ptr[T any](t T) *T {
	return &t
}

// convertHarmBlockThreshold converts googleai.HarmBlockThreshold (int32) to genai.HarmBlockThreshold (string)
func convertHarmBlockThreshold(threshold googleai.HarmBlockThreshold) genai.HarmBlockThreshold {
	switch threshold {
	case googleai.HarmBlockUnspecified:
		return genai.HarmBlockThreshold("HARM_BLOCK_THRESHOLD_UNSPECIFIED")
	case googleai.HarmBlockLowAndAbove:
		return genai.HarmBlockThreshold("BLOCK_LOW_AND_ABOVE")
	case googleai.HarmBlockMediumAndAbove:
		return genai.HarmBlockThreshold("BLOCK_MEDIUM_AND_ABOVE")
	case googleai.HarmBlockOnlyHigh:
		return genai.HarmBlockThreshold("BLOCK_ONLY_HIGH")
	case googleai.HarmBlockNone:
		return genai.HarmBlockThreshold("BLOCK_NONE")
	default:
		return genai.HarmBlockThreshold("BLOCK_MEDIUM_AND_ABOVE") // Safe default
	}
}

// Call implements the [llms.Model] interface.
func (g *Vertex) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, g, prompt, options...)
}

// GenerateContent implements the [llms.Model] interface.
func (g *Vertex) GenerateContent(
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

	// Build config for the new API
	config := &genai.GenerateContentConfig{
		CandidateCount:  int32(opts.CandidateCount),
		MaxOutputTokens: int32(opts.MaxTokens),
		Temperature:     Ptr(float32(opts.Temperature)),
		TopP:            Ptr(float32(opts.TopP)),
		TopK:            Ptr(float32(opts.TopK)),
		StopSequences:   opts.StopWords,
		SafetySettings: []*genai.SafetySetting{
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
		},
	}

	var err error
	if config.Tools, err = convertTools(opts.Tools); err != nil {
		return nil, err
	}

	// Handle ResponseMIMEType/JSONMode
	switch {
	case opts.ResponseMIMEType != "" && opts.JSONMode:
		return nil, fmt.Errorf("conflicting options, can't use JSONMode and ResponseMIMEType together")
	case opts.ResponseMIMEType != "" && !opts.JSONMode:
		config.ResponseMIMEType = opts.ResponseMIMEType
	case opts.ResponseMIMEType == "" && opts.JSONMode:
		config.ResponseMIMEType = ResponseMIMETypeJson
	}

	var response *llms.ContentResponse

	if len(messages) == 1 {
		theMessage := messages[0]
		if theMessage.Role != llms.ChatMessageTypeHuman {
			return nil, fmt.Errorf("got %v message role, want human", theMessage.Role)
		}
		response, err = generateFromSingleMessage(ctx, g, opts.Model, theMessage.Parts, config, &opts)
	} else {
		response, err = generateFromMessages(ctx, g, opts.Model, messages, config, &opts)
	}
	if err != nil {
		return nil, err
	}

	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

// convertCandidates converts a sequence of genai.Candidate to a response.
func convertCandidates(candidates []*genai.Candidate, usage *genai.GenerateContentResponseUsageMetadata) (*llms.ContentResponse, error) {
	var contentResponse llms.ContentResponse
	var toolCalls []llms.ToolCall

	for _, candidate := range candidates {
		buf := strings.Builder{}

		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				// New API uses Part struct with fields instead of interface types
				if part.Text != "" {
					_, err := buf.WriteString(part.Text)
					if err != nil {
						return nil, err
					}
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
		metadata[CITATIONS] = candidate.CitationMetadata
		metadata[SAFETY] = candidate.SafetyRatings

		if usage != nil {
			metadata["input_tokens"] = usage.PromptTokenCount
			metadata["output_tokens"] = usage.CandidatesTokenCount
			metadata["total_tokens"] = usage.TotalTokenCount
		}

		contentResponse.Choices = append(contentResponse.Choices,
			&llms.ContentChoice{
				Content:        buf.String(),
				StopReason:     string(candidate.FinishReason),
				GenerationInfo: metadata,
				ToolCalls:      toolCalls,
			})
	}
	return &contentResponse, nil
}

// convertParts converts between a sequence of langchain parts and genai parts.
// Returns []*genai.Part since the new API uses pointers.
func convertParts(parts []llms.ContentPart) ([]*genai.Part, error) {
	convertedParts := make([]*genai.Part, 0, len(parts))
	for _, part := range parts {
		var out *genai.Part

		switch p := part.(type) {
		case llms.TextContent:
			out = &genai.Part{Text: p.Text}
		case llms.BinaryContent:
			out = &genai.Part{InlineData: &genai.Blob{MIMEType: p.MIMEType, Data: p.Data}}
		case llms.ImageURLContent:
			typ, data, err := imageutil.DownloadImageData(p.URL)
			if err != nil {
				return nil, err
			}
			out = &genai.Part{InlineData: &genai.Blob{MIMEType: typ, Data: data}}
		case llms.ToolCall:
			fc := p.FunctionCall
			var argsMap map[string]any
			if err := json.Unmarshal([]byte(fc.Arguments), &argsMap); err != nil {
				return convertedParts, err
			}
			out = &genai.Part{FunctionCall: &genai.FunctionCall{
				Name: fc.Name,
				Args: argsMap,
			}}
		case llms.ToolCallResponse:
			out = &genai.Part{FunctionResponse: &genai.FunctionResponse{
				Name: p.Name,
				Response: map[string]any{
					"response": p.Content,
				},
			}}
		}

		convertedParts = append(convertedParts, out)
	}
	return convertedParts, nil
}

// convertContent converts between a langchain MessageContent and genai content.
func convertContent(content llms.MessageContent) (*genai.Content, error) {
	parts, err := convertParts(content.Parts)
	if err != nil {
		return nil, err
	}

	c := &genai.Content{
		Parts: parts,
	}

	switch content.Role {
	case llms.ChatMessageTypeSystem:
		c.Role = RoleSystem
	case llms.ChatMessageTypeAI:
		c.Role = RoleModel
	case llms.ChatMessageTypeHuman:
		c.Role = RoleUser
	case llms.ChatMessageTypeGeneric:
		c.Role = RoleUser
	case llms.ChatMessageTypeTool:
		c.Role = RoleUser
	case llms.ChatMessageTypeFunction:
		fallthrough
	default:
		return nil, fmt.Errorf("role %v not supported", content.Role)
	}

	return c, nil
}

// generateFromSingleMessage generates content from the parts of a single
// message using the new google.golang.org/genai API.
func generateFromSingleMessage(
	ctx context.Context,
	client *Vertex,
	modelName string,
	parts []llms.ContentPart,
	config *genai.GenerateContentConfig,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	convertedParts, err := convertParts(parts)
	if err != nil {
		return nil, err
	}

	contents := []*genai.Content{
		{
			Parts: convertedParts,
			Role:  genai.RoleUser,
		},
	}

	if opts.StreamingFunc == nil {
		// When no streaming is requested, just call GenerateContent and return
		// the complete response with a list of candidates.
		resp, err := client.client.Models.GenerateContent(ctx, modelName, contents, config)
		if err != nil {
			return nil, err
		}

		if len(resp.Candidates) == 0 {
			return nil, ErrNoContentInResponse
		}
		return convertCandidates(resp.Candidates, resp.UsageMetadata)
	}
	iter := client.client.Models.GenerateContentStream(ctx, modelName, contents, config)
	return convertAndStreamFromIterator(ctx, iter, opts)
}

func generateFromMessages(
	ctx context.Context,
	client *Vertex,
	modelName string,
	messages []llms.MessageContent,
	config *genai.GenerateContentConfig,
	opts *llms.CallOptions,
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
			config.SystemInstruction = systemInstruction
			continue
		}
		contents = append(contents, content)
	}

	if opts.StreamingFunc == nil {
		resp, err := client.client.Models.GenerateContent(ctx, modelName, contents, config)
		if err != nil {
			return nil, err
		}

		if len(resp.Candidates) == 0 {
			return nil, ErrNoContentInResponse
		}
		return convertCandidates(resp.Candidates, resp.UsageMetadata)
	}
	iter := client.client.Models.GenerateContentStream(ctx, modelName, contents, config)
	return convertAndStreamFromIterator(ctx, iter, opts)
}

// convertAndStreamFromIterator takes an iterator of GenerateContentResponse
// and produces a llms.ContentResponse reply from it, while streaming the
// resulting text into the opts-provided streaming function.
// Note that this is tricky in the face of multiple
// candidates, so this code assumes only a single candidate for now.
// TODO: Rewrite for new iter.Seq2 API
func convertAndStreamFromIterator(
	ctx context.Context,
	iter interface{}, // iter.Seq2[*GenerateContentResponse, error]
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	// TODO: Implement streaming for new iter.Seq2 API
	return nil, fmt.Errorf("streaming not yet implemented for new API")
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

		schema := &genai.Schema{}
		if ty, ok := params["type"]; ok {
			tyString, ok := ty.(string)
			if !ok {
				return nil, fmt.Errorf("tool [%d]: expected string for type", i)
			}
			schema.Type = convertToolSchemaType(tyString)
		}

		paramProperties, ok := params["properties"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool [%d]: expected to find a map of properties", i)
		}

		schema.Properties = make(map[string]*genai.Schema)
		for propName, propValue := range paramProperties {
			valueMap, ok := propValue.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("tool [%d], property [%v]: expect to find a value map", i, propName)
			}
			schema.Properties[propName] = &genai.Schema{}

			if ty, ok := valueMap["type"]; ok {
				tyString, ok := ty.(string)
				if !ok {
					return nil, fmt.Errorf("tool [%d]: expected string for type", i)
				}
				schema.Properties[propName].Type = convertToolSchemaType(tyString)
			}
			if desc, ok := valueMap["description"]; ok {
				descString, ok := desc.(string)
				if !ok {
					return nil, fmt.Errorf("tool [%d]: expected string for description", i)
				}
				schema.Properties[propName].Description = descString
			}
		}

		if required, ok := params["required"]; ok {
			if rs, ok := required.([]string); ok {
				schema.Required = rs
			} else if ri, ok := required.([]interface{}); ok {
				rs := make([]string, 0, len(ri))
				for _, r := range ri {
					rString, ok := r.(string)
					if !ok {
						return nil, fmt.Errorf("tool [%d]: expected string for required", i)
					}
					rs = append(rs, rString)
				}
				schema.Required = rs
			} else {
				return nil, fmt.Errorf("tool [%d]: expected string for required", i)
			}
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
		fmt.Fprintf(w, "[%d]: Role=%s\n", i, c.Role)
		for j, p := range c.Parts {
			fmt.Fprintf(w, "  Parts[%v]: ", j)
			// New API uses struct fields
			if p.Text != "" {
				fmt.Fprintf(w, "Text %q\n", p.Text)
			}
			if p.InlineData != nil {
				fmt.Fprintf(w, "Blob MIME=%q, size=%d\n", p.InlineData.MIMEType, len(p.InlineData.Data))
			}
			if p.FunctionCall != nil {
				fmt.Fprintf(w, "FunctionCall Name=%v, Args=%v\n", p.FunctionCall.Name, p.FunctionCall.Args)
			}
			if p.FunctionResponse != nil {
				fmt.Fprintf(w, "FunctionResponse Name=%v Response=%v\n", p.FunctionResponse.Name, p.FunctionResponse.Response)
			}
		}
	}
}
