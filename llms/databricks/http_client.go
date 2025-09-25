package databricks

import (
	"fmt"
	"io"
	"net/http"
)

// TokenRoundTripper is a custom RoundTripper that adds a Bearer token to each request.
type TokenRoundTripper struct {
	Token     string
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and adds the Bearer token to the request.
func (t *TokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	req.Header.Set("content-type", "application/json")
	// Use the underlying transport to perform the request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Handle status codes
	if resp.StatusCode >= 400 {
		// Read the response body for detailed error message (optional)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close() // Ensure the body is closed to avoid resource leaks

		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// NewClientWithToken creates a new HTTP client with a Bearer token.
func NewHTTPClient(token string) *http.Client {
	return &http.Client{
		Transport: &TokenRoundTripper{
			Token:     token,
			Transport: http.DefaultTransport, // Use http.DefaultTransport as the fallback
		},
	}
}
