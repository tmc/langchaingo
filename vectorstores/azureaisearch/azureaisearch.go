package azureaisearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/averikitsch/langchaingo/schema"
	"github.com/averikitsch/langchaingo/vectorstores"
	"github.com/google/uuid"
)

// Store is a wrapper to use azure AI search rest API.
type Store struct {
	azureAISearchEndpoint string
	azureAISearchAPIKey   string
	embedder              embeddings.Embedder
	client                *http.Client
}

var (
	// ErrNumberOfVectorDoesNotMatch when providing documents,
	// the number of vectors generated should be equal to the number of docs.
	ErrNumberOfVectorDoesNotMatch = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	// ErrAssertingMetadata SearchScore is stored as float64.
	ErrAssertingSearchScore = errors.New(
		"couldn't assert @search.score to float64",
	)
	// ErrAssertingMetadata Metadata is stored as string.
	ErrAssertingMetadata = errors.New(
		"couldn't assert metadata to string",
	)
	// ErrAssertingContent Content is stored as string.
	ErrAssertingContent = errors.New(
		"couldn't assert content to string",
	)
)

// New creates a vectorstore for azure AI search
// and returns the `Store` object needed by the other accessors.
func New(opts ...Option) (Store, error) {
	s := Store{
		client: http.DefaultClient,
	}

	if err := applyClientOptions(&s, opts...); err != nil {
		return s, err
	}

	return s, nil
}

var _ vectorstores.VectorStore = &Store{}

// AddDocuments adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
func (s *Store) AddDocuments(
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
		if err = s.UploadDocument(ctx, id, opts.NameSpace, doc.PageContent, vectors[i], doc.Metadata); err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s *Store) SimilaritySearch(
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
