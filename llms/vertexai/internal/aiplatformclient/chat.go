package aiplatformclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexschema"
	"google.golang.org/protobuf/types/known/structpb"
)

// CreateChat creates chat request.
func (p *PalmClient) CreateChat(ctx context.Context, model string, publisher string, r *vertexschema.ChatRequest) (*vertexschema.ChatResponse, error) { //nolint:lll
	responses, err := p.chat(ctx, model, publisher, r)
	if err != nil {
		return nil, err
	}
	chatResponse := &vertexschema.ChatResponse{}
	res := responses[0]
	value := res.GetStructValue().AsMap()
	candidates, ok := value["candidates"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %v", vertexschema.ErrMissingValue, "candidates")
	}
	for _, c := range candidates {
		candidate, ok := c.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a map[string]interface{}", vertexschema.ErrInvalidValue, "candidate")
		}
		author, ok := candidate["author"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a string", vertexschema.ErrInvalidValue, "author")
		}
		content, ok := candidate["content"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a string", vertexschema.ErrInvalidValue, "content")
		}
		chatResponse.Candidates = append(chatResponse.Candidates, vertexschema.ChatMessage{
			Author:  author,
			Content: content,
		})
	}
	return chatResponse, nil
}

func (p *PalmClient) chat(ctx context.Context, model string, publisher string, r *vertexschema.ChatRequest) ([]*structpb.Value, error) { //nolint:lll
	params := map[string]interface{}{
		"temperature": r.Temperature,
		"top_p":       r.TopP,
		"top_k":       r.TopK,
	}
	mergedParams, err := makeParams(params)
	if err != nil {
		return nil, err
	}

	messages := []interface{}{}
	for _, msg := range r.Messages {
		msgMap := map[string]interface{}{
			"author":  msg.Author,
			"content": msg.Content,
		}
		messages = append(messages, msgMap)
	}
	instance, err := structpb.NewStruct(map[string]interface{}{
		"context":  r.Context,
		"messages": messages,
	})
	if err != nil {
		return nil, err
	}
	instances := []*structpb.Value{
		structpb.NewStructValue(instance),
	}

	resp, err := p.client.Predict(ctx, &aiplatformpb.PredictRequest{
		Endpoint:   p.getModelPath(p.projectID, p.location, publisher, model),
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
