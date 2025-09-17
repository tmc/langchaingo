package httputil

import (
	"net/http"
)

var (
	// DefaultTransport is the default HTTP transport for LangChainGo.
	// It wraps [http.DefaultTransport] and adds a User-Agent header containing
	// the LangChainGo version, program information, and system details.
	// This transport is suitable for use with httprr in tests.
	DefaultTransport http.RoundTripper = &Transport{
		Transport: http.DefaultTransport,
	}

	// DefaultClient is the default HTTP client for LangChainGo.
	// It uses [DefaultTransport] to automatically include proper User-Agent
	// headers in all requests. This client is recommended for all LangChainGo
	// HTTP operations unless custom transport behavior is required.
	DefaultClient = &http.Client{
		Transport: DefaultTransport,
	}
)

// Transport adds LangChainGo User-Agent headers to HTTP requests.
// If the request already has a User-Agent header, the LangChainGo
// User-Agent is appended to it.
//
// The zero value is a valid Transport that uses [http.DefaultTransport].
type Transport struct {
	// Transport is the underlying [http.RoundTripper] to use.
	// If nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	r2 := req.Clone(req.Context())
	if ua := req.Header.Get("User-Agent"); ua != "" {
		r2.Header.Set("User-Agent", ua+" "+UserAgent())
	} else {
		r2.Header.Set("User-Agent", UserAgent())
	}
	return transport.RoundTrip(r2)
}

// APIKeyTransport adds API keys to URL query parameters.
// It is commonly used with Google APIs and other services that accept API keys
// as query parameters.
//
// The zero value is a valid APIKeyTransport that uses [http.DefaultTransport].
type APIKeyTransport struct {
	// Transport is the underlying [http.RoundTripper] to use.
	// If nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper
	// APIKey is the API key to add to requests as a "key" query parameter.
	APIKey string
}

func (t *APIKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	if t.APIKey == "" || req.URL.Query().Get("key") != "" {
		return transport.RoundTrip(req)
	}
	r2 := req.Clone(req.Context())
	q := r2.URL.Query()
	q.Set("key", t.APIKey)
	r2.URL.RawQuery = q.Encode()
	return transport.RoundTrip(r2)
}
