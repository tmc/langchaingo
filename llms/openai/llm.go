package openai

import (
	"errors"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

var (
	ErrEmptyResponse              = errors.New("no response")
	ErrMissingToken               = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable") //nolint:lll
	ErrMissingAzureModel          = errors.New("model needs to be provided when using Azure API")
	ErrMissingAzureEmbeddingModel = errors.New("embeddings model needs to be provided when using Azure API")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

// newClient creates an instance of the internal client.
func newClient(opts ...Option) (*options, *openaiclient.Client, error) {
	options := &options{
		token:        os.Getenv(tokenEnvVarName),
		model:        os.Getenv(modelEnvVarName),
		baseURL:      getEnvs(baseURLEnvVarName, baseAPIBaseEnvVarName),
		organization: os.Getenv(organizationEnvVarName),
		apiType:      APIType(openaiclient.APITypeOpenAI),
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	// set of options needed for Azure client
	if openaiclient.IsAzure(openaiclient.APIType(options.apiType)) && options.apiVersion == "" {
		options.apiVersion = DefaultAPIVersion
		if options.model == "" {
			return options, nil, ErrMissingAzureModel
		}
		if options.embeddingModel == "" {
			return options, nil, ErrMissingAzureEmbeddingModel
		}
	}

	if len(options.token) == 0 {
		return options, nil, ErrMissingToken
	}

	cli, err := openaiclient.New(options.token, options.model, options.baseURL, options.organization,
		openaiclient.APIType(options.apiType), options.apiVersion, options.httpClient, options.embeddingModel)
	return options, cli, err
}

func getEnvs(keys ...string) string {
	for _, key := range keys {
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
	}
	return ""
}
