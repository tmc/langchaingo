// Package httputil provides HTTP utilities for LangChainGo.
package httputil

import (
	"net/http"
)

var (
	// DefaultTransport is the default HTTP transport for LangChainGo.
	// It adds a User-Agent header with version information.
	DefaultTransport http.RoundTripper = &Transport{
		Transport: http.DefaultTransport,
	}

	// DefaultClient is the default HTTP client for LangChainGo.
	// It includes a User-Agent header with version information.
	DefaultClient = &http.Client{
		Transport: DefaultTransport,
	}
)

// Transport is a custom [http.RoundTripper] that adds a User-Agent header.
type Transport struct {
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction.
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
