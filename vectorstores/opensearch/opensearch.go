package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
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
	ErrAssertingMetadata = errors.New(
		"couldn't assert metadata to map",
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

func (s Store) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)
	ids := []string{}
	texts := []string{}

	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return ids, err
	}

	if len(vectors) != len(docs) {
		return ids, ErrNumberOfVectorDoesNotMatch
	}

	for i, doc := range docs {
		id := uuid.NewString()
		_, err := s.DocumentIndexing(ctx, id, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
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
				"contentVector": map[string]interface{}{
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
	searchResponse, err := search.Do(ctx, s.client)
	if err != nil {
		return output, fmt.Errorf("search.Do err: %w", err)
	}

	body, err := io.ReadAll(searchResponse.Body)
	if err != nil {
		return output, fmt.Errorf("error reading search response body: %w", err)
	}
	searchResults := SearchResults{}
	if err := json.Unmarshal(body, &searchResults); err != nil {
		return output, fmt.Errorf("error unmarshalling search response body: %w %s", err, body)
	}

	for _, hit := range searchResults.Hits.Hits {
		if opts.ScoreThreshold > 0 && opts.ScoreThreshold > hit.Score {
			continue
		}

		metadata := map[string]interface{}{}
		if err := json.Unmarshal([]byte(hit.Source.FieldsMetadata), &metadata); err != nil {
			return output, ErrAssertingMetadata
		}

		output = append(output, schema.Document{
			PageContent: hit.Source.FieldsContent,
			Metadata:    metadata,
			Score:       hit.Score,
		})
	}

	return output, nil
}
