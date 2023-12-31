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

	Model          string
	Publisher      string
	EmbeddingModel string
}

func newBase(ctx context.Context, model string, opts options) (*baseLLM, error) {
	if len(opts.projectID) == 0 {
		return nil, ErrMissingProjectID
	}

	client, err := common.New(ctx, opts.projectID, opts.clientOptions...)
	if err != nil {
		return nil, err
	}

	return &baseLLM{
		client:         client,
		Model:          model,
		EmbeddingModel: opts.embeddingModel,

		Publisher: opts.publisher,
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
