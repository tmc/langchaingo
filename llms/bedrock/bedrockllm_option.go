package bedrock

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/callbacks"
)

type Option func(*options)

type options struct {
	modelID         string
	client          *bedrockruntime.Client
	callbackHandler callbacks.Handler
}

// WithModel allows setting a custom modelId.
func WithModel(modelID string) Option {
	return func(o *options) {
		o.modelID = modelID
	}
}

// WithClient allows setting a custom bedrockruntime.Client.
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
