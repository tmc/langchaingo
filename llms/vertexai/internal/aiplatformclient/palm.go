package aiplatformclient

import (
	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexschema"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	EmbeddingModelName = "textembedding-gecko"
	TextModelName      = "text-bison"
	ChatModelName      = "chat-bison"
)

type PalmClient struct {
	client *aiplatform.PredictionClient

	projectID string
	location  string
}

func New(ctx context.Context, projectID string, location string, option ...option.ClientOption) (PalmClient, error) {
	gi := PalmClient{}
	client, err := aiplatform.NewPredictionClient(ctx, option...)
	if err != nil {
		return gi, err
	}

	gi.client = client
	gi.projectID = projectID
	gi.location = location

	return gi, nil
}

func (p PalmClient) CreateCompletion(ctx context.Context, r *vertexschema.CompletionRequest) ([]*vertexschema.Completion, error) {
	params := map[string]interface{}{
		"maxOutputTokens": r.MaxTokens,
		"temperature":     r.Temperature,
		"top_p":           r.TopP,
		"top_k":           r.TopK,
		"stopSequences":   convertArray(r.StopSequences),
	}
	predictions, err := p.BatchPredict(ctx, TextModelName, r.Prompts, params)
	if err != nil {
		return nil, err
	}
	completions := []*vertexschema.Completion{}
	for _, p := range predictions {
		value := p.GetStructValue().AsMap()
		text, ok := value["content"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v", vertexschema.ErrMissingValue, "content")
		}
		completions = append(completions, &vertexschema.Completion{
			Text: text,
		})
	}
	return completions, nil
}

func (p PalmClient) BatchPredict(ctx context.Context, model string, prompts []string, params map[string]interface{}) ([]*structpb.Value, error) { //nolint:lll
	mergedParams, err := makeParams(params)
	if err != nil {
		return nil, err
	}

	instances := []*structpb.Value{}
	for _, prompt := range prompts {
		content, _ := structpb.NewStruct(map[string]interface{}{
			"content": prompt,
		})
		instances = append(instances, structpb.NewStructValue(content))
	}
	resp, err := p.client.Predict(ctx, &aiplatformpb.PredictRequest{
		Endpoint:   p.getModelPath(p.projectID, "us-central1", "google", model),
		Instances:  instances,
		Parameters: structpb.NewStructValue(mergedParams),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.GetPredictions()) == 0 {
		return nil, vertexschema.ErrEmptyResponse
	}
	return resp.GetPredictions(), nil
}

func (p PalmClient) getModelPath(project, location, publisher, model string) string {
	// POST https://{REGION}-aiplatform.googleapis.com/v1/projects/{PROJECT_ID}/locations/{REGION}/publishers/google/models/gemini-pro:streamGenerateContent
	return fmt.Sprintf("projects/%s/locations/%s/publishers/%s/models/%s", project, location, publisher, model)
}
