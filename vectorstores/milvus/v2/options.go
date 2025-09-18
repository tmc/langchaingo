package v2

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	oldentity "github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
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

// ConfigAdapter handles conversion between v1 and v2 configurations
type ConfigAdapter struct{}

// ToV2Config converts various config types to milvusclient.ClientConfig
func (ca ConfigAdapter) ToV2Config(config interface{}) (milvusclient.ClientConfig, error) {
	switch cfg := config.(type) {
	case milvusclient.ClientConfig:
		// Already v2 config, return as-is
		return cfg, nil
	case client.Config:
		// Convert v1 config to v2 config
		return milvusclient.ClientConfig{
			Address: cfg.Address,
			// Add other field mappings as needed
		}, nil
	case string:
		// Simple address string
		return milvusclient.ClientConfig{
			Address: cfg,
		}, nil
	default:
		return milvusclient.ClientConfig{}, fmt.Errorf("unsupported config type: %T", config)
	}
}

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

// WithMetaField sets the name of the meta field in the collection schema.
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

// WithPrimaryField sets the name of the primary field in the collection schema.
func WithPrimaryField(str string) Option {
	return func(s *Store) {
		s.primaryField = str
	}
}

// WithVectorField sets the name of the vector field in the collection schema.
func WithVectorField(str string) Option {
	return func(s *Store) {
		s.vectorField = str
	}
}

// WithConsistencyLevel sets the consistency level for the collection.
func WithConsistencyLevel(level entity.ConsistencyLevel) Option {
	return func(s *Store) {
		s.consistencyLevel = level
	}
}

// WithConsistencyLevelV1 sets the consistency level from v1 type (compatibility).
func WithConsistencyLevelV1(level oldentity.ConsistencyLevel) Option {
	return func(s *Store) {
		// Convert v1 consistency level to v2
		s.consistencyLevel = entity.ConsistencyLevel(level)
	}
}

// WithDropOld sets the drop old collection flag.
func WithDropOld() Option {
	return func(s *Store) {
		s.dropOld = true
	}
}

// WithIndex sets the index to use for the vector field.
func WithIndex(idx index.Index) Option {
	return func(s *Store) {
		s.index = idx
	}
}

// WithIndexV1 sets the index from v1 type (compatibility).
func WithIndexV1(idx oldentity.Index) Option {
	return func(s *Store) {
		// Convert v1 index to v2 index
		s.index = convertV1IndexToV2(idx)
	}
}

// WithShards sets the number of shards for the collection.
func WithShards(num int32) Option {
	return func(s *Store) {
		s.shardNum = num
	}
}

// WithEF sets the ef parameter for HNSW index.
func WithEF(ef int) Option {
	return func(s *Store) {
		s.ef = ef
	}
}

// WithSearchParameters sets the search parameters.
func WithSearchParameters(sp map[string]interface{}) Option {
	return func(s *Store) {
		s.searchParameters = sp
	}
}

// WithSearchParametersV1 sets search parameters from v1 type (compatibility).
func WithSearchParametersV1(sp oldentity.SearchParam) Option {
	return func(s *Store) {
		// Convert v1 search params to v2 map
		s.searchParameters = convertV1SearchParamToV2(sp)
	}
}

// WithMetricType sets the metric type for the vector field.
func WithMetricType(metricType entity.MetricType) Option {
	return func(s *Store) {
		s.metricType = metricType
	}
}

// WithMetricTypeV1 sets metric type from v1 type (compatibility).
func WithMetricTypeV1(metricType oldentity.MetricType) Option {
	return func(s *Store) {
		// Convert v1 metric type to v2
		s.metricType = entity.MetricType(metricType)
	}
}

// WithSkipFlushOnWrite sets the skip flush on write flag.
func WithSkipFlushOnWrite() Option {
	return func(s *Store) {
		s.skipFlushOnWrite = true
	}
}

// Helper functions for v1 to v2 conversion

func convertV1IndexToV2(v1Index oldentity.Index) index.Index {
	// Simplified conversion that works with any v1 index type
	// Default to L2 metric since we can't easily extract it from v1 index interface
	defaultMetric := entity.L2

	indexType := v1Index.IndexType()
	params := v1Index.Params()

	switch indexType {
	case oldentity.AUTOINDEX:
		return index.NewAutoIndex(defaultMetric)
	case oldentity.Flat:
		return index.NewFlatIndex(defaultMetric)
	case oldentity.IvfFlat:
		// Extract nlist parameter if available
		nlist := 1024 // default
		if nlistVal, ok := params["nlist"]; ok {
			if nlistInt, err := strconv.Atoi(nlistVal); err == nil {
				nlist = nlistInt
			}
		}
		return index.NewIvfFlatIndex(defaultMetric, nlist)
	case oldentity.HNSW:
		// Extract M and efConstruction parameters
		m := 16               // default
		efConstruction := 200 // default
		if mVal, ok := params["M"]; ok {
			if mInt, err := strconv.Atoi(mVal); err == nil {
				m = mInt
			}
		}
		if efVal, ok := params["efConstruction"]; ok {
			if efInt, err := strconv.Atoi(efVal); err == nil {
				efConstruction = efInt
			}
		}
		return index.NewHNSWIndex(defaultMetric, m, efConstruction)
	default:
		// Fallback to auto index
		return index.NewAutoIndex(defaultMetric)
	}
}

func convertV1SearchParamToV2(v1Param oldentity.SearchParam) map[string]interface{} {
	// Convert v1 search parameters to v2 map format
	params := make(map[string]interface{})

	// Copy parameters from v1 to v2 format
	for key, value := range v1Param.Params() {
		params[key] = value
	}

	return params
}

func applyClientOptions(opts ...Option) (Store, error) {
	s := Store{
		collectionName:   _defaultCollectionName,
		consistencyLevel: _defaultConsistencyLevel,
		primaryField:     _defaultPrimaryField,
		textField:        _defaultTextField,
		metaField:        _defaultMetaField,
		vectorField:      _defaultVectorField,
		maxTextLength:    _defaultMaxLength,
		ef:               _defaultEF,
		shardNum:         1,
		searchParameters: make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(&s)
	}

	if s.embedder == nil {
		return s, fmt.Errorf("%w: embedder is required", ErrInvalidOptions)
	}

	// Set default index if not provided
	if s.index == nil {
		s.index = index.NewAutoIndex(s.metricType)
	}

	return s, nil
}
