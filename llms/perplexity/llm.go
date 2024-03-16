package perplexity

import (
	"errors"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/llms/perplexity/internal/perplexityclient"
)

var (
	ErrEmptyResponse              = errors.New("no response")
	ErrMissingToken               = errors.New("missing the perplexity API key, set it in the PERPLEXITY_API_KEY environment variable") //nolint:lll
	ErrMissingAzureEmbeddingModel = errors.New("embeddings model needs to be provided when using Azure API")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

// newClient is wrapper for perplexityclient internal package.
func newClient(opts ...Option) (*options, *perplexityclient.Client, error) {
	options := &options{
		token:      os.Getenv(tokenEnvVarName),
		model:      os.Getenv(modelEnvVarName),
		baseURL:    "https://api.perplexity.ai",
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return options, nil, ErrMissingToken
	}

	cli, err := perplexityclient.New(options.token, options.model, options.baseURL, options.httpClient)
	return options, cli, err
}
