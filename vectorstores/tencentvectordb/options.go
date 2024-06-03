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
	_tencentvectordbEnvVrName = "TENCENTVECTORDB_API_KEY"
	_defaultUserName          = "root"
	_defaultDatabase          = "LangChainDatabase"
	_defaultCollection        = "LangChainCollection"
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

// WithDatabase is an option for setting the database to use. Default is "LangChainDatabase".
func WithDatabase(database string) Option {
	return func(p *Store) {
		p.database = database
	}
}

// WithCollectionName is an option for setting the collection to use. Default is "LangChainCollection".
func WithCollectionName(collectionName string) Option {
	return func(p *Store) {
		p.collectionName = collectionName
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		userName:       _defaultUserName,
		database:       _defaultDatabase,
		collectionName: _defaultCollection,
		shardNum:       1,
		replicasNum:    2,
		indexType:      "HNSW",
		metricType:     "L2",
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.url == "" {
		return Store{}, fmt.Errorf("%w: missing url", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	if o.apiKey == "" {
		o.apiKey = os.Getenv(_tencentvectordbEnvVrName)
		if o.apiKey == "" {
			return Store{}, fmt.Errorf(
				"%w: missing api key. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions,
				_tencentvectordbEnvVrName,
			)
		}
	}

	return *o, nil
}
