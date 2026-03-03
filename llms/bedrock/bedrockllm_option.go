package bedrock

import (
	"time"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// Option is an option for the Bedrock LLM.
type Option func(*options)

type options struct {
	modelProvider     string
	modelID           string
	client            *bedrockruntime.Client
	callbackHandler   callbacks.Handler
	useConverseAPI    bool
	enableAutoCaching bool
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
// - Prompt caching support via cachePoint (requires AWS SDK types)
//
// Supported models: All Anthropic Claude, Amazon Nova, Meta Llama,
// Cohere Command, and AI21 Jamba models available through Bedrock.
//
// Note: This is the recommended approach for new applications.
//
// Prompt Caching:
// - Legacy API (InvokeModel) supports Anthropic's cache_control format
// - Converse API supports cachePoint via SystemContentBlockMemberCachePoint
// - Cache metrics are returned in response.Usage (CacheReadInputTokens, CacheWriteInputTokens)
// - Requires minimum tokens per checkpoint (1024 for Sonnet 4.5, 4096 for Haiku 4.5)
// - Supports 5m and 1h TTL for Claude 4.x models
func WithConverseAPI() Option {
	return func(o *options) {
		o.useConverseAPI = true
	}
}

// WithAutomaticCaching enables automatic prompt caching for supported Anthropic models.
//
// When enabled, caching is automatically applied for models matching these patterns:
// - claude-opus-4 (includes 4.6, 4.5, 4.1, 4.0)
// - claude-sonnet-4 (includes 4.6, 4.5, 4.0)
// - claude-haiku-4 (includes 4.5)
//
// The caching strategy automatically:
// - Adds cache points to system prompts
// - Adds cache points to conversation history (last message before new user input)
// - Uses ephemeral 5-minute TTL by default
//
// Benefits:
// - 90% cost reduction on cached input tokens
// - No manual cache control wrapper needed on client side
// - Transparent caching without modifying message chains
//
// Note: Automatic caching works with both Legacy and Converse APIs.
func WithAutomaticCaching() Option {
	return func(o *options) {
		o.enableAutoCaching = true
	}
}

// EphemeralCache creates a standard ephemeral cache control for Bedrock with 5-minute duration.
func EphemeralCache() *llms.CacheControl {
	return &llms.CacheControl{
		Type:     "ephemeral",
		Duration: 5 * time.Minute,
	}
}

// EphemeralCacheOneHour creates a 1-hour ephemeral cache control for Bedrock.
// Supported by Claude Opus 4.5, Haiku 4.5, and Sonnet 4.5.
func EphemeralCacheOneHour() *llms.CacheControl {
	return &llms.CacheControl{
		Type:     "ephemeral",
		Duration: time.Hour,
	}
}

// CachedContent represents content with caching instructions for Bedrock.
// This wraps any ContentPart and adds cache control metadata.
//
// Note: For most use cases, prefer using bedrock.WithAutomaticCaching() option
// which automatically applies caching to supported Anthropic models (Claude 4.x).
// This manual wrapper is only needed for fine-grained cache control.
//
// Automatic caching is supported in both Legacy and Converse APIs.
type CachedContent struct {
	llms.ContentPart
	CacheControl *llms.CacheControl `json:"cache_control,omitempty"`
}

// WithCacheControl wraps content with cache control instructions for Bedrock.
// This allows explicit control over what content should be cached.
//
// Recommended: Use bedrock.WithAutomaticCaching() option instead for transparent caching.
//
// Manual usage (when fine-grained control is needed):
//
//	bedrock.WithCacheControl(
//	    llms.TextPart("long context..."),
//	    bedrock.EphemeralCache(),
//	)
//
// Supported models: Claude Opus 4, Sonnet 4, Haiku 4 and their variants.
func WithCacheControl(content llms.ContentPart, control *llms.CacheControl) CachedContent {
	return CachedContent{
		ContentPart:  content,
		CacheControl: control,
	}
}
