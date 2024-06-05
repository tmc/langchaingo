package tencentvectordb

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tencent/vectordatabase-sdk-go/tcvectordb"
	"github.com/tmc/langchaingo/embeddings"
)

const (
	_tencentvectordbEnvVrApiKey = "TENCENTVECTORDB_API_KEY"
	_tencentvectordbEnvVrUrl    = "TENCENTVECTORDB_URL"
	_defaultUserName            = "root"
	_defaultDatabase            = "LangChainDatabase"
	_defaultCollection          = "LangChainCollection"
	_defaultEmbeddingModel      = "BGE_BASE_ZH"
	_defaultIndexType           = "HNSW"
	_defaultMetricType          = "COSINE"
	_defaultDimension           = 1536 // 1536 is the default dimension for OpenAI embedding
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithHost is an option for setting the url to use. Must be set.
func WithUrl(url string) Option {
	return func(p *Store) {
		p.url = strings.TrimSpace(url)
	}
}

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithAPIKey is an option for setting the api key. If the option is not set
// the api key is read from the TENCENTVECTORDB_API_KEY environment variable. If the
// variable is not present, an error will be returned.
func WithAPIKey(apiKey string) Option {
	return func(p *Store) {
		p.apiKey = apiKey
	}
}

// WithUserName is an option for setting the user name to use. Default is "root".
func WithUserName(userName string) Option {
	return func(p *Store) {
		p.userName = userName
	}
}

// WithUserOption is an option for setting the user option to use.
func WithUserOption(userOption *tcvectordb.ClientOption) Option {
	return func(p *Store) {
		p.userOption = userOption
	}
}

// WithMetaField is an option for setting the meta fields to use.
func WithMetaField(metaFields []MetaField) Option {
	return func(p *Store) {
		p.metaFields = metaFields
	}
}

// WithDatabase is an option for setting the database to use. Default is "LangChainDatabase".
func WithDatabase(database string) Option {
	return func(p *Store) {
		p.databaseName = database
	}
}

// WithCollectionName is an option for setting the collection to use. Default is "LangChainCollection".
func WithCollectionName(collectionName string) Option {
	return func(p *Store) {
		p.collectionName = collectionName
	}
}

func WithEmbeddingModel(embeddingModel string) Option {
	return func(p *Store) {
		p.embeddingModel = embeddingModel
	}
}

func WithIndexType(indexType string) Option {
	return func(p *Store) {
		p.indexType = indexType
	}
}

func WithMetricType(metricType string) Option {
	return func(p *Store) {
		p.metricType = metricType
	}
}
func WithDimension(dimension uint32) Option {
	return func(p *Store) {
		p.dimension = dimension
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		userName:       _defaultUserName,
		databaseName:   _defaultDatabase,
		collectionName: _defaultCollection,
		shardNum:       1,
		replicasNum:    0,
		indexType:      _defaultIndexType,
		metricType:     _defaultMetricType,
		embeddingModel: _defaultEmbeddingModel,
		dimension:      _defaultDimension,
	}

	for _, opt := range opts {
		opt(o)
	}
	if o.apiKey == "" {
		o.apiKey = os.Getenv(_tencentvectordbEnvVrApiKey)
		if o.apiKey == "" {
			return Store{}, fmt.Errorf(
				"%w: missing api key. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions,
				_tencentvectordbEnvVrApiKey,
			)
		}
	}
	if o.url == "" {
		o.url = os.Getenv(_tencentvectordbEnvVrUrl)
		if o.url == "" {
			return Store{}, fmt.Errorf(
				"%w: missing URL. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions,
				_tencentvectordbEnvVrUrl,
			)
		}
	}

	return *o, nil
}
