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
