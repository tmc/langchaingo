package googleai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/option"
)

// GoogleAI is a type that represents a Google AI API client.
type GoogleAI struct {
	client *genai.Client
}

var (
	_ llms.Model = GoogleAI{}

	ErrNoContentInResponse   = errors.New("no content in generation response")
	ErrUnknownPartInResponse = errors.New("unknown part type in generation response")
)

const (
	CITATIONS = "citations"
	SAFETY    = "safety"
)

// New creates a new GoogleAI struct.
func New(ctx context.Context, options ...option.ClientOption) (*GoogleAI, error) {
	gi := &GoogleAI{}

	client, err := genai.NewClient(ctx, options...)
	if err != nil {
		return gi, err
	}

	gi.client = client
	return gi, nil
}

// GenerateContent calls the LLM with the provided parts.
func (g GoogleAI) GenerateContent(ctx context.Context, parts []llms.ContentPart, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint:lll
	opts := defaultCallOptions

	for _, opt := range options {
		opt(&opts)
	}

	model := g.client.GenerativeModel(opts.Model)

	content := make([]genai.Part, 0, len(parts))
	for _, part := range parts {
		c, err := convertPart(part)
		if err != nil {
			return nil, err
		}

		content = append(content, c)
	}

	resp, err := model.GenerateContent(ctx, content...)
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, ErrNoContentInResponse
	}

	var contentResponse llms.ContentResponse
	for _, candidate := range resp.Candidates {
		c, err := convertCandidate(candidate)
		if err != nil {
			return nil, err
		}
		contentResponse.Choices = append(contentResponse.Choices, c)
	}

	return &contentResponse, nil
}

// downloadBytes downloads the content from the given URL and returns it as a *genai.Blob.
func downloadBytes(url string) (*genai.Blob, error) {
	resp, err := http.Get(url) //nolint
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image from url: %w", err)
	}
	defer resp.Body.Close()

	urlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image bytes: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")

	return &genai.Blob{MIMEType: mimeType, Data: urlData}, nil
}

// convertPart converts langchain parts to google genai parts.
func convertPart(part llms.ContentPart) (genai.Part, error) {
	var out genai.Part
	var err error

	switch p := part.(type) {
	case llms.TextContent:
		out = genai.Text(p.Text)
	case llms.BinaryContent:
		out = genai.Blob{MIMEType: p.MIMEType, Data: p.Data}
	case llms.ImageURLContent:
		out, err = downloadBytes(p.URL)
	}

	return out, err
}

// convertCandidate converts a genai.Candidate to a llms.ContentChoice.
func convertCandidate(candidate *genai.Candidate) (*llms.ContentChoice, error) {
	buf := strings.Builder{}

	for _, part := range candidate.Content.Parts {
		if v, ok := part.(genai.Text); ok {
			_, err := buf.WriteString(string(v))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, ErrUnknownPartInResponse
		}
	}

	metadata := make(map[string]any)
	metadata[CITATIONS] = candidate.CitationMetadata
	metadata[SAFETY] = candidate.SafetyRatings

	return &llms.ContentChoice{
		Content:        buf.String(),
		StopReason:     candidate.FinishReason.String(),
		GenerationInfo: metadata,
	}, nil
}
