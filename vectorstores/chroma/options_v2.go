package chroma

import (
	"fmt"
	"os"

	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromatypes "github.com/amikos-tech/chroma-go/types"
	"github.com/tmc/langchaingo/embeddings"
)

// OptionV2 is a function type that can be used to modify the client.
type OptionV2 func(p *StoreV2)

// WithNameSpaceV2 sets the nameSpace used to upsert and query the vectors from.
func WithNameSpaceV2(nameSpace string) OptionV2 {
	return func(p *StoreV2) {
		p.nameSpace = nameSpace
	}
}

// WithChromaURLV2 is an option for specifying the Chroma URL. Must be set.
func WithChromaURLV2(chromaURL string) OptionV2 {
	return func(p *StoreV2) {
		p.chromaURL = chromaURL
	}
}

// WithEmbedderV2 is an option for setting the embedder to use.
func WithEmbedderV2(e embeddings.Embedder) OptionV2 {
	return func(p *StoreV2) {
		p.embedder = e
	}
}

// WithDistanceFunctionV2 specifies the distance function which will be used (default is L2)
// see: https://github.com/amikos-tech/chroma-go/blob/ab1339d0ee1a863be7d6773bcdedc1cfd08e3d77/types/types.go#L22
func WithDistanceFunctionV2(distanceFunction chromatypes.DistanceFunction) OptionV2 {
	return func(p *StoreV2) {
		p.distanceFunction = distanceFunction
	}
}

// WithIncludesV2 is an option for setting the includes to query the vectors.
func WithIncludesV2(includes []chromago.Include) OptionV2 {
	return func(p *StoreV2) {
		p.includes = includes
	}
}

// WithOpenAIAPIKeyV2 is an option for setting the OpenAI api key. If the option is not set
// the api key is read from the OPENAI_API_KEY environment variable. If the
// variable is not present, an error will be returned.
func WithOpenAIAPIKeyV2(openAiAPIKey string) OptionV2 {
	return func(p *StoreV2) {
		p.openaiAPIKey = openAiAPIKey
	}
}

// WithOpenAIOrganizationV2 is an option for setting the OpenAI organization id.
func WithOpenAIOrganizationV2(openAiOrganization string) OptionV2 {
	return func(p *StoreV2) {
		p.openaiOrganization = openAiOrganization
	}
}

func applyV2ClientOptions(opts ...OptionV2) (StoreV2, error) {
	o := &StoreV2{
		nameSpace:          DefaultNameSpace,
		nameSpaceKey:       DefaultNameSpaceKey,
		distanceFunction:   DefaultDistanceFunc,
		openaiAPIKey:       os.Getenv(OpenAIAPIKeyEnvVarName),
		openaiOrganization: os.Getenv(OpenAIOrgIDEnvVarName),
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.chromaURL == "" {
		o.chromaURL = os.Getenv(ChromaURLKeyEnvVarName)
		if o.chromaURL == "" {
			return StoreV2{}, fmt.Errorf(
				"%w: missing chroma URL. Pass it as an option or set the %s environment variable",
				ErrInvalidOptions, ChromaURLKeyEnvVarName)
		}
	}

	// a embedder or an openai api key must be provided
	if o.openaiAPIKey == "" && o.embedder == nil {
		return StoreV2{}, fmt.Errorf("%w: missing embedder or openai api key", ErrInvalidOptions)
	}

	return *o, nil
}
