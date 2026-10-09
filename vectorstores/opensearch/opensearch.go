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

// Store is a wrapper around the chromaGo API and client.
type Store struct {
	embedder embeddings.Embedder
	client   *opensearchgo.Client
}

var (
	// ErrNumberOfVectorDoesNotMatch when providing documents,
	// the number of vectors generated should be equal to the number of docs.
	ErrNumberOfVectorDoesNotMatch = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	// ErrAssertingMetadata Metadata is stored as string, trigger.
	ErrAssertingMetadata = errors.New(
		"couldn't assert metadata to map",
	)
)

// New creates and returns a vectorstore object for Opensearch
// and returns the `Store` object needed by the other accessors.
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

// AddDocuments adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
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
		id := opts.GenerateDocumentID(ctx, doc, ids)
		_, err := s.documentIndexing(ctx, id, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
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
	searchResults := searchResults{}
	if err := json.Unmarshal(body, &searchResults); err != nil {
		return output, fmt.Errorf("error unmarshalling search response body: %w %s", err, body)
	}

	for _, hit := range searchResults.Hits.Hits {
		if opts.ScoreThreshold > 0 && opts.ScoreThreshold > hit.Score {
			continue
		}

		output = append(output, schema.Document{
			PageContent: hit.Source.FieldsContent,
			Metadata:    hit.Source.FieldsMetadata,
			Score:       hit.Score,
		})
	}

	return output, nil
}
