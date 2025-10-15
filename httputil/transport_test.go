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

			resp, err := transport.RoundTrip(req)
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
