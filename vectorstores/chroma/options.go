package chroma

import (
	"errors"
	"fmt"
	"os"

	chromago "github.com/amikos-tech/chroma-go"
)

const (
	_openAiAPIKeyEnvVrName = "OPENAI_API_KEY" // #nosec G101
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

// TODO (noodnik2): implement "NameSpace"
//  // NameSpace is an option for setting the nameSpace to upsert and query the vectors
//  // from. Must be set.
//  func WithNameSpace(nameSpace string) Option {
//  	return func(p *Store) {
//  		p.nameSpace = nameSpace
//  	}
//  }

// WithChromaURL is an option for specifying the Chroma URL. Must be set.
func WithChromaURL(chromaURL string) Option {
	return func(p *Store) {
		p.chromaURL = chromaURL
	}
}

// TODO (noodnik2): clarify need and implement if so
//  // WithProjectName is an option for specifying the project name. Must be set. The
//  // project name associated with the api key can be obtained using the whoami
//  // operation.
//  func WithProjectName(name string) Option {
//  	return func(p *Store) {
//  		p.projectName = name
//  	}
//  }
//

// TODO (noodnik2): implement
//  // WithEmbedder is an option for setting the embedder to use. Must be set.
//  func WithEmbedder(e embeddings.Embedder) Option {
//  	return func(p *Store) {
//  		p.embedder = e
//  	}
//  }

// WithResetChroma specifies whether chroma database is to be reset upon initialization of the vector store (true=yes)
// TODO (noodnik2): remove this functionality if not needed.
func WithResetChroma(resetFlag bool) Option {
	return func(p *Store) {
		p.resetChroma = resetFlag
	}
}

// WithDistanceFunction specifies the distance function which will be used
// see: https://github.com/amikos-tech/chroma-go/blob/d0087270239eccdb2f4f03d84b18d875c601ad6b/main.go#L96
func WithDistanceFunction(distanceFunction chromago.DistanceFunction) Option {
	return func(p *Store) {
		p.distanceFunction = distanceFunction
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
	o := &Store{}
	for _, opt := range opts {
		opt(o)
	}

	if o.collectionName == "" {
		return Store{}, fmt.Errorf("%w: missing collection name", ErrInvalidOptions)
	}

	if o.chromaURL == "" {
		return Store{}, fmt.Errorf("%w: missing chroma URL", ErrInvalidOptions)
	}

	if o.distanceFunction == "" {
		o.distanceFunction = chromago.COSINE
	}

	if o.openaiAPIKey == "" {
		o.openaiAPIKey = os.Getenv(_openAiAPIKeyEnvVrName)
		if o.openaiAPIKey == "" {
			return Store{}, fmt.Errorf(
				"%w: missing OpenAiApi key. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions,
				_openAiAPIKeyEnvVrName,
			)
		}
	}

	return *o, nil
}
