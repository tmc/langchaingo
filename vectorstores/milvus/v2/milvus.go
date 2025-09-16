// Package v2 provides a vectorstore implementation for Milvus using the new SDK.
// This package uses github.com/milvus-io/milvus/client/v2 which is actively maintained
// and replaces the archived milvus-sdk-go/v2 package.
package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// Store is a wrapper around the milvus client using the new SDK.
type Store struct {
	dropOld          bool
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
	index            index.Index
	embedder         embeddings.Embedder
	client           *milvusclient.Client
	metricType       entity.MetricType
	searchParameters map[string]interface{} // Using map for search parameters in v2
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
// Supports both v1 (client.Config) and v2 (milvusclient.ClientConfig) configurations for compatibility.
func New(ctx context.Context, config interface{}, opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return s, err
	}

	// Convert config to v2 format using adapter
	adapter := ConfigAdapter{}
	v2Config, err := adapter.ToV2Config(config)
	if err != nil {
		return s, fmt.Errorf("failed to convert config: %w", err)
	}

	if s.client, err = milvusclient.New(ctx, &v2Config); err != nil {
		return s, err
	}

	hasCollOpt := milvusclient.NewHasCollectionOption(s.collectionName)
	s.collectionExists, err = s.client.HasCollection(ctx, hasCollOpt)
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
	dropOpt := milvusclient.NewDropCollectionOption(name)
	return s.client.DropCollection(ctx, dropOpt)
}

func (s *Store) extractFields(ctx context.Context) error {
	if !s.collectionExists || s.schema != nil {
		return nil
	}
	descOpt := milvusclient.NewDescribeCollectionOption(s.collectionName)
	collection, err := s.client.DescribeCollection(ctx, descOpt)
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

	createOpt := milvusclient.NewCreateCollectionOption(s.collectionName, s.schema)
	createOpt.WithShardNum(s.shardNum)
	err := s.client.CreateCollection(ctx, createOpt)
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

	createIndexOpt := milvusclient.NewCreateIndexOption(s.collectionName, s.vectorField, s.index)
	_, err := s.client.CreateIndex(ctx, createIndexOpt)
	return err
}

func (s *Store) createSearchParams(ctx context.Context) error {
	if s.searchParameters != nil {
		return nil
	}
	return s.getIndex(ctx)
}

func (s *Store) getIndex(ctx context.Context) error {
	listIndexOpt := milvusclient.NewListIndexOption(s.collectionName)
	indexes, err := s.client.ListIndexes(ctx, listIndexOpt)
	if err != nil {
		return err
	}
	if len(indexes) > 0 {
		descIndexOpt := milvusclient.NewDescribeIndexOption(s.collectionName, indexes[0])
		_, err := s.client.DescribeIndex(ctx, descIndexOpt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) load(ctx context.Context) error {
	if s.loaded || !s.collectionExists {
		return nil
	}

	loadOpt := milvusclient.NewLoadCollectionOption(s.collectionName)
	_, err := s.client.LoadCollection(ctx, loadOpt)
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

	rows := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		docMap := map[string]any{
			s.metaField:   doc.Metadata,
			s.textField:   doc.PageContent,
			s.vectorField: vectors[i],
		}
		rows = append(rows, docMap)
	}

	insertOpt := milvusclient.NewRowBasedInsertOption(s.collectionName, rows...)
	if s.partitionName != "" {
		insertOpt.WithPartition(s.partitionName)
	}

	_, err = s.client.Insert(ctx, insertOpt)
	if err != nil {
		return nil, err
	}

	if !s.skipFlushOnWrite {
		flushOpt := milvusclient.NewFlushOption(s.collectionName)
		_, err = s.client.Flush(ctx, flushOpt)
		if err != nil {
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

func (s Store) convertResultToDocument(searchResults []milvusclient.ResultSet) ([]schema.Document, error) {
	docs := []schema.Document{}

	for _, resultSet := range searchResults {
		for i := 0; i < resultSet.ResultCount; i++ {
			doc := schema.Document{}

			// Get text content
			textCol := resultSet.GetColumn(s.textField)
			if textCol != nil {
				textVal, err := textCol.Get(i)
				if err == nil {
					if textStr, ok := textVal.(string); ok {
						doc.PageContent = textStr
					}
				}
			}

			// Get metadata
			metaCol := resultSet.GetColumn(s.metaField)
			if metaCol != nil {
				metaVal, err := metaCol.Get(i)
				if err == nil {
					if metaBytes, ok := metaVal.([]byte); ok {
						if err := json.Unmarshal(metaBytes, &doc.Metadata); err != nil {
							return nil, err
						}
					}
				}
			}

			// Get score
			if i < len(resultSet.Scores) {
				doc.Score = resultSet.Scores[i]
			}

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

	searchOpt := milvusclient.NewSearchOption(s.collectionName, numDocuments, vectors)
	searchOpt.WithFilter(filter)
	searchOpt.WithOutputFields(s.getSearchFields()...)

	if s.partitionName != "" {
		searchOpt.WithPartitions(s.partitionName)
	}

	if opts.ScoreThreshold > 0 {
		searchOpt.WithSearchParam("radius", fmt.Sprintf("%f", opts.ScoreThreshold))
	}

	searchResult, err := s.client.Search(ctx, searchOpt)
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
