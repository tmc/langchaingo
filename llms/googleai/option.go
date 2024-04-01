package googleai

// Options is a set of options for GoogleAI and Vertex clients.
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
}

func DefaultOptions() Options {
	return Options{
		APIKey:                "",
		CloudProject:          "",
		CloudLocation:         "",
		DefaultModel:          "gemini-pro",
		DefaultEmbeddingModel: "embedding-001",
		DefaultCandidateCount: 1,
		DefaultMaxTokens:      1024,
		DefaultTemperature:    0.5,
		DefaultTopK:           3,
		DefaultTopP:           0.95,
	}
}

type Option func(*Options)

// WithAPIKey passes the API KEY (token) to the client. This is useful for
// googleai clients.
func WithAPIKey(apiKey string) Option {
	return func(opts *Options) {
		opts.APIKey = apiKey
	}
}

// WithCloudProject passes the GCP cloud project name to the client. This is
// useful for vertex clients.
func WithCloudProject(p string) Option {
	return func(opts *Options) {
		opts.CloudProject = p
	}
}

// WithCloudLocation passes the GCP cloud location (region) name to the client.
// This is useful for vertex clients.
func WithCloudLocation(l string) Option {
	return func(opts *Options) {
		opts.CloudLocation = l
	}
}

// WithDefaultModel passes a default content model name to the client. This
// model name is used if not explicitly provided in specific client invocations.
func WithDefaultModel(defaultModel string) Option {
	return func(opts *Options) {
		opts.DefaultModel = defaultModel
	}
}

// WithDefaultModel passes a default embedding model name to the client. This
// model name is used if not explicitly provided in specific client invocations.
func WithDefaultEmbeddingModel(defaultEmbeddingModel string) Option {
	return func(opts *Options) {
		opts.DefaultEmbeddingModel = defaultEmbeddingModel
	}
}
