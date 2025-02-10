package googlegenai

import (
	"net/http"
	"os"
	"reflect"

	"google.golang.org/api/option"
	"google.golang.org/genai"
)

type ApiBackend genai.Backend

const (
	VERTEX_BACKEND = ApiBackend(genai.BackendVertexAI)
	GEMINI_BACKEND = ApiBackend(genai.BackendGeminiAPI)
)

type GoogleCredentials struct {
	Scopes          []string
	CredentialsJson []byte
	ApiKey          string
}

// Options is a set of options for GoogleAI and Vertex clients.
type Options struct {
	CloudProject          string
	CloudLocation         string
	Credentials           GoogleCredentials
	DefaultModel          string
	DefaultEmbeddingModel string
	DefaultCandidateCount int
	DefaultMaxTokens      int
	DefaultTemperature    float64
	DefaultTopK           int
	DefaultTopP           float64
	HarmThreshold         HarmBlockThreshold
	ApiBackend            ApiBackend
	ClientOptions         []option.ClientOption
	HttpClient            *http.Client
	HttpOptions           *HTTPOptions
}

func DefaultOptions() Options {
	return Options{
		CloudProject:          "",
		CloudLocation:         "",
		DefaultModel:          "gemini-pro",
		DefaultEmbeddingModel: "embedding-001",
		DefaultCandidateCount: 1,
		DefaultMaxTokens:      2048,
		DefaultTemperature:    0.5,
		DefaultTopK:           3,
		DefaultTopP:           0.95,
		HarmThreshold:         HarmBlockOnlyHigh,
		ApiBackend:            GEMINI_BACKEND,
	}
}

// EnsureAuthPresent attempts to ensure that the client has authentication information. If it does not, it will attempt to use the GOOGLE_API_KEY environment variable.
func (o *Options) EnsureAuthPresent() {
	if !hasAuthOptions(o.ClientOptions) {
		if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
			WithAPIKey(key)(o)
		}
	}
}

type Option func(*Options)

// WithAPIKey passes the API KEY (token) to the client. This is useful for
// googleai clients.
func WithAPIKey(apiKey string) Option {
	return func(opts *Options) {
		opts.Credentials.ApiKey = apiKey
	}
}

// WithCredentialsJSON append a ClientOption that authenticates
// API calls with the given service account or refresh token JSON
// credentials.
func WithCredentialsJSON(credentialsJSON []byte, scopes []string) Option {
	return func(opts *Options) {
		if len(credentialsJSON) == 0 {
			return
		}
		opts.Credentials = GoogleCredentials{
			Scopes:          scopes,
			CredentialsJson: credentialsJSON,
		}
	}
}

// WithCredentialsFile append a ClientOption that authenticates
// API calls with the given service account or refresh token JSON
// credentials file.
func WithCredentialsFile(credentialsFile []byte, scopes []string) Option {
	return func(opts *Options) {
		if credentialsFile == nil {
			return
		}
		opts.Credentials.CredentialsJson = credentialsFile
		opts.Credentials.Scopes = scopes
	}
}

// WithHTTPClient append a ClientOption that uses the provided HTTP client to
// make requests.
// This is useful for vertex clients.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(opts *Options) {
		opts.HttpClient = httpClient
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

// WithDefaultCandidateCount sets the candidate count for the model.
func WithDefaultCandidateCount(defaultCandidateCount int) Option {
	return func(opts *Options) {
		opts.DefaultCandidateCount = defaultCandidateCount
	}
}

// WithDefaultMaxTokens sets the maximum token count for the model.
func WithDefaultMaxTokens(maxTokens int) Option {
	return func(opts *Options) {
		opts.DefaultMaxTokens = maxTokens
	}
}

// WithDefaultTemperature sets the maximum token count for the model.
func WithDefaultTemperature(defaultTemperature float64) Option {
	return func(opts *Options) {
		opts.DefaultTemperature = defaultTemperature
	}
}

// WithDefaultTemperature sets the maximum token count for the model.
func WithApiBackend(backend ApiBackend) Option {
	return func(opts *Options) {
		opts.ApiBackend = backend
	}
}

// WithDefaultTopK sets the TopK for the model.
func WithDefaultTopK(defaultTopK int) Option {
	return func(opts *Options) {
		opts.DefaultTopK = defaultTopK
	}
}

// WithDefaultTopP sets the TopP for the model.
func WithDefaultTopP(defaultTopP float64) Option {
	return func(opts *Options) {
		opts.DefaultTopP = defaultTopP
	}
}

// WithHarmThreshold sets the safety/harm setting for the model, potentially
// limiting any harmful content it may generate.
func WithHarmThreshold(ht HarmBlockThreshold) Option {
	return func(opts *Options) {
		opts.HarmThreshold = ht
	}
}

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

// helper to inspect incoming client options for auth options.
func hasAuthOptions(opts []option.ClientOption) bool {
	for _, opt := range opts {
		v := reflect.ValueOf(opt)
		ts := v.Type().String()

		switch ts {
		case "option.withAPIKey":
			return v.String() != ""

		case "option.withHTTPClient",
			"option.withTokenSource",
			"option.withCredentialsFile",
			"option.withCredentialsJSON":
			return true
		}
	}
	return false
}
