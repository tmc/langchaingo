package openai

import (
	"errors"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*openaiclient.Client, error) {
	options := &options{
		token:        os.Getenv(tokenEnvVarName),
		model:        os.Getenv(modelEnvVarName),
		baseURL:      os.Getenv(baseURLEnvVarName),
		organization: os.Getenv(organizationEnvVarName),
		apiType:      openaiclient.APITypeOpenAI,
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	// set of options needed for Azure client
	if openaiclient.IsAzure(options.apiType) && options.apiVersion == "" {
		options.apiVersion = DefaultAPIVersion
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return openaiclient.New(options.token, options.model, options.baseURL, options.organization,
		options.apiType, options.apiVersion, options.httpClient)
}
