package pinecone

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/starmvp/langchaingo/embeddings"
)

const (
	_pineconeEnvVrName = "PINECONE_API_KEY"
	_defaultTextKey    = "text"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithHost is an option for setting the host to use. Must be set.
func WithHost(host string) Option {
	return func(p *Store) {
		p.host = strings.TrimPrefix(host, "https://")
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

// WithNameSpace is an option for setting the nameSpace to upsert and query the vectors
// from. Must be set.
func WithNameSpace(nameSpace string) Option {
	return func(p *Store) {
		p.nameSpace = nameSpace
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		textKey: _defaultTextKey,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.host == "" {
		return Store{}, fmt.Errorf("%w: missing host", ErrInvalidOptions)
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
