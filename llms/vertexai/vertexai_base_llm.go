package vertexai

import (
	"context"
	"errors"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/common"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingProjectID         = errors.New("missing the GCP Project ID, set it in the GOOGLE_CLOUD_PROJECT environment variable") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrNotImplemented           = errors.New("not implemented")
)

type baseLLM struct {
	CallbacksHandler callbacks.Handler
	client           *common.VertexClient

	// These are the defaults to use if not on the request
	Model     string
	Publisher string

	EmbeddingModel string
}

var defaultCallOptions = map[string]interface{}{ //nolint:gochecknoglobals
	"temperature":     0.2, //nolint:gomnd
	"maxOutputTokens": 256, //nolint:gomnd
	"topP":            0.8, //nolint:gomnd
	"topK":            40,  //nolint:gomnd
}

func newBase(ctx context.Context, opts Options) (*baseLLM, error) {
	if len(opts.ProjectID) == 0 {
		return nil, ErrMissingProjectID
	}

	client, err := common.New(ctx, opts.ConnectOptions, opts.ClientOptions...)
	if err != nil {
		return nil, err
	}

	return &baseLLM{
		client:         client,
		Model:          opts.model,
		EmbeddingModel: opts.embeddingModel,

		Publisher: opts.Publisher,
	}, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *baseLLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &common.EmbeddingRequest{
		Input: inputTexts,
		Model: o.EmbeddingModel,
	})
	if err != nil {
		return [][]float32{}, err
	}

	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}

	return embeddings, nil
}

func (o *baseLLM) GetNumTokens(text string) int {
	return llms.CountTokens(o.Model, text)
}

func (o *baseLLM) setDefaultCallOptions(opts *llms.CallOptions) {
	if len(opts.Model) == 0 {
		opts.Model = o.Model
	}

	if opts.MaxTokens == 0 {
		v, _ := defaultCallOptions["maxOutputTokens"].(int)
		opts.MaxTokens = v
	}

	if opts.Temperature == 0 {
		v, _ := defaultCallOptions["temperature"].(float64)
		opts.Temperature = v
	}

	if opts.TopP == 0 {
		v, _ := defaultCallOptions["topP"].(float64)
		opts.TopP = v
	}

	if opts.TopK == 0 {
		v, _ := defaultCallOptions["topK"].(int)
		opts.TopK = v
	}
}
