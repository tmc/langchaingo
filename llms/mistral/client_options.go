package mistral

import (
	"time"

	"github.com/0xDezzy/langchaingo/callbacks"
)

type clientOptions struct {
	apiKey           string
	endpoint         string
	maxRetries       int
	timeout          time.Duration
	model            string
	callbacksHandler callbacks.Handler
}

type Option func(*clientOptions)

// Passes the API Key (token) to the Mistral client. Defaults to os.getEnv("MISTRAL_API_KEY").
func WithAPIKey(apiKey string) Option {
	return func(o *clientOptions) {
		o.apiKey = apiKey
	}
}

// Sets the API endpoint for the Model being instantiated, defaults to "https://api.mistral.ai" (subject to change; this default is pulled from the mistral-go client library, https://github.com/Gage-Technologies/mistral-go).
func WithEndpoint(endpoint string) Option {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// Sets the maximum number of retries the client is permitted to perform, used in case a call to the model's API fails.
func WithMaxRetries(maxRetries int) Option {
	return func(o *clientOptions) {
		o.maxRetries = maxRetries
	}
}

// Sets the timeout duration for the client. This determines how long the client will wait for a response from the model's API before timing out.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// Sets the model name for the Model being instantiated. Defaults to "open-mistral-7b". See https://docs.mistral.ai/platform/endpoints/ for a full list of supported models.
func WithModel(model string) Option {
	return func(o *clientOptions) {
		o.model = model
	}
}

// Sets the Langchain callbacks handler to use for the Model being instantiated. Defaults to callbacks.SimpleHandler.
func WithCallbacksHandler(callbacksHandler callbacks.Handler) Option {
	return func(o *clientOptions) {
		o.callbacksHandler = callbacksHandler
	}
}
