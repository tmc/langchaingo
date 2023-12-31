package aiplatformclient

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms/vertexai/internal/schema"
	"google.golang.org/protobuf/types/known/structpb"
)

// CreateChat creates chat request.
func (c *PalmClient) CreateChat(ctx context.Context, model string, publisher string, r *schema.ChatRequest) (*schema.ChatResponse, error) {
	responses, err := c.chat(ctx, model, publisher, r)
	if err != nil {
		return nil, err
	}
	chatResponse := &schema.ChatResponse{}
	res := responses[0]
	value := res.GetStructValue().AsMap()
	candidates, ok := value["candidates"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %v", schema.ErrMissingValue, "candidates")
	}
	for _, c := range candidates {
		candidate, ok := c.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a map[string]interface{}", schema.ErrInvalidValue, "candidate")
		}
		author, ok := candidate["author"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a string", schema.ErrInvalidValue, "author")
		}
		content, ok := candidate["content"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a string", schema.ErrInvalidValue, "content")
		}
		chatResponse.Candidates = append(chatResponse.Candidates, schema.ChatMessage{
			Author:  author,
			Content: content,
		})
	}
	return chatResponse, nil
}

func (c *PalmClient) chat(ctx context.Context, model string, publisher string, r *schema.ChatRequest) ([]*structpb.Value, error) {
	params := map[string]interface{}{
		"temperature": r.Temperature,
		"top_p":       r.TopP,
		"top_k":       r.TopK,
	}
	mergedParams := mergeParams(schema.DefaultParameters, params)

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

	resp, err := c.client.Predict(ctx, &aiplatformpb.PredictRequest{
		Endpoint:   c.getModelPath(c.projectID, c.location, publisher, model),
		Instances:  instances,
		Parameters: structpb.NewStructValue(mergedParams),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.GetPredictions()) == 0 {
		return nil, schema.ErrEmptyResponse
	}
	return resp.GetPredictions(), nil
}
