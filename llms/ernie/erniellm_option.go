package ernie

import (
	"net/http"

	"github.com/tmc/langchaingo/callbacks"
)

const (
	ernieAPIKey    = "ERNIE_API_KEY"    //nolint:gosec
	ernieSecretKey = "ERNIE_SECRET_KEY" //nolint:gosec
)

type ModelName string

const (
	ModelNameERNIEBot       = "ERNIE-Bot"
	ModelNameERNIEBotTurbo  = "ERNIE-Bot-turbo"
	ModelNameERNIEBotPro    = "ERNIE-Bot-pro"
	ModelNameBloomz7B       = "BLOOMZ-7B"
	ModelNameLlama2_7BChat  = "Llama-2-7b-chat"
	ModelNameLlama2_13BChat = "Llama-2-13b-chat"
	ModelNameLlama2_70BChat = "Llama-2-70b-chat"
)

type options struct {
	apiKey           string
	secretKey        string
	accessToken      string
	modelName        ModelName
	callbacksHandler callbacks.Handler
	baseURL          string
	modelPath        string
	cacheType        string
	httpClient       *http.Client
}

type Option func(*options)

// WithAKSK passes the ERNIE API Key and Secret Key to the client. If not set, the keys
// are read from the ERNIE_API_KEY and ERNIE_SECRET_KEY environment variable.
// eg:
//
//	export ERNIE_API_KEY={Api Key}
//	export ERNIE_SECRET_KEY={Serect Key}
//
// Api Key,Serect Key from https://console.bce.baidu.com/qianfan/ais/console/applicationConsole/application
// More information available: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/flfmc9do2
func WithAKSK(apiKey, secretKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
		opts.secretKey = secretKey
	}
}

// WithAccessToken usually used for dev, Prod env recommend use WithAKSK.
func WithAccessToken(accessToken string) Option {
	return func(opts *options) {
		opts.accessToken = accessToken
	}
}

// WithModelName passes the Model Name to the client. If not set, use default ERNIE-Bot.
func WithModelName(modelName ModelName) Option {
	return func(opts *options) {
		opts.modelName = modelName
	}
}

// WithCallbackHandler passes the callback Handler to the client.
func WithCallbackHandler(callbacksHandler callbacks.Handler) Option {
	return func(opts *options) {
		opts.callbacksHandler = callbacksHandler
	}
}

// WithAPIKey passes the ERNIE API Key to the client.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithSecretKey passes the ERNIE Secret Key to the client.
func WithSecretKey(secretKey string) Option {
	return func(opts *options) {
		opts.secretKey = secretKey
	}
}

// WithModel passes the Model Name to the client. Alias for WithModelName.
func WithModel(modelName string) Option {
	return func(opts *options) {
		opts.modelName = ModelName(modelName)
	}
}

// WithBaseURL passes the base URL to the client.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// WithModelPath passes the model path to the client.
func WithModelPath(modelPath string) Option {
	return func(opts *options) {
		opts.modelPath = modelPath
	}
}

// WithCacheType passes the cache type to the client.
func WithCacheType(cacheType string) Option {
	return func(opts *options) {
		opts.cacheType = cacheType
	}
}

// WithHTTPClient passes a custom HTTP client to the client.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}
