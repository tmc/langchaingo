// Package milvus provides a vectorstore implementation for Milvus.
//
// Deprecated: This package uses github.com/milvus-io/milvus-sdk-go/v2 which has been
// archived by the Milvus maintainers on March 21, 2025. The new SDK is available at
// github.com/milvus-io/milvus/client/v2. Migration to the new SDK is planned but will
// require breaking changes in a future version. See issue #1397 for migration tracking.
package milvus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/vectorstores"
)

// Store is a wrapper around the milvus client.
type Store struct {
	dropOld          bool
	async            bool
	loaded           bool
	collectionExists bool
	shardNum         int32
	maxTextLength    int
	ef               int
	collectionName   string
	partitionName    string
	textField        string
	metaField        string
	primaryField     string
	vectorField      string
	consistencyLevel entity.ConsistencyLevel
	index            entity.Index
	embedder         embeddings.Embedder
	client           client.Client
	metricType       entity.MetricType
	searchParameters entity.SearchParam
	schema           *entity.Schema
	skipFlushOnWrite bool
}

var (
	_ vectorstores.VectorStore = Store{}

	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	ErrColumnNotFound = errors.New("invalid field")
	ErrInvalidFilters = errors.New("invalid filters")
)

// New creates an active client connection to the (specified, or default) collection in the Milvus server
// and returns the `Store` object needed by the other accessors.
func New(ctx context.Context, config client.Config, opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return s, err
	}
	if s.client, err = client.NewClient(ctx, config); err != nil {
		return s, err
	}
	s.collectionExists, err = s.client.HasCollection(ctx, s.collectionName)
	if err != nil {
		return s, err
	}

	if s.collectionExists && s.dropOld {
		if err := s.dropCollection(ctx, s.collectionName); err != nil {
			return s, err
		}
		s.collectionExists = false
	}

	return s, s.init(ctx, 0)
}

func (s *Store) init(ctx context.Context, dim int) error {
	if s.loaded {
		return nil
	}
	if err := s.createCollection(ctx, dim); err != nil {
		return err
	}
	if err := s.extractFields(ctx); err != nil {
		return err
	}
	if err := s.createIndex(ctx); err != nil {
		return err
	}
	if err := s.createSearchParams(ctx); err != nil {
		return err
	}
	return s.load(ctx)
}

func (s *Store) dropCollection(ctx context.Context, name string) error {
	return s.client.DropCollection(ctx, name)
}

func (s *Store) extractFields(ctx context.Context) error {
	if !s.collectionExists || s.schema != nil {
		return nil
	}
	collection, err := s.client.DescribeCollection(ctx, s.collectionName)
	if err != nil {
		return err
	}
	s.schema = collection.Schema
	return nil
}

func (s *Store) createCollection(ctx context.Context, dim int) error {
	if dim == 0 || s.collectionExists {
		return nil
	}
	s.schema = &entity.Schema{
		CollectionName: s.collectionName,
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       s.primaryField,
				DataType:   entity.FieldTypeInt64,
				AutoID:     true,
				PrimaryKey: true,
			},
			{
				Name:     s.textField,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: strconv.Itoa(s.maxTextLength),
				},
			},
			{
				Name:     s.metaField,
				DataType: entity.FieldTypeJSON,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: strconv.Itoa(s.maxTextLength),
				},
			},
			{
				Name:     s.vectorField,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: strconv.Itoa(dim),
				},
			},
		},
	}

	err := s.client.CreateCollection(ctx, s.schema, s.shardNum, client.WithMetricsType(s.metricType))
	if err != nil {
		return err
	}
	s.collectionExists = true
	return nil
}

func (s *Store) createIndex(ctx context.Context) error {
	if !s.collectionExists {
		return nil
	}

	return s.client.CreateIndex(ctx, s.collectionName, s.vectorField, s.index, s.async)
}

func (s *Store) createSearchParams(ctx context.Context) error {
	if s.searchParameters != nil {
		return nil
	}
	return s.getIndex(ctx)
}

func (s *Store) getIndex(ctx context.Context) error {
	idx, err := s.client.DescribeIndex(ctx, s.collectionName, s.vectorField)
	if err != nil {
		return err
	}
	s.index = idx[0]
	return nil
}

func (s *Store) load(ctx context.Context) error {
	if s.loaded || !s.collectionExists {
		return nil
	}

	err := s.client.LoadCollection(ctx, s.collectionName, s.async)
	if err == nil {
		s.loaded = true
	}
	return err
}

// AddDocuments adds the text and metadata from the documents to the Milvus collection associated with 'Store'.
// and returns the ids of the added documents.
func (s Store) AddDocuments(ctx context.Context, docs []schema.Document,
	_ ...vectorstores.Option,
) ([]string, error) {
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}
	if err := s.init(ctx, len(vectors[0])); err != nil {
		return nil, err
	}

	colsData := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		docMap := map[string]any{
			s.metaField:   doc.Metadata,
			s.textField:   doc.PageContent,
			s.vectorField: vectors[i],
		}
		colsData = append(colsData, docMap)
	}

	_, err = s.client.InsertRows(ctx, s.collectionName, s.partitionName, colsData)
	if err != nil {
		return nil, err
	}
	if !s.skipFlushOnWrite {
		if err = s.client.Flush(ctx, s.collectionName, false); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (s *Store) getSearchFields() []string {
	fields := []string{}
	for _, f := range s.schema.Fields {
		if f.DataType == entity.FieldTypeBinaryVector || f.DataType == entity.FieldTypeFloatVector {
			continue
		}
		fields = append(fields, f.Name)
	}
	return fields
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) convertResultToDocument(searchResult []client.SearchResult) ([]schema.Document, error) {
	docs := []schema.Document{}
	var err error

	for _, res := range searchResult {
		if res.ResultCount == 0 {
			continue
		}
		textcol, ok := res.Fields.GetColumn(s.textField).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("%w: text column missing", ErrColumnNotFound)
		}
		metacol, ok := res.Fields.GetColumn(s.metaField).(*entity.ColumnJSONBytes)
		if !ok {
			return nil, fmt.Errorf("%w: metadata column missing", ErrColumnNotFound)
		}
		for i := 0; i < res.ResultCount; i++ {
			doc := schema.Document{}

			doc.PageContent, err = textcol.ValueByIdx(i)
			if err != nil {
				return nil, err
			}
			metaStr, err := metacol.ValueByIdx(i)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(metaStr, &doc.Metadata); err != nil {
				return nil, err
			}
			doc.Score = res.Scores[i]
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	filter, err := s.getFilters(opts)
	if err != nil {
		return nil, err
	}
	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	if err := s.init(ctx, len(vector)); err != nil {
		return nil, err
	}
	vectors := []entity.Vector{
		entity.FloatVector(vector),
	}
	partitions := []string{}
	if s.partitionName != "" {
		partitions = append(partitions, s.partitionName)
	}
	sp := s.searchParameters
	if opts.ScoreThreshold > 0 {
		sp.AddRadius(float64(opts.ScoreThreshold))
	}
	searchResult, err := s.client.Search(ctx,
		s.collectionName,
		partitions,
		filter,
		s.getSearchFields(),
		vectors,
		s.vectorField,
		s.metricType,
		numDocuments,
		sp,
		client.WithSearchQueryConsistencyLevel(s.consistencyLevel),
	)
	if err != nil {
		return nil, err
	}

	return s.convertResultToDocument(searchResult)
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
