package milvus

import (
	"errors"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/starmvp/langchaingo/embeddings"
)

const (
	_defaultCollectionName   = "LangChainGoCollection"
	_defaultConsistencyLevel = entity.ClSession
	_defaultPrimaryField     = "pk"
	_defaultTextField        = "text"
	_defaultMetaField        = "meta"
	_defaultVectorField      = "vector"
	_defaultMaxLength        = 65535
	_defaultEF               = 10
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithPartitionName sets the milvus partition for the collection.
func WithPartitionName(name string) Option {
	return func(s *Store) {
		s.partitionName = name
	}
}

// WithCollectionName sets the collection for the milvus store.
func WithCollectionName(name string) Option {
	return func(s *Store) {
		s.collectionName = name
	}
}

// WithEmbedder sets the embedder to use.
func WithEmbedder(embedder embeddings.Embedder) Option {
	return func(s *Store) {
		s.embedder = embedder
	}
}

// WithTextField sets the name of the text field in the collection schema.
func WithTextField(str string) Option {
	return func(s *Store) {
		s.textField = str
	}
}

// WithMetaField sets te name of the meta field in the collection schema.
// default is 'meta'.
func WithMetaField(str string) Option {
	return func(s *Store) {
		s.metaField = str
	}
}

// WithMaxTextLength sets the maximum length of the text field in the collection.
func WithMaxTextLength(num int) Option {
	return func(s *Store) {
		s.maxTextLength = num
	}
}

// WithPrimaryField name of the primary field in the collection.
func WithPrimaryField(str string) Option {
	return func(s *Store) {
		s.primaryField = str
	}
}

// WithVectorField sets the name of the vector field in the collection.
func WithVectorField(str string) Option {
	return func(s *Store) {
		s.vectorField = str
	}
}

// WithConsistencyLevel sets the consistency level for seaarch.
func WithConsistencyLevel(level entity.ConsistencyLevel) Option {
	return func(s *Store) {
		s.consistencyLevel = level
	}
}

// WithDropOld store will drop and recreate collection on initialization.
func WithDropOld() Option {
	return func(s *Store) {
		s.dropOld = true
	}
}

// WithIndex for vector search.
func WithIndex(idx entity.Index) Option {
	return func(s *Store) {
		s.index = idx
	}
}

// WithShards number of shards to create for a collection.
func WithShards(num int32) Option {
	return func(s *Store) {
		s.shardNum = num
	}
}

// WithEF sets ef.
func WithEF(ef int) Option {
	return func(s *Store) {
		s.ef = ef
	}
}

// WithSearchParameters sets the search parameters.
func WithSearchParameters(sp entity.SearchParam) Option {
	return func(s *Store) {
		s.searchParameters = sp
	}
}

// WithMetricType sets the metric type for the collection.
func WithMetricType(metricType entity.MetricType) Option {
	return func(p *Store) {
		p.metricType = metricType
	}
}

// WithSkipFlushOnWrite disables flushing on write.
func WithSkipFlushOnWrite() Option {
	return func(s *Store) {
		s.skipFlushOnWrite = true
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	s := Store{
		metricType:       entity.L2,
		primaryField:     _defaultPrimaryField,
		vectorField:      _defaultVectorField,
		maxTextLength:    _defaultMaxLength,
		textField:        _defaultTextField,
		metaField:        _defaultMetaField,
		consistencyLevel: _defaultConsistencyLevel,
		collectionName:   _defaultCollectionName,
		ef:               _defaultEF,
		shardNum:         entity.DefaultShardNumber,
	}

	for _, opt := range opts {
		opt(&s)
	}

	if s.embedder == nil {
		return s, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	if s.index == nil {
		return s, fmt.Errorf("%w: missing index function", ErrInvalidOptions)
	}
	if s.searchParameters == nil {
		idx, err := entity.NewIndexHNSWSearchParam(s.ef)
		if err != nil {
			return s, err
		}
		s.searchParameters = idx
	}
	return s, nil
}
