package chroma

import (
	"errors"
	"fmt"
	"os"

	chromago "github.com/amikos-tech/chroma-go"
)

const (
	_openAiApiKeyEnvVrName = "OPENAI_API_KEY"
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

// TODO (noodnik2): determine whether (or not) to use this instead of the more Chroma-specific "WithCollectionName"
//// NameSpace is an option for setting the nameSpace to upsert and query the vectors
//// from. Must be set.
//func WithNameSpace(nameSpace string) Option {
//	return func(p *Store) {
//		p.nameSpace = nameSpace
//	}
//}

// WithChromaUrl is an option for specifying the Chroma URL. Must be set.
func WithChromaUrl(chromaUrl string) Option {
	return func(p *Store) {
		p.chromaUrl = chromaUrl
	}
}

// TODO (noodnik2): clarify need and implement if so
//// WithProjectName is an option for specifying the project name. Must be set. The
//// project name associated with the api key can be obtained using the whoami
//// operation.
//func WithProjectName(name string) Option {
//	return func(p *Store) {
//		p.projectName = name
//	}
//}
//

// TODO (noodnik2): implement
//// WithEmbedder is an option for setting the embedder to use. Must be set.
//func WithEmbedder(e embeddings.Embedder) Option {
//	return func(p *Store) {
//		p.embedder = e
//	}
//}

// WithResetChroma specifies whether chroma database is to be reset upon initialization of the vector store (true=yes)
// TODO (noodnik2): see if this stands the test of scrutiny / is justified
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

// WithOpenAiApiKey is an option for setting the OpenAI api key. If the option is not set
// the api key is read from the OPENAI_API_KEY environment variable. If the
// variable is not present, an error will be returned.
func WithOpenAiApiKey(openAiApiKey string) Option {
	return func(p *Store) {
		p.openaiApiKey = openAiApiKey
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		//textKey: _defaultTextKey,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.collectionName == "" {
		return Store{}, fmt.Errorf("%w: missing collection name", ErrInvalidOptions)
	}

	if o.chromaUrl == "" {
		return Store{}, fmt.Errorf("%w: missing chroma URL", ErrInvalidOptions)
	}

	if o.distanceFunction == "" {
		o.distanceFunction = chromago.COSINE
	}

	if o.openaiApiKey == "" {
		o.openaiApiKey = os.Getenv(_openAiApiKeyEnvVrName)
		if o.openaiApiKey == "" {
			return Store{}, fmt.Errorf(
				"%w: missing OpenAiApi key. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions,
				_openAiApiKeyEnvVrName,
			)
		}
	}

	return *o, nil
}
