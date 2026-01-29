package httputil

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
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

// ApiKeyTransport is an [http.RoundTripper] that adds API keys to URL query parameters,
// optionally rewrites request URLs to use a custom base URL, and supports HTTP proxy configuration.
// This is commonly used with Google APIs and other services that accept API keys
// as query parameters. It wraps another RoundTripper and automatically adds
// the API key if not already present in the request.
//
// This transport is particularly useful when working with client libraries that
// don't properly set API keys when using custom HTTP clients, such as the
// Google AI client library when used with httprr for testing, or when you need
// to redirect requests to a custom endpoint (e.g., for testing or proxy setups).
//
// Proxy Configuration:
// If ProxyURL is set and Transport is nil or http.DefaultTransport, a new http.Transport
// with the specified proxy will be created automatically. If a custom Transport is already
// provided, ProxyURL will be ignored to avoid conflicting configurations.
type ApiKeyTransport struct {
	// Transport is the underlying [http.RoundTripper] to use.
	// If nil, [http.DefaultTransport] is used (or a custom transport with proxy if ProxyURL is set).
	Transport http.RoundTripper
	// APIKey is the API key to add to requests as a "key" query parameter.
	APIKey string
	// BaseURL is the base URL to use for rewriting request URLs.
	// If set, requests will be rewritten to use this base URL instead of their original URL.
	// Both the base URL path and the original request path are combined safely, and query
	// parameters are preserved.
	//
	// Examples:
	//   - BaseURL: "http://localhost:8080"
	//     Request: "https://api.example.com/v1/resource?param=value"
	//     Result:  "http://localhost:8080/v1/resource?param=value"
	//
	//   - BaseURL: "http://localhost:8080/litellm/v1"
	//     Request: "https://api.example.com/models/gemini?key=abc"
	//     Result:  "http://localhost:8080/litellm/v1/models/gemini?key=abc"
	//
	// The path combination is done safely to prevent path traversal attacks.
	BaseURL string
	// ProxyURL is the HTTP proxy URL to use for requests.
	// If set and Transport is nil or http.DefaultTransport, a new http.Transport with
	// this proxy will be created. Example: "http://proxy.example.com:8080".
	// If a custom Transport is already set, this field is ignored.
	ProxyURL string
}

// RoundTrip implements the [http.RoundTripper] interface. It optionally configures
// an HTTP proxy, rewrites the request URL to use a custom base URL, adds the API key
// as a "key" query parameter if not already present, then delegates to the underlying transport.
func (t *ApiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport

	// Configure proxy if ProxyURL is set and no custom transport is provided
	if t.ProxyURL != "" && (transport == nil || transport == http.DefaultTransport) {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ProxyURL: %w", err)
		}

		// Create a new http.Transport with proxy configuration
		// We cast http.DefaultTransport to get default settings, then override Proxy
		defaultTransport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			// Fallback if DefaultTransport is not *http.Transport
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		} else {
			// Clone the default transport and set proxy
			transport = defaultTransport.Clone()
			transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
		}
	} else if transport == nil {
		transport = http.DefaultTransport
	}

	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// Preserve original query parameters before URL manipulation
	q := newReq.URL.Query()

	// Rewrite URL to use custom base URL if configured
	if t.BaseURL != "" {
		baseURL, err := url.Parse(t.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse BaseURL: %w", err)
		}

		// Safely combine base URL path with original request path
		// This preserves any path from BaseURL (e.g., "/litellm/v1") and appends the original path
		basePath := baseURL.Path
		originalPath := newReq.URL.Path

		// Clean paths to prevent path traversal attacks and normalize them
		basePath = path.Clean(basePath)
		originalPath = path.Clean(originalPath)

		// Combine paths using path.Join which handles all edge cases safely
		// path.Join automatically:
		// - Removes duplicate slashes
		// - Normalizes "." and ".." segments
		// - Handles trailing slashes properly
		var combinedPath string
		if originalPath == "" || originalPath == "." || originalPath == "/" {
			// If original path is empty, ".", or just "/", use base path only
			combinedPath = basePath
		} else {
			// Join base path with original path
			// We need to ensure originalPath doesn't start with "/" for proper joining
			originalPath = strings.TrimPrefix(originalPath, "/")
			combinedPath = path.Join(basePath, originalPath)
		}

		// Ensure the path starts with "/" (path.Join might remove it for root)
		if combinedPath == "" || combinedPath[0] != '/' {
			combinedPath = "/" + combinedPath
		}

		// Create new URL with combined path
		newReq.URL = &url.URL{
			Scheme: baseURL.Scheme,
			Host:   baseURL.Host,
			Path:   combinedPath,
		}

		// Restore query parameters
		newReq.URL.RawQuery = q.Encode()

		// Update Host header to match the new URL
		newReq.Host = newReq.URL.Host
	}

	// Add API key as query parameter if not already present
	if q.Get("key") == "" && t.APIKey != "" {
		q.Set("key", t.APIKey)
		newReq.URL.RawQuery = q.Encode()
	}

	return transport.RoundTrip(newReq)
}
