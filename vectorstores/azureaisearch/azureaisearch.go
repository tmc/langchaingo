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

func New(ctx context.Context, opts ...Option) (Store, error) {
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
		fmt.Printf("err embedding documents: %v\n", err)
		return err
	}

	if len(vectors) != len(docs) {
		return errors.New(
			"number of vectors from embedder does not match number of documents",
		)
	}

	for i, doc := range docs {
		if err = s.UploadDocument(ctx, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata); err != nil {
			fmt.Printf("error uploading document to vector: %v\n", err)
			return err
		}
	}

	return nil
}

func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
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
		fmt.Printf("error searching documents vector: %v\n", err)
		return nil, err
	}
	output := []schema.Document{}
	for _, searchResult := range searchResults.Value {
		score := float32(searchResult["@search.score"].(float64))
		fmt.Printf("result: %v | score: %v\n", searchResult["content"].(string), score)
		if opts.ScoreThreshold > 0 && opts.ScoreThreshold > score {
			continue
		}

		metadata := map[string]interface{}{}
		if resultMetadata, ok := searchResult["metadata"].(string); ok {
			if err := json.Unmarshal([]byte(resultMetadata), &metadata); err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
		}

		output = append(output, schema.Document{
			PageContent: searchResult["content"].(string),
			Metadata:    metadata,
			Score:       score,
		})
	}

	return output, nil
}
