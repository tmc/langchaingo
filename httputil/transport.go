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

// Transport is an [http.RoundTripper] that adds LangChainGo User-Agent headers
// to outgoing HTTP requests. It wraps another RoundTripper (typically
// [http.DefaultTransport]) and can be used to add User-Agent headers to any
// HTTP client.
//
// If the wrapped request already has a User-Agent header, the LangChainGo
// User-Agent is appended to it rather than replacing it.
type Transport struct {
	// Transport is the underlying [http.RoundTripper] to use.
	// If nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper
}

// RoundTrip implements the [http.RoundTripper] interface. It adds the LangChainGo
// User-Agent header to the request and then delegates to the underlying transport.
// If the request already has a User-Agent header, the LangChainGo information is
// appended to preserve existing client identification.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	newReq := req.Clone(req.Context())
	ua := UserAgent()
	// Append to existing User-Agent if present, otherwise set it
	existingUA := req.Header.Get("User-Agent")
	if existingUA != "" {
		newReq.Header.Set("User-Agent", existingUA+" "+ua)
	} else {
		newReq.Header.Set("User-Agent", ua)
	}
	return transport.RoundTrip(newReq)
}

// ApiKeyTransport is an [http.RoundTripper] that adds API keys to URL query parameters.
// This is commonly used with Google APIs and other services that accept API keys
// as query parameters. It wraps another RoundTripper and automatically adds
// the API key if not already present in the request.
//
// This transport is particularly useful when working with client libraries that
// don't properly set API keys when using custom HTTP clients, such as the
// Google AI client library when used with httprr for testing.
type ApiKeyTransport struct {
	// Transport is the underlying [http.RoundTripper] to use.
	// If nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper
	// APIKey is the API key to add to requests as a "key" query parameter.
	APIKey string
}

// RoundTrip implements the [http.RoundTripper] interface. It adds the API key
// as a "key" query parameter if not already present, then delegates to the
// underlying transport.
func (t *ApiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())
	q := newReq.URL.Query()
	if q.Get("key") == "" && t.APIKey != "" {
		q.Set("key", t.APIKey)
		newReq.URL.RawQuery = q.Encode()
	}
	return transport.RoundTrip(newReq)
}
