package awsai

import (
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms/awsai/internal/awsclient"
)

const (
	ApiTypeSagemaker = "sagemaker"
	ApiTypeBedrock   = "bedrock"
)

// Option is an option for the Bedrock LLM.
type Option func(*options)

type options struct {
	modelID         string
	client          interface{}
	callbackHandler callbacks.Handler
	apitype         string
}

// WithApiType allows setting a custom API type.
func WithApiType(apiType string) Option {
	return func(o *options) {
		o.apitype = apiType
	}
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

// WithClient allows setting a custom bedrockruntime.Client.
//
// You may use this to pass a custom bedrockruntime.Client
// with custom configuration options
// such as setting custom credentials, region, endpoint, etc.
//
// By default, a new client will be created using the default credentials chain.
func WithClient(client interface{}) Option {
	return func(o *options) {
		o.client = client.(awsclient.AwsRuntimeClient)
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) Option {
	return func(o *options) {
		o.callbackHandler = callbackHandler
	}
}
