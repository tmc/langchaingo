package googleai

// options is a set of options for GoogleAI clients.
type options struct {
	apiKey                string
	defaultModel          string
	defaultEmbeddingModel string
	defaultCandidateCount int
	defaultMaxTokens      int
	defaultTemperature    float64
	defaultTopK           int
	defaultTopP           float64
}

func defaultOptions() options {
	return options{
		apiKey:                "",
		defaultModel:          "gemini-pro",
		defaultEmbeddingModel: "embedding-001",
		defaultCandidateCount: 1,
		defaultMaxTokens:      256,
		defaultTemperature:    0.5,
		defaultTopK:           3,
		defaultTopP:           0.95,
	}
}

type Option func(*options)

// WithAPIKey passes the API KEY (token) to the client.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithDefaultModel passes a default content model name to the client. This
// model name is used if not explicitly provided in specific client invocations.
func WithDefaultModel(defaultModel string) Option {
	return func(opts *options) {
		opts.defaultModel = defaultModel
	}
}

// WithDefaultModel passes a default embedding model name to the client. This
// model name is used if not explicitly provided in specific client invocations.
func WithDefaultEmbeddingModel(defaultEmbeddingModel string) Option {
	return func(opts *options) {
		opts.defaultEmbeddingModel = defaultEmbeddingModel
	}
}
