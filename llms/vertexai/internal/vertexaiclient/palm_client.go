package vertexaiclient

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/tmc/langchaingo/schema"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	defaultAPIEndpoint = "us-central1-aiplatform.googleapis.com:443"
	defaultLocation    = "us-central1"
	defaultPublisher   = "google"
)

var (
	// ErrMissingValue is returned when a value is missing.
	ErrMissingValue = errors.New("missing value")
	// ErrInvalidValue is returned when a value is invalid.
	ErrInvalidValue = errors.New("invalid value")
)

var defaultParameters = map[string]interface{}{ //nolint:gochecknoglobals
	"temperature":     0.2, //nolint:gomnd
	"maxOutputTokens": 256, //nolint:gomnd
	"topP":            0.8, //nolint:gomnd
	"topK":            40,  //nolint:gomnd
}

const (
	embeddingModelName = "textembedding-gecko"
	TextModelName      = "text-bison"
	ChatModelName      = "chat-bison"

	defaultMaxConns = 4
)

// PaLMClient represents a Vertex AI based PaLM API client.
type PaLMClient struct {
	client    *aiplatform.PredictionClient
	projectID string
}

// New returns a new Vertex AI based PaLM API client.
func New(projectID string, opts ...option.ClientOption) (*PaLMClient, error) {
	numConns := runtime.GOMAXPROCS(0)
	if numConns > defaultMaxConns {
		numConns = defaultMaxConns
	}
	o := []option.ClientOption{
		option.WithGRPCConnectionPool(numConns),
		option.WithEndpoint(defaultAPIEndpoint),
	}
	o = append(o, opts...)

	ctx := context.Background()
	client, err := aiplatform.NewPredictionClient(ctx, o...)
	if err != nil {
		return nil, err
	}
	return &PaLMClient{
		client:    client,
		projectID: projectID,
	}, nil
}

// ErrEmptyResponse is returned when the OpenAI API returns an empty response.
var ErrEmptyResponse = errors.New("empty response")

// CompletionRequest is a request to create a completion.
type CompletionRequest struct {
	Prompts     []string `json:"prompts"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        int      `json:"top_p,omitempty"`
	TopK        int      `json:"top_k,omitempty"`
}

// Completion is a completion.
type Completion struct {
	Text string `json:"text"`
}

// CreateCompletion creates a completion.
func (c *PaLMClient) CreateCompletion(ctx context.Context, r *CompletionRequest) ([]*Completion, error) {
	params := map[string]interface{}{
		"maxOutputTokens": r.MaxTokens,
		"temperature":     r.Temperature,
		"top_p":           r.TopP,
		"top_k":           r.TopK,
	}
	predictions, err := c.batchPredict(ctx, TextModelName, r.Prompts, params)
	if err != nil {
		return nil, err
	}
	completions := []*Completion{}
	for _, p := range predictions {
		value := p.GetStructValue().AsMap()
		text, ok := value["content"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v", ErrMissingValue, "content")
		}
		completions = append(completions, &Completion{
			Text: text,
		})
	}
	return completions, nil
}

// EmbeddingRequest is a request to create an embedding.
type EmbeddingRequest struct {
	Input []string `json:"input"`
}

// CreateEmbedding creates embeddings.
func (c *PaLMClient) CreateEmbedding(ctx context.Context, r *EmbeddingRequest) ([][]float64, error) {
	params := map[string]interface{}{}
	responses, err := c.batchPredict(ctx, embeddingModelName, r.Input, params)
	if err != nil {
		return nil, err
	}

	embeddings := [][]float64{}
	for _, res := range responses {
		value := res.GetStructValue().AsMap()
		embedding, ok := value["embeddings"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v", ErrMissingValue, "embeddings")
		}
		values, ok := embedding["values"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v", ErrMissingValue, "values")
		}
		floatValues := []float64{}
		for _, v := range values {
			val, ok := v.(float64)
			if !ok {
				return nil, fmt.Errorf("%w: %v is not a float64", ErrInvalidValue, "value")
			}
			floatValues = append(floatValues, val)
		}
		embeddings = append(embeddings, floatValues)
	}
	return embeddings, nil
}

// ChatRequest is a request to create an embedding.
type ChatRequest struct {
	Context        string         `json:"context"`
	Messages       []*ChatMessage `json:"messages"`
	Temperature    float64        `json:"temperature,omitempty"`
	TopP           int            `json:"top_p,omitempty"`
	TopK           int            `json:"top_k,omitempty"`
	CandidateCount int            `json:"candidate_count,omitempty"`
}

// ChatMessage is a message in a chat.
type ChatMessage struct {
	// The content of the message.
	Content string `json:"content"`
	// The name of the author of this message. user or bot
	Author string `json:"author,omitempty"`
}

// Statically assert that the types implement the interface.
var _ schema.ChatMessage = ChatMessage{}

// GetType returns the type of the message.
func (m ChatMessage) GetType() schema.ChatMessageType {
	switch m.Author {
	case "user":
		return schema.ChatMessageTypeHuman
	default:
		return schema.ChatMessageTypeAI
	}
}

// GetText returns the text of the message.
func (m ChatMessage) GetContent() string {
	return m.Content
}

// ChatResponse is a response to a chat request.
type ChatResponse struct {
	Candidates []ChatMessage
}

// CreateChat creates chat request.
func (c *PaLMClient) CreateChat(ctx context.Context, r *ChatRequest) (*ChatResponse, error) {
	responses, err := c.chat(ctx, r)
	if err != nil {
		return nil, err
	}
	chatResponse := &ChatResponse{}
	res := responses[0]
	value := res.GetStructValue().AsMap()
	candidates, ok := value["candidates"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrMissingValue, "candidates")
	}
	for _, c := range candidates {
		candidate, ok := c.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a map[string]interface{}", ErrInvalidValue, "candidate")
		}
		author, ok := candidate["author"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a string", ErrInvalidValue, "author")
		}
		content, ok := candidate["content"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %v is not a string", ErrInvalidValue, "content")
		}
		chatResponse.Candidates = append(chatResponse.Candidates, ChatMessage{
			Author:  author,
			Content: content,
		})
	}
	return chatResponse, nil
}

func mergeParams(defaultParams, params map[string]interface{}) *structpb.Struct {
	mergedParams := map[string]interface{}{}
	for paramKey, paramValue := range defaultParameters {
		mergedParams[paramKey] = paramValue
	}
	for paramKey, paramValue := range params {
		switch value := paramValue.(type) {
		case float64:
			if value != 0 {
				mergedParams[paramKey] = value
			}
		case int:
		case int32:
		case int64:
			if value != 0 {
				mergedParams[paramKey] = value
			}
		}
	}
	smergedParams, err := structpb.NewStruct(mergedParams)
	if err != nil {
		smergedParams, _ = structpb.NewStruct(defaultParams)
		return smergedParams
	}
	return smergedParams
}

func (c *PaLMClient) batchPredict(ctx context.Context, model string, prompts []string, params map[string]interface{}) ([]*structpb.Value, error) { //nolint:lll
	mergedParams := mergeParams(defaultParameters, params)
	instances := []*structpb.Value{}
	for _, prompt := range prompts {
		content, _ := structpb.NewStruct(map[string]interface{}{
			"content": prompt,
		})
		instances = append(instances, structpb.NewStructValue(content))
	}
	resp, err := c.client.Predict(ctx, &aiplatformpb.PredictRequest{
		Endpoint:   c.projectLocationPublisherModelPath(c.projectID, "us-central1", "google", model),
		Instances:  instances,
		Parameters: structpb.NewStructValue(mergedParams),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Predictions) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp.Predictions, nil
}

func (c *PaLMClient) chat(ctx context.Context, r *ChatRequest) ([]*structpb.Value, error) {
	params := map[string]interface{}{
		"temperature": r.Temperature,
		"top_p":       r.TopP,
		"top_k":       r.TopK,
	}
	mergedParams := mergeParams(defaultParameters, params)
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
		Endpoint:   c.projectLocationPublisherModelPath(c.projectID, "us-central1", "google", ChatModelName),
		Instances:  instances,
		Parameters: structpb.NewStructValue(mergedParams),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Predictions) == 0 {
		return nil, ErrEmptyResponse
	}
	return resp.Predictions, nil
}

func (c *PaLMClient) projectLocationPublisherModelPath(projectID, location, publisher, model string) string {
	return fmt.Sprintf("projects/%s/locations/%s/publishers/%s/models/%s", projectID, location, publisher, model)
}
