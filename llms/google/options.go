package google

import (
	"os"

	"github.com/vendasta/langchaingo/llms"
)

// Options is a unified set of options for the Google AI client.
// These options are translated to the appropriate underlying provider options.
type Options struct {
	APIKey                string
	CloudProject          string
	CloudLocation         string
	DefaultModel          string
	DefaultEmbeddingModel string
	DefaultCandidateCount int
	DefaultMaxTokens      int
	DefaultTemperature    float64
	DefaultTopK           int
	DefaultTopP           float64
	HarmThreshold         HarmBlockThreshold
}

// DefaultOptions returns the default options for the Google AI client.
func DefaultOptions() Options {
	return Options{
		DefaultModel:          "gemini-2.0-flash",
		DefaultEmbeddingModel: "embedding-001",
		DefaultCandidateCount: 1,
		DefaultMaxTokens:      2048,
		DefaultTemperature:    0.5,
		DefaultTopK:           3,
		DefaultTopP:           0.95,
		HarmThreshold:         HarmBlockOnlyHigh,
	}
}

// Option is a function that configures Options.
type Option func(*Options)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) Option {
	return func(opts *Options) {
		opts.APIKey = apiKey
	}
}

// WithCloudProject sets the GCP cloud project name.
// This is useful for Vertex AI clients.
func WithCloudProject(p string) Option {
	return func(opts *Options) {
		opts.CloudProject = p
	}
}

// WithCloudLocation sets the GCP cloud location (region) name.
// This is useful for Vertex AI clients.
func WithCloudLocation(l string) Option {
	return func(opts *Options) {
		opts.CloudLocation = l
	}
}

// WithDefaultModel sets the default content model name.
// This model name is used if not explicitly provided in specific client invocations.
// Note: The model name determines which underlying SDK is used:
// - gemini-3+ models use googleaiv2 (google.golang.org/genai)
// - older models use googleai (github.com/google/generative-ai-go)
func WithDefaultModel(defaultModel string) Option {
	return func(opts *Options) {
		opts.DefaultModel = defaultModel
	}
}

// WithDefaultEmbeddingModel sets the default embedding model name.
func WithDefaultEmbeddingModel(defaultEmbeddingModel string) Option {
	return func(opts *Options) {
		opts.DefaultEmbeddingModel = defaultEmbeddingModel
	}
}

// WithDefaultCandidateCount sets the default candidate count.
func WithDefaultCandidateCount(count int) Option {
	return func(opts *Options) {
		opts.DefaultCandidateCount = count
	}
}

// WithDefaultMaxTokens sets the default maximum token count.
func WithDefaultMaxTokens(maxTokens int) Option {
	return func(opts *Options) {
		opts.DefaultMaxTokens = maxTokens
	}
}

// WithDefaultTemperature sets the default temperature.
func WithDefaultTemperature(temperature float64) Option {
	return func(opts *Options) {
		opts.DefaultTemperature = temperature
	}
}

// WithDefaultTopK sets the default TopK.
func WithDefaultTopK(topK int) Option {
	return func(opts *Options) {
		opts.DefaultTopK = topK
	}
}

// WithDefaultTopP sets the default TopP.
func WithDefaultTopP(topP float64) Option {
	return func(opts *Options) {
		opts.DefaultTopP = topP
	}
}

// WithHarmThreshold sets the safety/harm threshold for the model.
func WithHarmThreshold(ht HarmBlockThreshold) Option {
	return func(opts *Options) {
		opts.HarmThreshold = ht
	}
}

// HarmBlockThreshold represents the harm blocking threshold.
type HarmBlockThreshold int32

const (
	// HarmBlockUnspecified means threshold is unspecified.
	HarmBlockUnspecified HarmBlockThreshold = 0
	// HarmBlockLowAndAbove means content with NEGLIGIBLE will be allowed.
	HarmBlockLowAndAbove HarmBlockThreshold = 1
	// HarmBlockMediumAndAbove means content with NEGLIGIBLE and LOW will be allowed.
	HarmBlockMediumAndAbove HarmBlockThreshold = 2
	// HarmBlockOnlyHigh means content with NEGLIGIBLE, LOW, and MEDIUM will be allowed.
	HarmBlockOnlyHigh HarmBlockThreshold = 3
	// HarmBlockNone means all content will be allowed.
	HarmBlockNone HarmBlockThreshold = 4
)

// WithCachedContent enables the use of pre-created cached content.
// The cached content must be created separately using the caching helper.
func WithCachedContent(name string) llms.CallOption {
	return func(o *llms.CallOptions) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]interface{})
		}
		o.Metadata["CachedContentName"] = name
	}
}

// EnsureAuthPresent attempts to ensure that the client has authentication.
// If no API key is set, it will attempt to use the GOOGLE_API_KEY environment variable.
func (o *Options) EnsureAuthPresent() {
	if o.APIKey == "" {
		if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
			o.APIKey = key
		}
	}
}
