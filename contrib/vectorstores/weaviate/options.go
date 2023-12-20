package weaviate

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"golang.org/x/exp/slices"
)

const (
	_defaultNameSpaceKey = "nameSpace"
	_defaultTextKey      = "text"
	_defaultNameSpace    = "default"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithTextKey is an option for setting the text key in the metadata to the vectors
// in the index. The text key stores the text of the document the vector represents.
func WithTextKey(textKey string) Option {
	return func(p *Store) {
		p.textKey = textKey
	}
}

// WithNameSpaceKey is an option for setting the nameSpace key in the metadata to the vectors
// in the index. The nameSpace key stores the nameSpace of the document the vector represents.
func WithNameSpaceKey(nameSpaceKey string) Option {
	return func(p *Store) {
		p.nameSpaceKey = nameSpaceKey
	}
}

// WithIndexName is an option for specifying the index name. Must be set.
// The index name is the name of the class in weaviate.
// Multiple words should be concatenated in CamelCase, e.g. ArticleAuthor.
// https://weaviate.io/developers/weaviate/api/rest/schema#create-a-class
func WithIndexName(indexName string) Option {
	return func(p *Store) {
		p.indexName = indexName
	}
}

// WithNameSpace is an option for setting the nameSpace to upsert and query the vectors.
func WithNameSpace(nameSpace string) Option {
	return func(p *Store) {
		p.nameSpace = nameSpace
	}
}

// WithHost is an option for setting the host of the weaviate server.
func WithHost(host string) Option {
	return func(p *Store) {
		p.host = host
	}
}

// WithScheme is an option for setting the scheme of the weaviate server.
func WithScheme(scheme string) Option {
	return func(p *Store) {
		p.scheme = scheme
	}
}

// WithAPIKey is an option for setting the api key. If the option is not set
// the api key is read from the WEAVIATE_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(p *Store) {
		p.apiKey = &apiKey
	}
}

// WithAuthConfig is an option for setting the auth config of the weaviate server.
func WithAuthConfig(authConfig auth.Config) Option {
	return func(p *Store) {
		p.authConfig = authConfig
	}
}

// WithConnectionClient is an option for setting the connection client of the weaviate server.
func WithConnectionClient(connectionClient *http.Client) Option {
	return func(p *Store) {
		p.connectionClient = connectionClient
	}
}

// WithQueryAttrs is an option for setting the query attributes of the weaviate server.
func WithQueryAttrs(queryAttrs []string) Option {
	return func(p *Store) {
		p.queryAttrs = queryAttrs
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		textKey:      _defaultTextKey,
		nameSpaceKey: _defaultNameSpaceKey,
		nameSpace:    _defaultNameSpace,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.indexName == "" {
		return Store{}, fmt.Errorf("%w: missing indexName", ErrInvalidOptions)
	}

	if o.scheme == "" {
		return Store{}, fmt.Errorf("%w: missing scheme", ErrInvalidOptions)
	}

	if o.host == "" {
		return Store{}, fmt.Errorf("%w: missing host", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	// add default Attributes
	if o.queryAttrs == nil {
		o.queryAttrs = []string{o.textKey, o.nameSpaceKey}
	}
	if !slices.Contains(o.queryAttrs, o.textKey) {
		o.queryAttrs = append(o.queryAttrs, o.textKey)
	}
	if !slices.Contains(o.queryAttrs, o.nameSpaceKey) {
		o.queryAttrs = append(o.queryAttrs, o.nameSpaceKey)
	}

	return *o, nil
}
