package bedrock

import (
	"context"
	"errors"
	"strings"

	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// Bedrock is the embedder used generate text embeddings through Amazon Bedrock.
type Bedrock struct {
	ModelID       string
	client        *bedrockruntime.Client
	StripNewLines bool
	BatchSize     int
}

// NewBedrock returns a new embeddings.Embedder that uses Amazon Bedrock to generate embeddings.
func NewBedrock(opts ...Option) (*Bedrock, error) {
	v, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func getProvider(modelID string) string {
	return strings.Split(modelID, ".")[0]
}

// EmbedDocuments implements embeddings.Embedder
// and generates embeddings for the supplied texts.
func (b *Bedrock) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, b.StripNewLines),
		b.BatchSize,
	)
	provider := getProvider(b.ModelID)

	allEmbeds := make([][]float32, 0, len(texts))
	var embeddings [][]float32
	var err error

	for _, batch := range batchedTexts {
		switch provider {
		case "amazon":
			embeddings, err = FetchAmazonTextEmbeddings(ctx, b.client, b.ModelID, batch)
		case "cohere":
			embeddings, err = FetchCohereTextEmbeddings(ctx, b.client, b.ModelID, batch, CohereInputTypeText)
		default:
			err = errors.New("unsupported text embedding provider: " + provider)
		}

		if err != nil {
			return nil, err
		}
		allEmbeds = append(allEmbeds, embeddings...)
	}
	return allEmbeds, nil
}

// EmbedQuery implements embeddings.Embedder
// and generates an embedding for the supplied text.
func (b *Bedrock) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	var embeddings [][]float32
	var err error

	switch provider := getProvider(b.ModelID); provider {
	case "amazon":
		embeddings, err = FetchAmazonTextEmbeddings(ctx, b.client, b.ModelID, []string{text})
	case "cohere":
		embeddings, err = FetchCohereTextEmbeddings(ctx, b.client, b.ModelID, []string{text}, CohereInputTypeQuery)
	default:
		err = errors.New("unsupported text embedding provider: " + provider)
	}

	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

var _ embeddings.Embedder = &Bedrock{}
