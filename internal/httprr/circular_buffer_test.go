package httprr

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCircularBuffer tests that identical requests get different responses
// in the order they were recorded, cycling back to the beginning
func TestCircularBuffer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := dir + "/circular.httprr"

	// Create a test server that returns different responses for the same request
	responseCounter := 0
	responses := []string{
		"Response 1",
		"Response 2",
		"Response 3",
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responses[responseCounter]))
		responseCounter++
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Phase 1: Record mode - make 3 identical requests
	t.Run("record", func(t *testing.T) {
		rr, err := create(file, http.DefaultTransport)
		require.NoError(t, err)
		defer rr.Close()

		// Make 3 identical requests
		for i := 0; i < 3; i++ {
			req, err := http.NewRequest("GET", srv.URL+"/test", nil)
			require.NoError(t, err)

			resp, err := rr.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			require.NoError(t, err)

			expected := responses[i]
			assert.Equal(t, expected, string(body), "Recording request %d should get response %d", i+1, i+1)
		}
	})

	// Phase 2: Replay mode - verify responses are returned in order
	t.Run("replay-first-cycle", func(t *testing.T) {
		rr, err := open(file, http.DefaultTransport)
		require.NoError(t, err)

		// Request the same endpoint 3 times, should get responses in order
		for i := 0; i < 3; i++ {
			req, err := http.NewRequest("GET", srv.URL+"/test", nil)
			require.NoError(t, err)

			resp, err := rr.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			require.NoError(t, err)

			expected := responses[i]
			assert.Equal(t, expected, string(body), "First cycle request %d should get response %d", i+1, i+1)
		}
	})

	// Phase 3: Verify circular behavior - should cycle back to first response
	t.Run("replay-second-cycle", func(t *testing.T) {
		rr, err := open(file, http.DefaultTransport)
		require.NoError(t, err)

		// Make 6 more requests to test the circular buffer (2 full cycles)
		for cycle := 0; cycle < 2; cycle++ {
			for i := 0; i < 3; i++ {
				req, err := http.NewRequest("GET", srv.URL+"/test", nil)
				require.NoError(t, err)

				resp, err := rr.Client().Do(req)
				require.NoError(t, err)

				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				require.NoError(t, err)

				expected := responses[i]
				assert.Equal(t, expected, string(body), "Cycle %d request %d should get response %d", cycle+1, i+1, i+1)
			}
		}
	})
}

// TestCircularBufferWithDifferentResponses tests that different requests
// don't interfere with each other's response buffers
func TestCircularBufferWithDifferentRequests(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := dir + "/different.httprr"

	// Create a test server that responds differently based on the URL path
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		switch r.URL.Path {
		case "/endpoint1":
			w.Write([]byte("Endpoint 1"))
		case "/endpoint2":
			w.Write([]byte("Endpoint 2"))
		default:
			w.Write([]byte("Unknown"))
		}
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Phase 1: Record mode
	t.Run("record", func(t *testing.T) {
		rr, err := create(file, http.DefaultTransport)
		require.NoError(t, err)
		defer rr.Close()

		// Record endpoint1 twice
		for i := 0; i < 2; i++ {
			req, err := http.NewRequest("GET", srv.URL+"/endpoint1", nil)
			require.NoError(t, err)
			resp, err := rr.Client().Do(req)
			require.NoError(t, err)
			resp.Body.Close()
		}

		// Record endpoint2 once
		req, err := http.NewRequest("GET", srv.URL+"/endpoint2", nil)
		require.NoError(t, err)
		resp, err := rr.Client().Do(req)
		require.NoError(t, err)
		resp.Body.Close()
	})

	// Phase 2: Replay mode - verify different requests maintain separate buffers
	t.Run("replay", func(t *testing.T) {
		rr, err := open(file, http.DefaultTransport)
		require.NoError(t, err)

		// Request endpoint1 twice
		for i := 0; i < 2; i++ {
			req, err := http.NewRequest("GET", srv.URL+"/endpoint1", nil)
			require.NoError(t, err)
			resp, err := rr.Client().Do(req)
			require.NoError(t, err)
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, "Endpoint 1", string(body))
		}

		// Request endpoint2 once
		req, err := http.NewRequest("GET", srv.URL+"/endpoint2", nil)
		require.NoError(t, err)
		resp, err := rr.Client().Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, "Endpoint 2", string(body))

		// Request endpoint1 again - should cycle back to first response
		req, err = http.NewRequest("GET", srv.URL+"/endpoint1", nil)
		require.NoError(t, err)
		resp, err = rr.Client().Do(req)
		require.NoError(t, err)
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
		assert.Equal(t, "Endpoint 1", string(body))
	})
}

// TestCircularBufferWithBody tests circular buffer with POST requests containing bodies
func TestCircularBufferWithBody(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := dir + "/withbody.httprr"

	// Create a test server that echoes request number
	counter := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		counter++
		w.Write([]byte("Response " + string(rune('0'+counter))))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Phase 1: Record mode - make identical POST requests
	t.Run("record", func(t *testing.T) {
		rr, err := create(file, http.DefaultTransport)
		require.NoError(t, err)
		defer rr.Close()

		// Make 3 identical POST requests with the same body
		for i := 0; i < 3; i++ {
			req, err := http.NewRequest("POST", srv.URL+"/api", strings.NewReader(`{"key":"value"}`))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := rr.Client().Do(req)
			require.NoError(t, err)
			resp.Body.Close()
		}
	})

	// Phase 2: Replay mode - verify responses cycle
	t.Run("replay", func(t *testing.T) {
		rr, err := open(file, http.DefaultTransport)
		require.NoError(t, err)

		expectedResponses := []string{"Response 1", "Response 2", "Response 3"}

		// Make 6 requests to verify cycling (2 full cycles)
		for i := 0; i < 6; i++ {
			req, err := http.NewRequest("POST", srv.URL+"/api", strings.NewReader(`{"key":"value"}`))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := rr.Client().Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			require.NoError(t, err)

			expectedIdx := i % 3
			assert.Equal(t, expectedResponses[expectedIdx], string(body),
				"Request %d should get response %d (cycling)", i+1, expectedIdx+1)
		}
	})
}
