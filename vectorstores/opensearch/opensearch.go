package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

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
	fmt.Printf("opts: %v\n", opts)
	for i, doc := range docs {
		_, err := s.DocumentIndexing(ctx, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata)
		if err != nil {
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
	searchResponse, err := search.Do(context.Background(), s.client)
	if err != nil {
		fmt.Printf("search.Do err: %v\n", err)
		return output, err
	}

	body, err := io.ReadAll(searchResponse.Body)
	if err != nil {
		return output, fmt.Errorf("error reading search reponse body: %w", err)
	}
	searchResults := SearchResults{}
	if err := json.Unmarshal(body, &searchResults); err != nil {
		return output, fmt.Errorf("error unmarshalling search reponse body: %w %s", err, body)
	}

	for _, hit := range searchResults.Hits.Hits {
		fmt.Printf("score: %v | content %s \n", hit.Score, hit.Source.FieldsContent)
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
			Score:       float32(hit.Score),
		})
	}

	return output, nil
}
