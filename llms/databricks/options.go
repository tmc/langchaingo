package databricks

import (
	"fmt"
	"net/http"
)

// WithFullURL sets the full URL for the LLM.
func WithFullURL(fullURL string) Option {
	return func(llm *LLM) {
		llm.url = fullURL
	}
}

// WithURLComponents constructs the URL from individual components.
func WithURLComponents(databricksInstance, modelName, modelVersion string) Option {
	return func(llm *LLM) {
		llm.url = fmt.Sprintf("https://%s/model/%s/%s/invocations", databricksInstance, modelName, modelVersion)
	}
}

// WithToken pass the token for authentication.
func WithToken(token string) Option {
	return func(llm *LLM) {
		llm.token = token
	}
}

// WithHTTPClient sets the HTTP client for the LLM.
func WithHTTPClient(client *http.Client) Option {
	return func(llm *LLM) {
		llm.httpClient = client
	}
}
