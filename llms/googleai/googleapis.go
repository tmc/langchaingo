// Package googleai implements Google AI integration using the new googleapis/go-genai SDK.
// This file provides a migration path from github.com/google/generative-ai-go to google.golang.org/genai.
package googleai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/internal/imageutil"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/genai"
)

var (
	ErrNoContentInGoogleApisResponse   = errors.New("no content in generation response")
	ErrUnknownPartInGoogleApisResponse = errors.New("unknown part type in generation response")
	ErrInvalidMimeTypeGoogleApis       = errors.New("invalid mime type on content")
)

// GoogleApisAI is the Google AI LLM implementation using the new googleapis genai SDK.
type GoogleApisAI struct {
	client             *genai.Client
	opts               *GoogleAIOptions
	CallbacksHandler   llms.CallbacksHandler
}

// NewGoogleApisAI creates a new Google AI LLM using the googleapis genai SDK.
func NewGoogleApisAI(ctx context.Context, opts ...GoogleAIOption) (*GoogleApisAI, error) {
	options := &GoogleAIOptions{
		DefaultModel: "gemini-1.5-flash",
		APIKey:       "",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.APIKey == "" {
		return nil, errors.New("Google AI API key is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: options.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create googleapis genai client: %w", err)
	}

	return &GoogleApisAI{
		client: client,
		opts:   options,
	}, nil
}

// Call implements the [llms.Model] interface.
func (g *GoogleApisAI) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, g, prompt, options...)
}

// GenerateContent implements the [llms.Model] interface.
func (g *GoogleApisAI) GenerateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{
		Model:       g.opts.DefaultModel,
		MaxTokens:   g.opts.MaxTokens,
		Temperature: g.opts.Temperature,
		TopP:        g.opts.TopP,
		TopK:        g.opts.TopK,
	}

	for _, opt := range options {
		opt(&opts)
	}

	// Get the generative model
	model := g.client.GenerativeModel(opts.Model)

	// Configure generation parameters
	if opts.Temperature > 0 {
		model.Temperature = &opts.Temperature
	}
	if opts.TopP > 0 {
		model.TopP = &opts.TopP
	}
	if opts.TopK > 0 {
		topK := int32(opts.TopK)
		model.TopK = &topK
	}
	if opts.MaxTokens > 0 {
		maxTokens := int32(opts.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}

	// Convert messages to googleapis format
	parts, err := g.messagesToParts(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		if g.CallbacksHandler != nil {
			g.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("googleapis genai generation failed: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, ErrNoContentInGoogleApisResponse
	}

	// Convert response to langchain format
	choices := make([]*llms.ContentChoice, 0, len(resp.Candidates))
	for _, candidate := range resp.Candidates {
		choice, err := g.candidateToContentChoice(candidate)
		if err != nil {
			return nil, fmt.Errorf("failed to convert candidate: %w", err)
		}
		choices = append(choices, choice)
	}

	response := &llms.ContentResponse{
		Choices: choices,
	}

	// Handle finish reason and usage information if available
	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		if candidate.FinishReason != nil {
			response.Choices[0].StopReason = g.convertFinishReason(*candidate.FinishReason)
		}
	}

	if resp.UsageMetadata != nil {
		response.Usage = llms.Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}
	}

	if g.CallbacksHandler != nil {
		g.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

// messagesToParts converts langchain messages to googleapis genai parts
func (g *GoogleApisAI) messagesToParts(ctx context.Context, messages []llms.MessageContent) ([]*genai.Part, error) {
	parts := make([]*genai.Part, 0)

	for _, message := range messages {
		for _, part := range message.Parts {
			switch p := part.(type) {
			case llms.TextPart:
				parts = append(parts, &genai.Part{
					Text: string(p),
				})
			case llms.BinaryPart:
				// Handle image content
				mimeType := p.MIMEType
				if mimeType == "" {
					return nil, ErrInvalidMimeTypeGoogleApis
				}

				var data []byte
				var err error

				// Handle different binary part types
				if len(p.Data) > 0 {
					data = p.Data
				} else if p.URL != "" {
					// Download image from URL
					data, err = imageutil.DownloadImageFromURL(ctx, p.URL)
					if err != nil {
						return nil, fmt.Errorf("failed to download image: %w", err)
					}
				} else {
					return nil, errors.New("binary part has no data or URL")
				}

				parts = append(parts, &genai.Part{
					InlineData: &genai.Blob{
						MIMEType: mimeType,
						Data:     data,
					},
				})
			case llms.ToolCallPart:
				// Handle tool calls if supported
				return nil, errors.New("tool calls not yet implemented for googleapis genai")
			default:
				return nil, fmt.Errorf("unsupported part type: %T", part)
			}
		}
	}

	return parts, nil
}

// candidateToContentChoice converts a googleapis candidate to langchain content choice
func (g *GoogleApisAI) candidateToContentChoice(candidate *genai.Candidate) (*llms.ContentChoice, error) {
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, ErrNoContentInGoogleApisResponse
	}

	var content strings.Builder
	var parts []llms.ContentPart

	for _, part := range candidate.Content.Parts {
		switch {
		case part.Text != "":
			content.WriteString(part.Text)
			parts = append(parts, llms.TextPart(part.Text))
		case part.InlineData != nil:
			// Handle binary responses (rare for text generation)
			parts = append(parts, llms.BinaryPart{
				MIMEType: part.InlineData.MIMEType,
				Data:     part.InlineData.Data,
			})
		case part.FunctionCall != nil:
			// Handle function calls if needed
			return nil, errors.New("function calls not yet implemented for googleapis genai")
		default:
			return nil, ErrUnknownPartInGoogleApisResponse
		}
	}

	choice := &llms.ContentChoice{
		Content:    content.String(),
		Parts:      parts,
		GenerationInfo: map[string]any{},
	}

	// Add safety ratings and other metadata
	if len(candidate.SafetyRatings) > 0 {
		safetyRatings := make(map[string]any)
		for _, rating := range candidate.SafetyRatings {
			safetyRatings[rating.Category.String()] = map[string]any{
				"probability": rating.Probability.String(),
				"blocked":     rating.Blocked,
			}
		}
		choice.GenerationInfo[SAFETY] = safetyRatings
	}

	if len(candidate.CitationMetadata) > 0 {
		citations := make([]map[string]any, len(candidate.CitationMetadata))
		for i, citation := range candidate.CitationMetadata {
			citations[i] = map[string]any{
				"start_index": citation.StartIndex,
				"end_index":   citation.EndIndex,
				"uri":         citation.URI,
				"title":       citation.Title,
				"license":     citation.License,
				"publication_date": citation.PublicationDate,
			}
		}
		choice.GenerationInfo[CITATIONS] = citations
	}

	return choice, nil
}

// convertFinishReason converts googleapis finish reason to langchain stop reason
func (g *GoogleApisAI) convertFinishReason(reason genai.FinishReason) string {
	switch reason {
	case genai.FinishReasonStop:
		return "stop"
	case genai.FinishReasonMaxTokens:
		return "length"
	case genai.FinishReasonSafety:
		return "content_filter"
	case genai.FinishReasonRecitation:
		return "content_filter"
	case genai.FinishReasonOther:
		return "other"
	default:
		return "unknown"
	}
}

// Close closes the googleapis genai client
func (g *GoogleApisAI) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// CreateEmbedding creates embeddings using the googleapis genai SDK
func (g *GoogleApisAI) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	if len(inputTexts) == 0 {
		return nil, errors.New("no input texts provided")
	}

	// Use text-embedding model
	model := g.client.EmbeddingModel("text-embedding-004")
	
	embeddings := make([][]float32, len(inputTexts))
	
	for i, text := range inputTexts {
		// Create embedding request
		resp, err := model.EmbedContent(ctx, &genai.Part{Text: text})
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}
		
		if resp == nil || resp.Embedding == nil || len(resp.Embedding.Values) == 0 {
			return nil, fmt.Errorf("no embedding returned for text %d", i)
		}
		
		embeddings[i] = resp.Embedding.Values
	}
	
	return embeddings, nil
}