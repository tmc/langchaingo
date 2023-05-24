package vertexai

import (
	"os"
	"sync"

	"google.golang.org/api/option"
)

const (
	projectIdEnvVarName = "GOOGLE_CLOUD_PROJECT" //nolint:gosec
)

var (
	// nolint: gochecknoglobals
	initOptions sync.Once

	// nolint: gochecknoglobals
	defaultOptions *options
)

type options struct {
	projectID     string
	clientOptions []option.ClientOption
}

type Option func(*options)

// initOpts initializes defaultOptions with the environment variables.
func initOpts() {
	defaultOptions = &options{
		projectID: os.Getenv(projectIdEnvVarName),
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
var WithAPIKey = convertStringOption(option.WithAPIKey)

// WithCredentialsFile returns a ClientOption that authenticates
// API calls with the given service account or refresh token JSON
// credentials file.
var WithCredentialsFile = convertStringOption(option.WithCredentialsFile)

// WithCredentialsJSON returns a ClientOption that authenticates
// API calls with the given service account or refresh token JSON
// credentials.
var WithCredentialsJSON = convertByteArrayOption(option.WithCredentialsJSON)

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
