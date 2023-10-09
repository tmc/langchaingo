package pinecone

import (
	"errors"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/embeddings"
)

const (
	_pineconeEnvVrName = "PINECONE_API_KEY"
	_defaultTextKey    = "text"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithIndexName is an option for specifying the index name. Must be set.
func WithIndexName(name string) Option {
	return func(p *Store) {
		p.indexName = name
	}
}

// WithEnvironment is an option for specifying the environment. Must be set.
func WithEnvironment(environment string) Option {
	return func(p *Store) {
		p.environment = environment
	}
}

// WithProjectName is an option for specifying the project name. Must be set. The
// project name associated with the api key can be obtained using the whoami
// operation.
func WithProjectName(name string) Option {
	return func(p *Store) {
		p.projectName = name
	}
}

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithAPIKey is an option for setting the api key. If the option is not set
// the api key is read from the PINECONE_API_KEY environment variable. If the
// variable is not present, an error will be returned.
func WithAPIKey(apiKey string) Option {
	return func(p *Store) {
		p.apiKey = apiKey
	}
}

// WithTextKey is an option for setting the text key in the metadata to the vectors
// in the index. The text key stores the text of the document the vector represents.
func WithTextKey(textKey string) Option {
	return func(p *Store) {
		p.textKey = textKey
	}
}

// NameSpace is an option for setting the nameSpace to upsert and query the vectors
// from. Must be set.
func WithNameSpace(nameSpace string) Option {
	return func(p *Store) {
		p.nameSpace = nameSpace
	}
}

// withGrpc is an option for using the grpc api instead of the rest api.
func withGrpc() Option { // nolint: unused
	return func(p *Store) {
		p.useGRPC = true
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		textKey: _defaultTextKey,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.indexName == "" {
		return Store{}, fmt.Errorf("%w: missing index name", ErrInvalidOptions)
	}

	if o.environment == "" {
		return Store{}, fmt.Errorf("%w: missing environment", ErrInvalidOptions)
	}

	if o.projectName == "" {
		return Store{}, fmt.Errorf("%w: missing project name", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	if o.apiKey == "" {
		o.apiKey = os.Getenv(_pineconeEnvVrName)
		if o.apiKey == "" {
			return Store{}, fmt.Errorf(
				"%w: missing api key. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions,
				_pineconeEnvVrName,
			)
		}
	}

	return *o, nil
}
