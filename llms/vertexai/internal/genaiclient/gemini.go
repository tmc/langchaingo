package genaiclient

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"github.com/tmc/langchaingo/llms/vertexai/internal/schema"
	"google.golang.org/api/option"
)

const (
	EmbeddingModelName = "textembedding-gecko"
	TextModelName      = "gemini-pro"
	ChatModelName      = "gemini-pro"
)

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContents struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float64 `json:"temperature"`
	TopP            int     `json:"topP"`
}

type GeminiCall struct {
	Contents         []GeminiContents       `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generation_config"`
}

type GenAIClient struct {
	client *genai.Client
}

func New(ctx context.Context, projectID string, location string, option ...option.ClientOption) (GenAIClient, error) {
	gi := GenAIClient{}
	client, err := genai.NewClient(ctx, projectID, location, option...)
	if err != nil {
		return gi, err
	}

	gi.client = client
	return gi, nil
}

func (p GenAIClient) CreateCompletion(ctx context.Context, r *schema.CompletionRequest) ([]*schema.Completion, error) {
	model := p.client.GenerativeModel(r.Model)

	model.SetTemperature(float32(r.Temperature))
	model.SetMaxOutputTokens(int32(r.MaxTokens))
	model.SetTopP(float32(r.TopP))
	model.SetTopK(float32(r.TopK))
	model.GenerationConfig.StopSequences = r.StopSequences

	// Callers only know how to handle one response per prompt
	model.SetCandidateCount(1)

	completions := make([]*schema.Completion, len(r.Prompts))
	for i, v := range r.Prompts {
		content, err := model.GenerateContent(ctx, genai.Text(v))
		if err != nil {
			return nil, err
		}

		result := content.Candidates[0].Content.Parts[0]
		value, ok := result.(genai.Text)
		if !ok {
			return nil, schema.ErrInvalidReturnType
		}

		completions[i] = &schema.Completion{
			Text: string(value),
		}
	}

	return completions, nil
}
