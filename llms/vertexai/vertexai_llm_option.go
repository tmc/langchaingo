package vertexai

import (
	"github.com/tmc/langchaingo/llms/vertexai/internal/aiplatformclient"
	"net/http"
	"os"
	"sync"

	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

const (
	projectIDEnvVarName = "GOOGLE_CLOUD_PROJECT" //nolint:gosec
)

var (
	// nolint: gochecknoglobals
	initOptions sync.Once

	// nolint: gochecknoglobals
	defaultOptions *options
)

type options struct {
	model          string
	chatModel      string
	embeddingModel string
	publisher      string

	projectID     string
	clientOptions []option.ClientOption
}

// Option is a function that can be passed to NewClient to configure options.
type Option func(*options)

// initOpts initializes defaultOptions with the environment variables.
func initOpts() {
	defaultOptions = &options{
		model:          aiplatformclient.TextModelName,
		chatModel:      aiplatformclient.ChatModelName,
		embeddingModel: aiplatformclient.EmbeddingModelName,

		publisher: "google",

		projectID: os.Getenv(projectIDEnvVarName),
	}
}

// WithModel passes the VertexAI model to the client. If not set, the model
// will default to the one used by PaLM.  This is to preserve existing behavior.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithProjectID passes the Google Cloud project ID to the client. If not set, the project
// is read from the GOOGLE_CLOUD_PROJECT environment variable.
func WithProjectID(projectID string) Option {
	return func(opts *options) {
		opts.projectID = projectID
	}
}

// WithAPIKey returns a ClientOption that specifies an API key to be used
// as the basis for authentication.
func WithAPIKey(apiKey string) Option {
	return convertStringOption(option.WithAPIKey)(apiKey)
}

// WithCredentialsFile returns a ClientOption that authenticates
// API calls with the given service account or refresh token JSON
// credentials file.
func WithCredentialsFile(path string) Option {
	return convertStringOption(option.WithCredentialsFile)(path)
}

// WithCredentialsJSON returns a ClientOption that authenticates
// API calls with the given service account or refresh token JSON
// credentials.
func WithCredentialsJSON(json []byte) Option {
	return convertByteArrayOption(option.WithCredentialsJSON)(json)
}

func WithGRPCDialOption(opt grpc.DialOption) Option {
	return func(opts *options) {
		opts.clientOptions = append(opts.clientOptions, option.WithGRPCDialOption(opt))
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.clientOptions = append(opts.clientOptions, option.WithHTTPClient(client))
	}
}

func convertStringOption(fopt func(string) option.ClientOption) func(string) Option {
	return func(param string) Option {
		return func(opts *options) {
			opts.clientOptions = append(opts.clientOptions, fopt(param))
		}
	}
}

func convertByteArrayOption(fopt func([]byte) option.ClientOption) func([]byte) Option {
	return func(param []byte) Option {
		return func(opts *options) {
			opts.clientOptions = append(opts.clientOptions, fopt(param))
		}
	}
}
