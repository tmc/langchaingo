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
