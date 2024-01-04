package genaiclient

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexschema"
	"github.com/tmc/langchaingo/schema"
	"google.golang.org/api/option"
)

const (
	TextModelName = "gemini-pro"
	ChatModelName = "gemini-pro"
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

func (p GenAIClient) CreateCompletion(ctx context.Context, r *vertexschema.CompletionRequest) ([]*vertexschema.Completion, error) {
	model := p.client.GenerativeModel(r.Model)

	model.SetTemperature(float32(r.Temperature))
	model.SetMaxOutputTokens(int32(r.MaxTokens))
	model.SetTopP(float32(r.TopP))
	model.SetTopK(float32(r.TopK))
	model.GenerationConfig.StopSequences = r.StopSequences

	// Callers only know how to handle one response per prompt
	model.SetCandidateCount(1)

	completions := make([]*vertexschema.Completion, len(r.Prompts))
	for i, v := range r.Prompts {
		content, err := model.GenerateContent(ctx, genai.Text(v))
		if err != nil {
			return nil, err
		}

		result := content.Candidates[0].Content.Parts[0]
		value, ok := result.(genai.Text)
		if !ok {
			return nil, vertexschema.ErrInvalidReturnType
		}

		completions[i] = &vertexschema.Completion{
			Text: string(value),
		}
	}

	return completions, nil
}

// CreateChat creates chat request.
func (p GenAIClient) CreateChat(ctx context.Context, modelName string, publisher string, r *vertexschema.ChatRequest) (*vertexschema.ChatResponse, error) {
	model := p.client.GenerativeModel(modelName)

	model.SetTemperature(float32(r.Temperature))
	model.SetTopP(float32(r.TopP))
	model.SetTopK(float32(r.TopK))

	model.SetCandidateCount(r.CandidateCount)

	chat := model.StartChat()
	for _, message := range r.Messages {
		switch message.GetType() {
		case schema.ChatMessageTypeAI:
			chat.History = append(chat.History, &genai.Content{
				Role:  "model",
				Parts: []genai.Part{genai.Text(message.Content)},
			})
		case schema.ChatMessageTypeHuman:
			chat.History = append(chat.History, &genai.Content{
				Role:  "user",
				Parts: []genai.Part{genai.Text(message.Content)},
			})
		default:
			return nil, schema.ErrUnexpectedChatMessageType
		}
	}

	lastPart := chat.History[len(chat.History)-1].Parts
	chat.History = make([]*genai.Content, 0)

	chatResponse, err := chat.SendMessage(ctx, lastPart...)
	if err != nil {
		return nil, err
	}

	resp := &vertexschema.ChatResponse{Candidates: make([]vertexschema.ChatMessage, 0)}
	for _, generation := range chatResponse.Candidates {
		message := generation.Content.Parts[0].(genai.Text)
		chatMessage := vertexschema.ChatMessage{
			Content: string(message),
			Author:  "bot",
		}
		resp.Candidates = append(resp.Candidates, chatMessage)
	}

	return resp, nil
}
