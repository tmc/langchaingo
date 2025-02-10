//nolint:all
package googlegenai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/tmc/langchaingo/internal/imageutil"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/iterator"
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

	topK := float64(opts.TopK)
	candidateCount := int64(opts.CandidateCount)
	maxtokens := int64(opts.MaxTokens)
	seed := int64(opts.Seed)
	config := &genai.GenerateContentConfig{
		SystemInstruction: nil,
		Temperature:       &opts.Temperature,
		TopP:              nil,
		TopK:              &topK,
		CandidateCount:    &candidateCount,
		MaxOutputTokens:   &maxtokens,
		StopSequences:     opts.StopWords,
		ResponseLogprobs:  false,
		Logprobs:          nil,
		PresencePenalty:   nil,
		FrequencyPenalty:  nil,
		Seed:              &seed,
		ResponseMIMEType:  "",
		RoutingConfig:     nil,
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThreshold(g.opts.HarmThreshold),
			},
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThreshold(g.opts.HarmThreshold),
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThreshold(g.opts.HarmThreshold),
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThreshold(g.opts.HarmThreshold),
			},
		},
		Tools:              nil,
		ToolConfig:         nil,
		CachedContent:      "",
		ResponseModalities: nil,
		MediaResolution:    "",
		SpeechConfig:       nil,
		AudioTimestamp:     false,
		ThinkingConfig:     nil,
	}

	var content []*genai.Content
	for _, message := range messages {
		c, err := convertContent(message)
		if err != nil {
			return nil, err
		}
		content = append(content, c)
	}
	var err error
	var response *llms.ContentResponse

	if len(messages) == 1 {
		theMessage := messages[0]
		if theMessage.Role != llms.ChatMessageTypeHuman {
			return nil, fmt.Errorf("got %v message role, want human", theMessage.Role)
		}
		response, err = generateFromSingleMessage(ctx, g.client.Models, theMessage.Parts, opts, config)
	} else {
		response, err = generateFromMessages(ctx, g.client.Models, messages, &opts, config)
	}
	if err != nil {
		return nil, err
	}

	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

func (g *GoogleAI) GenerateImage(
	ctx context.Context,
	message llms.TextContent,
	options ...llms.CallOption,
) ([]*llms.BinaryContent, error) {
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

	var numberOfImages int64 = 1
	numberOfImagesI, ok := opts.Metadata["number_of_images"]
	if ok {
		numberOfImages = numberOfImagesI.(int64)
	}

	var seed int64 = int64(opts.Seed)
	config := &genai.GenerateImagesConfig{
		NumberOfImages: &numberOfImages,
		Seed:           &seed,
		OutputMIMEType: opts.ResponseMIMEType,
	}

	var response []*llms.BinaryContent
	resp, err := g.client.Models.GenerateImages(ctx, opts.Model, message.Text, config)
	if err != nil {
		return nil, err
	}
	if len(resp.GeneratedImages) == 0 {
		return nil, ErrNoContentInResponse
	}

	for _, image := range resp.GeneratedImages {
		if image.Image == nil {
			continue
		}

		response = append(response, &llms.BinaryContent{
			Data:     image.Image.ImageBytes,
			MIMEType: image.Image.MIMEType,
		})
	}
	return response, nil
}

// convertCandidates converts a sequence of genai.Candidate to a response.
func convertCandidates(candidates []*genai.Candidate) (*llms.ContentResponse, error) {
	var contentResponse llms.ContentResponse
	var toolCalls []llms.ToolCall

	for _, candidate := range candidates {
		buf := strings.Builder{}

		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				switch {
				case part.Text != "":
					_, err := buf.WriteString(part.Text)
					if err != nil {
						return nil, err
					}
				case part.FunctionCall != nil:
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
				default:
					return nil, ErrUnknownPartInResponse
				}
			}
		}

		metadata := make(map[string]any)
		metadata[CITATIONS] = candidate.CitationMetadata
		metadata[SAFETY] = candidate.SafetyRatings

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
func convertParts(parts []llms.ContentPart) ([]*genai.Part, error) {
	convertedParts := make([]*genai.Part, 0, len(parts))
	for _, part := range parts {
		var out *genai.Part

		switch p := part.(type) {
		case llms.TextContent:
			out = genai.NewPartFromText(p.Text)
		case llms.BinaryContent:
			out = genai.NewPartFromBytes(p.Data, p.MIMEType)
		case llms.ImageURLContent:
			typ, data, err := imageutil.DownloadImageData(p.URL)
			if err != nil {
				return nil, err
			}
			out = genai.NewPartFromBytes(data, typ)
		case llms.ToolCall:
			fc := p.FunctionCall
			var argsMap map[string]any
			if err := json.Unmarshal([]byte(fc.Arguments), &argsMap); err != nil {
				return convertedParts, err
			}
			out = genai.NewPartFromFunctionCall(fc.Name, argsMap)
		case llms.ToolCallResponse:
			out = genai.NewPartFromFunctionResponse(p.Name, map[string]any{
				"response": p.Content,
			})
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
// message.
func generateFromSingleMessage(
	ctx context.Context,
	model *genai.Models,
	parts []llms.ContentPart,
	llmOpts llms.CallOptions,
	opts *genai.GenerateContentConfig,
) (*llms.ContentResponse, error) {
	convertedParts, err := convertParts(parts)
	if err != nil {
		return nil, err
	}

	contents := []*genai.Content{&genai.Content{
		Parts: convertedParts,
	}}
	if llmOpts.StreamingFunc == nil {
		// When no streaming is requested, just call GenerateContent and return
		// the complete response with a list of candidates.

		resp, err := model.GenerateContent(ctx, llmOpts.Model, contents, opts)
		if err != nil {
			return nil, err
		}

		if len(resp.Candidates) == 0 {
			return nil, ErrNoContentInResponse
		}
		return convertCandidates(resp.Candidates)
	}

	return convertAndStreamFromIterator(ctx, model.GenerateContentStream(ctx, llmOpts.Model, contents, opts), &llmOpts)
}

func generateFromMessages(
	ctx context.Context,
	model *genai.Models,
	messages []llms.MessageContent,
	opts *llms.CallOptions,
	genaiOpts *genai.GenerateContentConfig,
) (*llms.ContentResponse, error) {
	history := make([]*genai.Content, 0, len(messages))
	for _, mc := range messages {
		content, err := convertContent(mc)
		if err != nil {
			return nil, err
		}
		if mc.Role == RoleSystem {
			genaiOpts.SystemInstruction = content
			continue
		}
		history = append(history, content)
	}

	// Given N total messages, genai's chat expects the first N-1 messages as
	// history and the last message as the actual request.
	var contents []*genai.Content
	for _, msg := range messages {
		conv, err := convertContent(msg)
		if err != nil {
			return nil, err
		}
		contents = append(contents, conv)
	}

	return convertAndStreamFromIterator(ctx, model.GenerateContentStream(ctx, opts.Model, contents, genaiOpts), opts)
}

// convertAndStreamFromIterator takes an iterator of GenerateContentResponse
// and produces a llms.ContentResponse reply from it, while streaming the
// resulting text into the opts-provided streaming function.
// Note that this is tricky in the face of multiple
// candidates, so this code assumes only a single candidate for now.

func convertAndStreamFromIterator(
	ctx context.Context,
	iter iter.Seq2[*genai.GenerateContentResponse, error],
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	candidate := &genai.Candidate{
		Content: &genai.Content{},
	}
DoStream:
	for resp, err := range iter {
		if errors.Is(err, iterator.Done) {
			break DoStream
		}
		if err != nil {
			return nil, fmt.Errorf("error in stream mode: %w", err)
		}

		if len(resp.Candidates) != 1 {
			return nil, fmt.Errorf("expect single candidate in stream mode; got %v", len(resp.Candidates))
		}
		respCandidate := resp.Candidates[0]

		if respCandidate.Content == nil {
			break DoStream
		}
		candidate.Content.Parts = append(candidate.Content.Parts, respCandidate.Content.Parts...)
		candidate.Content.Role = respCandidate.Content.Role
		candidate.FinishReason = respCandidate.FinishReason
		candidate.SafetyRatings = respCandidate.SafetyRatings
		candidate.CitationMetadata = respCandidate.CitationMetadata
		candidate.TokenCount = respCandidate.TokenCount

		for _, part := range respCandidate.Content.Parts {
			if ok := part.Text != ""; ok {
				if opts.StreamingFunc(ctx, []byte(part.Text)) != nil {
					break DoStream
				}
			}
		}
	}

	return convertCandidates([]*genai.Candidate{candidate})
}

// convertTools converts from a list of langchaingo tools to a list of genai
// tools.
func convertTools(tools []llms.Tool) ([]*genai.Tool, error) {
	genaiTools := make([]*genai.Tool, 0, len(tools))
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

		genaiTools = append(genaiTools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{genaiFuncDecl},
		})
	}

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
			switch {
			case p.Text != "":
				fmt.Fprintf(w, "Text %q\n", p.Text)
			case p.InlineData != nil:
				fmt.Fprintf(w, "Blob MIME=%q, size=%d\n", p.InlineData.MIMEType, len(p.InlineData.Data))
			case p.FunctionCall != nil:
				fmt.Fprintf(w, "FunctionCall Name=%v, Args=%v\n", p.FunctionCall.Name, p.FunctionCall.Args)
			case p.FunctionResponse != nil:
				fmt.Fprintf(w, "FunctionResponse Name=%v Response=%v\n", p.FunctionResponse.Name, p.FunctionResponse.Response)
			default:
				fmt.Fprintf(w, "unknown type")
			}
		}
	}
}
