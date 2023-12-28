package azureaisearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type Store struct {
	cognitiveSearchEndpoint string
	cognitiveSearchAPIKey   string
	embedder                embeddings.Embedder
	client                  *http.Client
}

var (
	ErrNumberOfVectorDoesNotMatch = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	ErrAssertingSearchScore = errors.New(
		"couldn't assert @search.score to float64",
	)
	ErrAssertingMetadata = errors.New(
		"couldn't assert metadata to string",
	)
	ErrAssertingContent = errors.New(
		"couldn't assert content to string",
	)
)

func New(opts ...Option) (Store, error) {
	s := Store{
		client: http.DefaultClient,
	}

	if err := applyClientOptions(&s, opts...); err != nil {
		return s, err
	}

	return s, nil
}

var _ vectorstores.VectorStore = Store{}

func (s Store) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) error {
	opts := s.getOptions(options...)

	texts := []string{}

	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	if len(vectors) != len(docs) {
		return ErrNumberOfVectorDoesNotMatch
	}

	for i, doc := range docs {
		if err = s.UploadDocument(ctx, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata); err != nil {
			return err
		}
	}

	return nil
}

func (s Store) SimilaritySearch(
	ctx context.Context, query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	queryVector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	payload := SearchDocumentsRequestInput{
		Vectors: []SearchDocumentsRequestInputVector{{
			Fields: "contentVector",
			Value:  queryVector,
			K:      numDocuments,
		}},
	}

	if filter, ok := opts.Filters.(string); ok {
		payload.Filter = filter
	}

	searchResults := SearchDocumentsRequestOuput{}
	if err := s.SearchDocuments(ctx, opts.NameSpace, payload, &searchResults); err != nil {
		return nil, err
	}

	output := []schema.Document{}
	for _, searchResult := range searchResults.Value {
		doc, err := assertResultValues(searchResult)
		if err != nil {
			return output, err
		}

		if opts.ScoreThreshold > 0 && opts.ScoreThreshold > doc.Score {
			continue
		}

		output = append(output, *doc)
	}

	return output, nil
}

func assertResultValues(searchResult map[string]interface{}) (*schema.Document, error) {
	var score float32
	if scoreFloat64, ok := searchResult["@search.score"].(float64); ok {
		score = float32(scoreFloat64)
	} else {
		return nil, ErrAssertingSearchScore
	}

	metadata := map[string]interface{}{}
	if resultMetadata, ok := searchResult["metadata"].(string); ok {
		if err := json.Unmarshal([]byte(resultMetadata), &metadata); err != nil {
			return nil, fmt.Errorf("couldn't unmarshall metadata %w", err)
		}
	} else {
		return nil, ErrAssertingMetadata
	}

	var pageContent string
	var ok bool
	if pageContent, ok = searchResult["content"].(string); !ok {
		return nil, ErrAssertingContent
	}

	return &schema.Document{
		PageContent: pageContent,
		Metadata:    metadata,
		Score:       score,
	}, nil
}
