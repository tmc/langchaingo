package redisvector

import (
	"context"
	"errors"

	"github.com/0xDezzy/langchaingo/embeddings"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/vectorstores"
)

const (
	// same as langchain python version.
	defaultContentFieldKey       = "content"        // page_content
	defaultContentVectorFieldKey = "content_vector" // vector
	defaultDistanceFieldKey      = "distance"       // distance
)

var (
	ErrEmptyIndexName         = errors.New("empty redis index name")
	ErrNotExistedIndex        = errors.New("redis index name does not exist")
	ErrInvalidEmbeddingVector = errors.New("embedding vector error")
	ErrInvalidScoreThreshold  = errors.New("score threshold must be between 0 and 1")
	ErrInvalidFilters         = errors.New("invalid filters")
)

// Store is a wrapper around the redis client.
type Store struct {
	embedder               embeddings.Embedder
	client                 RedisClient
	redisURL               string
	indexName              string
	createIndexIfNotExists bool
	indexSchema            *IndexSchema
	schemaGenerator        *schemaGenerator
}

var _ vectorstores.VectorStore = &Store{}

// New creates a new Store with options.
func New(ctx context.Context, opts ...Option) (*Store, error) {
	var s *Store
	var err error

	s, err = applyClientOptions(opts...)
	if err != nil {
		return nil, err
	}

	client, err := NewRueidisClient(s.redisURL)
	if err != nil {
		return nil, err
	}

	s.client = client

	if !s.client.CheckIndexExists(ctx, s.indexName) {
		if !s.createIndexIfNotExists {
			return nil, ErrNotExistedIndex
		} else if s.indexSchema != nil {
			// create index with input schema
			if err := s.client.CreateIndexIfNotExists(ctx, s.indexName, s.indexSchema); err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

// AddDocuments adds the text and metadata from the documents to the redis associated with 'Store'.
// and returns the ids of the added documents.
// Note: currently save documents with Hset command
// return `docIDs` that prefix with `doc:{index_name}`
//
//	if doc.metadata has `keys` or `ids` field, the docId will use `keys` or `ids` value
//	if not, the docId is uuid string
func (s *Store) AddDocuments(ctx context.Context, docs []schema.Document, _ ...vectorstores.Option) ([]string, error) {
	err := s.appendDocumentsWithVectors(ctx, docs)
	if err != nil {
		return nil, err
	}

	indexSchema, err := generateSchemaWithMetadata(docs[0].Metadata)
	if err != nil {
		return nil, err
	}

	if s.indexSchema == nil {
		s.indexSchema = indexSchema
	}

	if s.createIndexIfNotExists && !s.client.CheckIndexExists(ctx, s.indexName) {
		if err := s.client.CreateIndexIfNotExists(ctx, s.indexName, indexSchema); err != nil {
			return nil, err
		}
	}

	docIDs, err := s.client.AddDocsWithHash(ctx, getPrefix(s.indexName), docs)
	if err != nil {
		return nil, err
	}

	return docIDs, nil
}

// SimilaritySearch similarity search docs with `ScoreThreshold` `Filters` `Embedder`
// Support options:
//
//	WithScoreThreshold:
//	WithFilters: filter string should match redis search pre-filter query pattern.(eg: @title:Dune)
//		ref: https://redis.io/docs/latest/develop/interact/search-and-query/advanced-concepts/vectors/#pre-filter-query-attributes-hybrid-approach
//	WithEmbedder: if set, it will embed query string with this embedder; otherwise embed with vector's embedder
//
// ref: https://redis.io/docs/latest/develop/interact/search-and-query/advanced-concepts/vectors/#pre-filter-query-attributes-hybrid-approach
func (s *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	filter, err := s.getFilters(opts)
	if err != nil {
		return nil, err
	}
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	embedderData, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	searchOpts := []SearchOption{WithScoreThreshold(scoreThreshold), WithOffsetLimit(0, numDocuments), WithPreFilters(filter)}
	if s.indexSchema != nil {
		var keys []string
		for k := range s.indexSchema.MetadataKeys() {
			keys = append(keys, k)
		}
		searchOpts = append(searchOpts, WithReturns(keys))
	}

	search, err := NewIndexVectorSearch(
		s.indexName,
		embedderData,
		searchOpts...,
	)
	if err != nil {
		return nil, err
	}

	_, docs, err := s.client.Search(ctx, *search)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (s *Store) DropIndex(ctx context.Context, index string, deleteDocuments bool) error {
	if !s.client.CheckIndexExists(ctx, index) {
		return ErrNotExistedIndex
	}
	return s.client.DropIndex(ctx, index, deleteDocuments)
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

// getFilters return metadata filters.
func (s Store) getFilters(opts vectorstores.Options) (string, error) {
	if opts.Filters != nil {
		if filters, ok := opts.Filters.(string); ok {
			return filters, nil
		}
		return "", ErrInvalidFilters
	}
	return "", nil
}

// append content & content_vector into doc.Metadata.
func (s Store) appendDocumentsWithVectors(ctx context.Context, docs []schema.Document) error {
	if len(docs) == 0 {
		return nil
	}

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}
	if len(vectors) != len(docs) {
		return ErrInvalidEmbeddingVector
	}

	// append content & content_vector info metadata
	for i := range docs {
		if docs[i].Metadata == nil {
			docs[i].Metadata = map[string]any{}
		}
		docs[i].Metadata[defaultContentFieldKey] = docs[i].PageContent
		docs[i].Metadata[defaultContentVectorFieldKey] = vectors[i]
	}

	// vectorDimension := len(vectors[0])
	return nil
}
