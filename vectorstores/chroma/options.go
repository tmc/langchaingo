package chroma

import (
	"errors"
	"fmt"
	"os"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/tmc/langchaingo/embeddings"
)

const (
	OpenAiAPIKeyEnvVarName = "OPENAI_API_KEY" // #nosec G101
	ChromaURLKeyEnvVarName = "CHROMA_URL"
	DefaultNameSpace       = "langchain"
	DefaultNameSpaceKey    = "nameSpace"
	DefaultCollectionName  = "langchaingo"
	DefaultDistanceFunc    = chromago.L2
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithCollectionName is an option for specifying the name of the Chroma collection to use. Must be set.
func WithCollectionName(name string) Option {
	return func(p *Store) {
		p.collectionName = name
	}
}

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

// WithDistanceFunction specifies the distance function which will be used
// see: https://github.com/amikos-tech/chroma-go/blob/d0087270239eccdb2f4f03d84b18d875c601ad6b/main.go#L96
func WithDistanceFunction(distanceFunction chromago.DistanceFunction) Option {
	return func(p *Store) {
		p.distanceFunction = distanceFunction
	}
}

// WithIncludes is an option for setting the includes to query the vectors.
func WithIncludes(includes []chromago.QueryEnum) Option {
	return func(p *Store) {
		p.includes = includes
	}
}

// WithOpenAiAPIKey is an option for setting the OpenAI api key. If the option is not set
// the api key is read from the OPENAI_API_KEY environment variable. If the
// variable is not present, an error will be returned.
func WithOpenAiAPIKey(openAiAPIKey string) Option {
	return func(p *Store) {
		p.openaiAPIKey = openAiAPIKey
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		nameSpace:        DefaultNameSpace,
		nameSpaceKey:     DefaultNameSpaceKey,
		collectionName:   DefaultCollectionName,
		distanceFunction: DefaultDistanceFunc,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.collectionName == "" {
		return Store{}, fmt.Errorf("%w: missing collection name", ErrInvalidOptions)
	}

	if o.chromaURL == "" {
		o.chromaURL = os.Getenv(ChromaURLKeyEnvVarName)
		if o.chromaURL == "" {
			return Store{}, fmt.Errorf(
				"%w: missing chroma URL. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions, ChromaURLKeyEnvVarName)
		}
	}

	if o.openaiAPIKey == "" {
		o.openaiAPIKey = os.Getenv(OpenAiAPIKeyEnvVarName)
		if o.openaiAPIKey == "" {
			return Store{}, fmt.Errorf(
				"%w: missing OpenAiAPI key. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions, OpenAiAPIKeyEnvVarName)
		}
	}

	return *o, nil
}
