package milvus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// Store is a wrapper around the milvus client.
type Store struct {
	dropOld          bool
	async            bool
	initialized      bool
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
}

var (
	_ vectorstores.VectorStore = Store{}

	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	ErrColumnNotFound = errors.New("invalid field")
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
	ok, err := s.client.HasCollection(ctx, s.collectionName)
	if err != nil {
		return s, err
	}

	if ok && s.dropOld {
		if err := s.dropCollection(ctx, s.collectionName); err != nil {
			return s, err
		}
	}

	return s, s.init(ctx, 0)
}

func (s *Store) init(ctx context.Context, dim int) error {
	if s.initialized {
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
	if !s.initialized || s.schema.Fields != nil {
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
	if dim < 1 {
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
				DataType: entity.FieldTypeVarChar,
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
	err := s.client.CreateCollection(ctx, s.schema, s.shardNum)
	if err != nil {
		return err
	}
	s.initialized = true
	return nil
}

func (s *Store) createIndex(ctx context.Context) error {
	if !s.initialized {
		return nil
	}

	return s.client.CreateIndex(ctx, s.collectionName, s.vectorField, s.index, s.async)
}

func (s *Store) createSearchParams(ctx context.Context) error {
	if !s.initialized || s.searchParameters != nil {
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
	if !s.initialized {
		return nil
	}
	return s.client.LoadCollection(ctx, s.collectionName, s.async)
}

// AddDocuments adds the text and metadata from the documents to the Milvus collection associated with 'Store'.
func (s Store) AddDocuments(ctx context.Context, docs []schema.Document,
	_ ...vectorstores.Option,
) error {
	texts := make([]string, 0, len(docs))
	metadatas := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
		buf, err := json.Marshal(doc.Metadata)
		if err != nil {
			return err
		}
		metadatas = append(metadatas, string(buf))
	}
	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	if len(vectors) != len(docs) {
		return ErrEmbedderWrongNumberVectors
	}
	if err := s.init(ctx, len(vectors[0])); err != nil {
		return err
	}

	textCol := entity.NewColumnVarChar(s.textField, texts)
	metaCol := entity.NewColumnVarChar(s.metaField, metadatas)
	vectorCol := entity.NewColumnFloatVector(s.vectorField, len(vectors[0]), vectors)
	_, err = s.client.Insert(ctx, s.collectionName, s.partitionName, vectorCol, metaCol, textCol)
	if err != nil {
		return err
	}

	return nil
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
		textcol, ok := res.Fields.GetColumn(s.textField).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("%w: text column missing", ErrColumnNotFound)
		}
		metacol, ok := res.Fields.GetColumn(s.metaField).(*entity.ColumnVarChar)
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

			if err := json.Unmarshal([]byte(metaStr), &doc.Metadata); err != nil {
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

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	if err := s.init(ctx, len(vector)); err != nil {
		return nil, err
	}
	fv := entity.FloatVector(vector)
	vectors := []entity.Vector{fv}
	partitions := []string{}
	if s.partitionName != "" {
		partitions = append(partitions, s.partitionName)
	}
	sp := s.searchParameters
	if opts.ScoreThreshold > 0 {
		sp.AddRadius(float64(opts.ScoreThreshold))
	}

	searchResult, err := s.client.Search(ctx, s.collectionName,
		partitions,
		"",
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
