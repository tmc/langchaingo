package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	opensearchgo "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type Store struct {
	embedder embeddings.Embedder
	client   *opensearchgo.Client
}

var (
	ErrNumberOfVectorDoesNotMatch = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
)

func New(client *opensearchgo.Client, opts ...Option) (Store, error) {
	s := Store{
		client: client,
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

	fmt.Printf("texts: %v\n", len(texts))

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}
	fmt.Printf("vectors: %v\n", vectors)
	fmt.Printf("len(vectors): %v\n", len(vectors))
	fmt.Printf("len(docs): %v\n", len(docs))
	if len(vectors) != len(docs) {
		return ErrNumberOfVectorDoesNotMatch
	}
	fmt.Printf("opts: %v\n", opts)
	for i, doc := range docs {
		if _, err = s.DocumentIndexing(ctx, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata); err != nil {
			return err
		}
	}

	return nil
}

func (s Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	queryVector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	searchPayload := map[string]interface{}{
		"size": numDocuments,
		"query": map[string]interface{}{
			"knn": map[string]interface{}{
				"vector_field": map[string]interface{}{
					"vector": queryVector,
					"k":      numDocuments,
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(searchPayload); err != nil {
		return nil, fmt.Errorf("error encoding index schema to json buffer %w", err)
	}

	search := opensearchapi.SearchRequest{
		Index: []string{opts.NameSpace},
		Body:  buf,
	}
	output := []schema.Document{}
	searchResponse, err := search.Do(context.Background(), s.client)
	if err != nil {
		return output, err
	}
	fmt.Printf("searchResponse: %v\n", searchResponse)

	return []schema.Document{}, nil
}
