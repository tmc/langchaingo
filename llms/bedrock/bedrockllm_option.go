package bedrock

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/vendasta/langchaingo/callbacks"
)

// Option is an option for the Bedrock LLM.
type Option func(*options)

type options struct {
	modelProvider   string
	modelID         string
	client          *bedrockruntime.Client
	callbackHandler callbacks.Handler
}

// WithModel allows setting a custom modelId.
//
// If not set, the default model is used
// i.e. "amazon.titan-text-lite-v1".
func WithModel(modelID string) Option {
	return func(o *options) {
		o.modelID = modelID
	}
}

// WithModelProvider allows setting a custom model provider.
//
// If not set, the default model provider is used
// i.e. "anthropic".
func WithModelProvider(modelProvider string) Option {
	return func(o *options) {
		o.modelProvider = modelProvider
	}
}

// WithClient allows setting a custom bedrockruntime.Client.
//
// You may use this to pass a custom bedrockruntime.Client
// with custom configuration options
// such as setting custom credentials, region, endpoint, etc.
//
// By default, a new client will be created using the default credentials chain.
func WithClient(client *bedrockruntime.Client) Option {
	return func(o *options) {
		o.client = client
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) Option {
	return func(o *options) {
		o.callbackHandler = callbackHandler
	}
}
