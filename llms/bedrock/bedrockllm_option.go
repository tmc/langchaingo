package bedrock

import (
	"github.com/vxcontrol/langchaingo/callbacks"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// Option is an option for the Bedrock LLM.
type Option func(*options)

type options struct {
	modelProvider   string
	modelID         string
	client          *bedrockruntime.Client
	callbackHandler callbacks.Handler
	useConverseAPI  bool
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

// WithConverseAPI enables the use of the unified Bedrock Converse API
// instead of the model-specific legacy implementations.
//
// The Converse API provides:
// - Unified interface for all supported Bedrock models
// - Built-in tool calling support
// - Streaming responses with ConverseStream
// - Reasoning content support for Claude 3.7+ and Nova models
// - Multimodal input support (text, images, documents)
// - Better error handling and response consistency
//
// Supported models: All Anthropic Claude, Amazon Nova, Meta Llama,
// Cohere Command, and AI21 Jamba models available through Bedrock.
//
// Note: This is the recommended approach for new applications.
func WithConverseAPI() Option {
	return func(o *options) {
		o.useConverseAPI = true
	}
}
