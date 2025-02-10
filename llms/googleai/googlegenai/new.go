// package googleai implements a langchaingo provider for Google AI LLMs.
// See https://ai.google.dev/ for more details.
package googlegenai

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"golang.org/x/oauth2/google"
	"google.golang.org/genai"
)

// HTTP options to be used in each of the requests.
type HTTPOptions struct {
	// BaseURL specifies the base URL for the API endpoint. If unset, defaults to "https://generativelanguage.googleapis.com/"
	// for the Gemini API backend, and location-specific Vertex AI endpoint (e.g., "https://us-central1-aiplatform.googleapis.com/
	BaseURL string `json:"baseUrl,omitempty"`
	// APIVersion specifies the version of the API to use.
	APIVersion string `json:"apiVersion,omitempty"`
	// Timeout sets the timeout for HTTP requests in milliseconds. If unset, defaults to
	// "v1beta" for the Gemini API, and "v1beta1" for the Vertex AI.
	Timeout int64 `json:"timeout,omitempty"`
}

// GoogleAI is a type that represents a Google AI API client.
type GoogleAI struct {
	CallbacksHandler callbacks.Handler
	client           *genai.Client
	opts             Options
}

var _ llms.Model = &GoogleAI{}

// New creates a new GoogleAI client.
func New(ctx context.Context, opts ...Option) (*GoogleAI, error) {
	clientOptions := DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}
	clientOptions.EnsureAuthPresent()

	gi := &GoogleAI{
		opts: clientOptions,
	}

	var googleCredentials *google.Credentials
	if clientOptions.Credentials.CredentialsJSON != nil {
		var err error
		googleCredentials, err = google.CredentialsFromJSON(ctx, clientOptions.Credentials.CredentialsJSON, clientOptions.Credentials.Scopes...)
		if err != nil {
			return nil, err
		}
	}

	var httpOptions genai.HTTPOptions
	if clientOptions.HTTPOPtions != nil {
		httpOptions = genai.HTTPOptions{
			BaseURL:    clientOptions.HTTPOPtions.BaseURL,
			APIVersion: clientOptions.HTTPOPtions.APIVersion,
			Timeout:    clientOptions.HTTPOPtions.Timeout,
		}
	}

	cfg := &genai.ClientConfig{
		Credentials: googleCredentials,
		Backend:     genai.Backend(clientOptions.ApiBackend),
		Project:     clientOptions.CloudProject,
		Location:    clientOptions.CloudLocation,
		APIKey:      clientOptions.Credentials.APIKey,
		HTTPClient:  clientOptions.HTTPClient,
		HTTPOptions: httpOptions,
	}
	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return gi, err
	}

	gi.client = client
	return gi, nil
}
