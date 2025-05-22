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

func WithNameSpaceV2(nameSpace string) OptionV2 {
	return func(p *StoreV2) {
		p.nameSpace = nameSpace
	}
}

func WithChromaURLV2(chromaURL string) OptionV2 {
	return func(p *StoreV2) {
		p.chromaURL = chromaURL
	}
}

func WithEmbedderV2(e embeddings.Embedder) OptionV2 {
	return func(p *StoreV2) {
		p.embedder = e
	}
}

func WithDistanceFunctionV2(distanceFunction chromatypes.DistanceFunction) OptionV2 {
	return func(p *StoreV2) {
		p.distanceFunction = distanceFunction
	}
}

func WithIncludesV2(includes []chromago.Include) OptionV2 {
	return func(p *StoreV2) {
		p.includes = includes
	}
}

func WithOpenAIAPIKeyV2(openAiAPIKey string) OptionV2 {
	return func(p *StoreV2) {
		p.openaiAPIKey = openAiAPIKey
	}
}

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
