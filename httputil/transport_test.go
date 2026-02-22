package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRoundTripper is a test helper that captures requests
type mockRoundTripper struct {
	lastRequest *http.Request
	response    *http.Response
	err         error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.lastRequest = req
	if m.response != nil {
		return m.response, m.err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}, m.err
}

func TestTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		existingUA     string
		expectedUAFunc func(string) bool
	}{
		{
			name:       "adds User-Agent when none exists",
			existingUA: "",
			expectedUAFunc: func(ua string) bool {
				return ua == UserAgent()
			},
		},
		{
			name:       "appends to existing User-Agent",
			existingUA: "MyApp/1.0",
			expectedUAFunc: func(ua string) bool {
				return ua == "MyApp/1.0 "+UserAgent()
			},
		},
		{
			name:       "appends to complex existing User-Agent",
			existingUA: "Mozilla/5.0 (compatible; MyBot/1.0)",
			expectedUAFunc: func(ua string) bool {
				return ua == "Mozilla/5.0 (compatible; MyBot/1.0) "+UserAgent()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRoundTripper{}
			transport := &Transport{Transport: mock}

			req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
			require.NoError(t, err)

			if tt.existingUA != "" {
				req.Header.Set("User-Agent", tt.existingUA)
			}

			resp, err := transport.RoundTrip(req) //nolint:bodyclose
			require.NoError(t, err)
			assert.NotNil(t, resp)

			// Check that the User-Agent was set correctly
			assert.True(t, tt.expectedUAFunc(mock.lastRequest.Header.Get("User-Agent")))

			// Verify original request wasn't modified
			if tt.existingUA != "" {
				assert.Equal(t, tt.existingUA, req.Header.Get("User-Agent"))
			} else {
				assert.Empty(t, req.Header.Get("User-Agent"))
			}
		})
	}
}

func TestTransport_NilTransport(t *testing.T) {
	// Create a test server to verify the request reaches it
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header is present
		ua := r.Header.Get("User-Agent")
		assert.Contains(t, ua, "langchaingo/")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &Transport{Transport: nil} // Should use http.DefaultTransport
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDefaultTransport(t *testing.T) {
	assert.NotNil(t, DefaultTransport)

	// Verify it's a Transport type
	transport, ok := DefaultTransport.(*Transport)
	assert.True(t, ok)
	assert.NotNil(t, transport.Transport)
	assert.Equal(t, http.DefaultTransport, transport.Transport)
}

func TestDefaultClient(t *testing.T) {
	assert.NotNil(t, DefaultClient)
	assert.Equal(t, DefaultTransport, DefaultClient.Transport)

	// Test that DefaultClient adds User-Agent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		assert.Contains(t, ua, "langchaingo/")
		assert.Contains(t, ua, "Go/")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := DefaultClient.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestApiKeyTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		existingKey    string
		transportKey   string
		expectedKey    string
		expectKeyAdded bool
	}{
		{
			name:           "adds API key when none exists",
			existingKey:    "",
			transportKey:   "test-key-123",
			expectedKey:    "test-key-123",
			expectKeyAdded: true,
		},
		{
			name:           "preserves existing API key",
			existingKey:    "existing-key",
			transportKey:   "transport-key",
			expectedKey:    "existing-key",
			expectKeyAdded: false,
		},
		{
			name:           "no key added when transport key is empty",
			existingKey:    "",
			transportKey:   "",
			expectedKey:    "",
			expectKeyAdded: false,
		},
		{
			name:           "empty transport key doesn't override existing",
			existingKey:    "existing-key",
			transportKey:   "",
			expectedKey:    "existing-key",
			expectKeyAdded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRoundTripper{}
			transport := &ApiKeyTransport{
				Transport: mock,
				APIKey:    tt.transportKey,
			}

			req, err := http.NewRequest(http.MethodGet, "https://api.example.com/data", nil)
			require.NoError(t, err)

			// Set existing API key if specified
			if tt.existingKey != "" {
				q := req.URL.Query()
				q.Set("key", tt.existingKey)
				req.URL.RawQuery = q.Encode()
			}

			resp, err := transport.RoundTrip(req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			defer resp.Body.Close()

			// Check the API key in the processed request
			actualKey := mock.lastRequest.URL.Query().Get("key")
			assert.Equal(t, tt.expectedKey, actualKey)

			// Verify original request wasn't modified
			originalKey := req.URL.Query().Get("key")
			assert.Equal(t, tt.existingKey, originalKey)
		})
	}
}

func TestApiKeyTransport_NilTransport(t *testing.T) {
	// Create a test server to verify the request reaches it with API key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		assert.Equal(t, "test-api-key", key)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &ApiKeyTransport{
		Transport: nil, // Should use http.DefaultTransport
		APIKey:    "test-api-key",
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestApiKeyTransport_PreservesOtherParams(t *testing.T) {
	mock := &mockRoundTripper{}
	transport := &ApiKeyTransport{
		Transport: mock,
		APIKey:    "my-api-key",
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/data?foo=bar&baz=qux", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()

	// Check that all query parameters are preserved
	query := mock.lastRequest.URL.Query()
	assert.Equal(t, "bar", query.Get("foo"))
	assert.Equal(t, "qux", query.Get("baz"))
	assert.Equal(t, "my-api-key", query.Get("key"))

	// Verify original request wasn't modified
	originalQuery := req.URL.Query()
	assert.Equal(t, "bar", originalQuery.Get("foo"))
	assert.Equal(t, "qux", originalQuery.Get("baz"))
	assert.Empty(t, originalQuery.Get("key"))
}

func TestApiKeyTransport_BaseURL(t *testing.T) { //nolint:funlen
	tests := []struct {
		name           string
		baseURL        string
		originalURL    string
		apiKey         string
		expectedScheme string
		expectedHost   string
		expectedPath   string
		expectedKey    string
	}{
		{
			name:           "rewrites URL with custom base URL",
			baseURL:        "http://localhost:8080",
			originalURL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini:generateContent",
			apiKey:         "test-key",
			expectedScheme: "http",
			expectedHost:   "localhost:8080",
			expectedPath:   "/v1beta/models/gemini:generateContent",
			expectedKey:    "test-key",
		},
		{
			name:           "rewrites URL preserving query parameters",
			baseURL:        "http://localhost:9000",
			originalURL:    "https://api.example.com/data?foo=bar&baz=qux",
			apiKey:         "my-api-key",
			expectedScheme: "http",
			expectedHost:   "localhost:9000",
			expectedPath:   "/data",
			expectedKey:    "my-api-key",
		},
		{
			name:           "rewrites URL without API key",
			baseURL:        "http://custom-server.local",
			originalURL:    "https://api.example.com/v1/resource",
			apiKey:         "",
			expectedScheme: "http",
			expectedHost:   "custom-server.local",
			expectedPath:   "/v1/resource",
			expectedKey:    "",
		},
		{
			name:           "no rewrite when BaseURL is empty",
			baseURL:        "",
			originalURL:    "https://api.example.com/data",
			apiKey:         "key-123",
			expectedScheme: "https",
			expectedHost:   "api.example.com",
			expectedPath:   "/data",
			expectedKey:    "key-123",
		},
		{
			name:           "BaseURL with path - combines paths correctly",
			baseURL:        "http://localhost:8080/litellm/v1",
			originalURL:    "https://generativelanguage.googleapis.com/models/gemini",
			apiKey:         "test-key",
			expectedScheme: "http",
			expectedHost:   "localhost:8080",
			expectedPath:   "/litellm/v1/models/gemini",
			expectedKey:    "test-key",
		},
		{
			name:           "BaseURL with path and trailing slash",
			baseURL:        "http://localhost:8080/api/v2/",
			originalURL:    "https://api.example.com/resources/item",
			apiKey:         "key-123",
			expectedScheme: "http",
			expectedHost:   "localhost:8080",
			expectedPath:   "/api/v2/resources/item",
			expectedKey:    "key-123",
		},
		{
			name:           "BaseURL with deep path",
			baseURL:        "http://proxy.local/gateway/llm/google",
			originalURL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini:generateContent",
			apiKey:         "secret",
			expectedScheme: "http",
			expectedHost:   "proxy.local",
			expectedPath:   "/gateway/llm/google/v1beta/models/gemini:generateContent",
			expectedKey:    "secret",
		},
		{
			name:           "original URL with only root path",
			baseURL:        "http://localhost:8080/prefix",
			originalURL:    "https://api.example.com/",
			apiKey:         "key",
			expectedScheme: "http",
			expectedHost:   "localhost:8080",
			expectedPath:   "/prefix",
			expectedKey:    "key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRoundTripper{}
			transport := &ApiKeyTransport{
				Transport: mock,
				APIKey:    tt.apiKey,
				BaseURL:   tt.baseURL,
			}

			req, err := http.NewRequest(http.MethodPost, tt.originalURL, nil)
			require.NoError(t, err)

			resp, err := transport.RoundTrip(req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			defer resp.Body.Close()

			// Check the rewritten URL
			assert.Equal(t, tt.expectedScheme, mock.lastRequest.URL.Scheme)
			assert.Equal(t, tt.expectedHost, mock.lastRequest.URL.Host)
			assert.Equal(t, tt.expectedPath, mock.lastRequest.URL.Path)

			// Check API key
			if tt.expectedKey != "" {
				assert.Equal(t, tt.expectedKey, mock.lastRequest.URL.Query().Get("key"))
			} else {
				assert.Empty(t, mock.lastRequest.URL.Query().Get("key"))
			}

			// Verify Host header matches the new URL
			if tt.baseURL != "" {
				assert.Equal(t, tt.expectedHost, mock.lastRequest.Host)
			}

			// Verify original request wasn't modified
			assert.Equal(t, tt.originalURL, req.URL.String())
		})
	}
}

func TestApiKeyTransport_BaseURL_PreservesQueryParams(t *testing.T) {
	mock := &mockRoundTripper{}
	transport := &ApiKeyTransport{
		Transport: mock,
		APIKey:    "my-key",
		BaseURL:   "http://localhost:8080",
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/data?foo=bar&baz=qux", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()

	// Check that URL was rewritten
	assert.Equal(t, "http", mock.lastRequest.URL.Scheme)
	assert.Equal(t, "localhost:8080", mock.lastRequest.URL.Host)
	assert.Equal(t, "/data", mock.lastRequest.URL.Path)

	// Check that original query parameters are preserved
	query := mock.lastRequest.URL.Query()
	assert.Equal(t, "bar", query.Get("foo"))
	assert.Equal(t, "qux", query.Get("baz"))
	assert.Equal(t, "my-key", query.Get("key"))
}

func TestApiKeyTransport_BaseURL_InvalidURL(t *testing.T) {
	mock := &mockRoundTripper{}
	transport := &ApiKeyTransport{
		Transport: mock,
		APIKey:    "test-key",
		BaseURL:   "ht!tp://invalid url with spaces",
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/data", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req) //nolint:bodyclose
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse BaseURL")
}

func TestApiKeyTransport_BaseURL_PathTraversalProtection(t *testing.T) {
	// These tests verify that path traversal attacks are prevented
	// The path.Clean function normalizes paths and removes ".." segments
	// The base path is always preserved, and the original path is appended to it
	tests := []struct {
		name         string
		baseURL      string
		originalPath string
		expectedPath string
		description  string
	}{
		{
			name:         "path traversal attempt with ../ is neutralized",
			baseURL:      "http://localhost:8080/api/v1",
			originalPath: "https://evil.com/../../../etc/passwd",
			expectedPath: "/api/v1/etc/passwd", // path.Clean removes ../.. , then we append to base
			description:  "Path traversal is cleaned and appended to base path safely",
		},
		{
			name:         "multiple slashes are normalized",
			baseURL:      "http://localhost:8080/api//v1///",
			originalPath: "https://api.example.com///models////gemini",
			expectedPath: "/api/v1/models/gemini",
			description:  "Multiple consecutive slashes are collapsed",
		},
		{
			name:         "dot segments are removed",
			baseURL:      "http://localhost:8080/./api/./v1/.",
			originalPath: "https://api.example.com/./models/./gemini/.",
			expectedPath: "/api/v1/models/gemini",
			description:  "Single dot segments are removed",
		},
		{
			name:         "complex path with mixed traversal attempts",
			baseURL:      "http://localhost:8080/base/path",
			originalPath: "https://api.example.com/foo/../bar/./baz/../qux",
			expectedPath: "/base/path/bar/qux",
			description:  "Complex path is properly normalized",
		},
		{
			name:         "empty path segments",
			baseURL:      "http://localhost:8080/api/v1",
			originalPath: "https://api.example.com/",
			expectedPath: "/api/v1",
			description:  "Empty original path uses base path",
		},
		{
			name:         "traversal within original path only",
			baseURL:      "http://localhost:8080/api",
			originalPath: "https://api.example.com/v1/models/../resources",
			expectedPath: "/api/v1/resources",
			description:  "Traversal within original path is resolved correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRoundTripper{}
			transport := &ApiKeyTransport{
				Transport: mock,
				BaseURL:   tt.baseURL,
				APIKey:    "test-key",
			}

			req, err := http.NewRequest(http.MethodGet, tt.originalPath, nil)
			require.NoError(t, err)

			resp, err := transport.RoundTrip(req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			defer resp.Body.Close()

			actualPath := mock.lastRequest.URL.Path
			assert.Equal(t, tt.expectedPath, actualPath, tt.description)

			// Verify the path doesn't contain any ".." or "." segments
			assert.NotContains(t, actualPath, "/..", "Path should not contain '..' segments")
			assert.NotContains(t, actualPath, "/./", "Path should not contain '.' segments")
			assert.NotContains(t, actualPath, "//", "Path should not contain double slashes")
		})
	}
}

func TestApiKeyTransport_BaseURL_PathCombinations(t *testing.T) {
	// Additional tests for various path combination scenarios
	tests := []struct {
		name         string
		baseURL      string
		originalURL  string
		expectedPath string
	}{
		{
			name:         "both paths with trailing and leading slashes",
			baseURL:      "http://localhost:8080/api/",
			originalURL:  "https://example.com/v1/resource",
			expectedPath: "/api/v1/resource",
		},
		{
			name:         "base path without trailing slash, original with leading slash",
			baseURL:      "http://localhost:8080/api",
			originalURL:  "https://example.com/v1/resource",
			expectedPath: "/api/v1/resource",
		},
		{
			name:         "base URL is just domain with trailing slash",
			baseURL:      "http://localhost:8080/",
			originalURL:  "https://example.com/v1/resource",
			expectedPath: "/v1/resource",
		},
		{
			name:         "base URL path with colon (like port or special chars)",
			baseURL:      "http://localhost:8080/api:v1",
			originalURL:  "https://example.com/models:generate",
			expectedPath: "/api:v1/models:generate",
		},
		{
			name:         "nested paths with special characters",
			baseURL:      "http://localhost:8080/api/v1.2/beta-test",
			originalURL:  "https://example.com/models/gemini-2.0-flash",
			expectedPath: "/api/v1.2/beta-test/models/gemini-2.0-flash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRoundTripper{}
			transport := &ApiKeyTransport{
				Transport: mock,
				BaseURL:   tt.baseURL,
			}

			req, err := http.NewRequest(http.MethodGet, tt.originalURL, nil)
			require.NoError(t, err)

			resp, err := transport.RoundTrip(req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedPath, mock.lastRequest.URL.Path)
		})
	}
}

func TestApiKeyTransport_ProxyURL(t *testing.T) { //nolint:funlen
	// Create a test HTTP server that will act as the final destination
	destinationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key is present
		assert.Equal(t, "test-api-key", r.URL.Query().Get("key"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("response from destination"))
	}))
	defer destinationServer.Close()

	// Create a simple HTTP proxy server
	proxyRequestCount := 0
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyRequestCount++

		// For CONNECT method (HTTPS tunneling), establish tunnel
		if r.Method == http.MethodConnect {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For regular HTTP requests, forward the request
		client := &http.Client{}
		proxyReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)

		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		_, _ = w.Write(body[:n])
	}))
	defer proxyServer.Close()

	t.Run("requests_go_through_proxy", func(t *testing.T) {
		proxyRequestCount = 0

		transport := &ApiKeyTransport{
			APIKey:   "test-api-key",
			ProxyURL: proxyServer.URL,
		}

		client := &http.Client{Transport: transport}
		resp, err := client.Get(destinationServer.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Greater(t, proxyRequestCount, 0, "Request should have gone through proxy")
	})

	t.Run("proxy_with_base_url_rewrite", func(t *testing.T) {
		// Create a mock server that will receive rewritten requests
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "test-api-key", r.URL.Query().Get("key"))
			assert.Equal(t, "/v1/test", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mock response"))
		}))
		defer mockServer.Close()

		transport := &ApiKeyTransport{
			APIKey:   "test-api-key",
			BaseURL:  mockServer.URL,
			ProxyURL: proxyServer.URL,
		}

		client := &http.Client{Transport: transport}

		// Original URL will be rewritten to mockServer.URL
		resp, err := client.Get("https://original-api.example.com/v1/test")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("proxy_ignored_with_custom_transport", func(t *testing.T) {
		customTransport := &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			},
		}

		transport := &ApiKeyTransport{
			Transport: customTransport,
			APIKey:    "test-key",
			ProxyURL:  proxyServer.URL, // Should be ignored
		}

		req, err := http.NewRequest(http.MethodGet, "https://api.example.com/data", nil)
		require.NoError(t, err)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		defer resp.Body.Close()

		// Verify custom transport was used (ProxyURL was ignored)
		assert.NotNil(t, customTransport.lastRequest)
	})
}

func TestApiKeyTransport_ProxyURL_InvalidURL(t *testing.T) {
	transport := &ApiKeyTransport{
		APIKey:   "test-key",
		ProxyURL: "ht!tp://invalid proxy url",
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/data", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req) //nolint:bodyclose
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse ProxyURL")
}

func TestApiKeyTransport_CombinedFeatures(t *testing.T) {
	// This test verifies that all features work together:
	// - ProxyURL
	// - BaseURL rewriting
	// - API key injection

	// Create destination server
	destinationCalled := false
	destinationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		destinationCalled = true

		// Verify all features worked
		assert.Equal(t, "test-api-key", r.URL.Query().Get("key"), "API key should be present")
		assert.Equal(t, "/api/v1/resource", r.URL.Path, "Path should be preserved")
		assert.Equal(t, "value", r.URL.Query().Get("param"), "Query param should be preserved")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer destinationServer.Close()

	// Create simple proxy (just forwards requests)
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Forward request
		client := &http.Client{}
		proxyReq, _ := http.NewRequest(r.Method, r.URL.String(), r.Body)
		for k, v := range r.Header {
			proxyReq.Header[k] = v
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)

		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		_, _ = w.Write(body[:n])
	}))
	defer proxyServer.Close()

	transport := &ApiKeyTransport{
		APIKey:   "test-api-key",
		BaseURL:  destinationServer.URL,
		ProxyURL: proxyServer.URL,
	}

	client := &http.Client{Transport: transport}

	// Original URL will be rewritten to destinationServer.URL
	resp, err := client.Get("https://original-api.example.com/api/v1/resource?param=value")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, destinationCalled, "Destination server should have been called")
}
