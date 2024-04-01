package vertex

// options is a set of options for GoogleAI clients.
type options struct {
	cloudProject          string
	cloudLocation         string
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
		cloudProject:          "",
		cloudLocation:         "",
		defaultModel:          "gemini-pro",
		defaultEmbeddingModel: "embedding-001",
		defaultCandidateCount: 1,
		defaultMaxTokens:      1024,
		defaultTemperature:    0.5,
		defaultTopK:           3,
		defaultTopP:           0.95,
	}
}

type Option func(*options)

// WithCloudProject passes the GCP cloud project name to the client.
func WithCloudProject(p string) Option {
	return func(opts *options) {
		opts.cloudProject = p
	}
}

// WithCloudLocation passes the GCP cloud location (region) name to the client.
func WithCloudLocation(l string) Option {
	return func(opts *options) {
		opts.cloudLocation = l
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
