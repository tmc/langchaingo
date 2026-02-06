package chroma

import (
	"errors"
	"fmt"
	"os"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaembed "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/tmc/langchaingo/embeddings"
)

const (
	ChromaURLKeyEnvVarName = "CHROMA_URL"
	ChromaAPIKeyEnvVarName = "CHROMA_API_KEY" //nolint:gosec // env var name, not a credential
	DefaultNameSpace       = "langchain"
	DefaultNameSpaceKey    = "nameSpace"
	DefaultDistanceFunc    = chromaembed.L2
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithNameSpace sets the nameSpace used to upsert and query the vectors from.
func WithNameSpace(nameSpace string) Option {
	return func(p *Store) {
		p.nameSpace = nameSpace
	}
}

// WithChromaURL is an option for specifying the Chroma URL. Must be set.
func WithChromaURL(chromaURL string) Option {
	return func(p *Store) {
		p.chromaURL = chromaURL
	}
}

// WithEmbedder is an option for setting the embedder to use.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithDistanceFunction specifies the distance function which will be used (default is L2).
func WithDistanceFunction(distanceFunction chromaembed.DistanceMetric) Option {
	return func(p *Store) {
		p.distanceFunction = distanceFunction
	}
}

// WithIncludes is an option for setting the includes to query the vectors.
func WithIncludes(includes []chroma.Include) Option {
	return func(p *Store) {
		p.includes = includes
	}
}

// WithClient injects a pre-configured Chroma client, skipping URL/cloud validation.
func WithClient(client chroma.Client) Option {
	return func(p *Store) {
		p.client = client
	}
}

// WithCloudAPIKey sets the API key for Chroma Cloud.
// See https://docs.trychroma.com/cloud/getting-started for details.
func WithCloudAPIKey(apiKey string) Option {
	return func(p *Store) {
		p.cloudAPIKey = apiKey
	}
}

// WithTenant sets the tenant for Chroma Cloud.
func WithTenant(tenant string) Option {
	return func(p *Store) {
		p.tenant = tenant
	}
}

// WithDatabase sets the database for Chroma Cloud.
func WithDatabase(database string) Option {
	return func(p *Store) {
		p.database = database
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		nameSpace:        DefaultNameSpace,
		nameSpaceKey:     DefaultNameSpaceKey,
		distanceFunction: DefaultDistanceFunc,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	// If a pre-configured client was provided, skip URL/cloud validation.
	if o.client != nil {
		return *o, nil
	}

	// Cloud mode: API key provided via option or env var.
	if o.cloudAPIKey != "" || os.Getenv(ChromaAPIKeyEnvVarName) != "" {
		return *o, nil
	}

	// HTTP mode: require chromaURL.
	if o.chromaURL == "" {
		o.chromaURL = os.Getenv(ChromaURLKeyEnvVarName)
		if o.chromaURL == "" {
			return Store{}, fmt.Errorf(
				"%w: missing chroma URL. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions, ChromaURLKeyEnvVarName)
		}
	}

	return *o, nil
}
