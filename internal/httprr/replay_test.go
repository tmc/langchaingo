package httprr

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordAndReplay(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create a temp dir for recordings
	recordingsDir, err := os.MkdirTemp("", "httprr-replay-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(recordingsDir)

	// Step 1: Record the interaction
	{
		// Create a recorder
		recorder := NewRecorder(http.DefaultTransport)
		recorder.Dir = recordingsDir
		recorder.Mode = ModeRecord

		// Create a client
		client := &http.Client{Transport: recorder}

		// Make a request
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		// Verify the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "test response", string(body))

		// Verify that we recorded the interaction
		assert.Equal(t, 1, len(recorder.Records()))
	}

	// Step 2: Replay the interaction
	{
		// Create a new recorder in replay mode
		recorder := NewRecorder(http.DefaultTransport)
		recorder.Dir = recordingsDir
		recorder.Mode = ModeReplay

		// Load the recordings
		err := recorder.loadRecordings()
		require.NoError(t, err)

		// Create a client
		client := &http.Client{Transport: recorder}

		// Server is still running, but we should not hit it
		// Instead, we should get the recorded response
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		// Verify we got the recorded response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "test response", string(body))
	}
}

func TestReplayClient(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create a temp dir for recordings
	recordingsDir, err := os.MkdirTemp("", "httprr-replay-client-*")
	require.NoError(t, err)
	defer os.RemoveAll(recordingsDir)

	// Step 1: Record the interaction
	{
		recorder := NewRecorder(http.DefaultTransport)
		recorder.Dir = recordingsDir
		recorder.Mode = ModeRecord

		client := &http.Client{Transport: recorder}

		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	// Step 2: Use ReplayClient
	{
		client := ReplayClient(recordingsDir, http.DefaultTransport)

		// Server is still running, but we should not hit it
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		// Verify we got the recorded response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "test response", string(body))
	}
}

func TestAutoHelperRecordMode(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("auto helper test"))
	}))
	defer server.Close()

	// Create a temp dir path that doesn't exist yet
	recordingsDir := filepath.Join(os.TempDir(), "httprr-auto-test-"+t.Name())
	defer os.RemoveAll(recordingsDir)

	// Set environment variable to force record mode
	oldMode := os.Getenv("HTTPRR_MODE")
	defer os.Setenv("HTTPRR_MODE", oldMode)
	os.Setenv("HTTPRR_MODE", "record")

	// Create the helper
	helper := NewAutoHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Verify it's in record mode
	assert.Equal(t, ModeRecord, helper.Recorder.Mode)

	// Make a request
	resp, err := helper.Client.Get(server.URL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "auto helper test", string(body))

	// Verify that we recorded the interaction
	assert.Equal(t, 1, len(helper.Recorder.Records()))
}

func TestAutoHelperReplayMode(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("auto helper test"))
	}))
	defer server.Close()

	// Create a temp dir for recordings
	recordingsDir, err := os.MkdirTemp("", "httprr-auto-replay-*")
	require.NoError(t, err)
	defer os.RemoveAll(recordingsDir)

	// Step 1: Record the interaction
	{
		recorder := NewRecorder(http.DefaultTransport)
		recorder.Dir = recordingsDir
		recorder.Mode = ModeRecord

		client := &http.Client{Transport: recorder}

		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	// Step 2: Use AutoHelper with existing recordings dir
	{
		// Set environment variable to force replay mode
		oldMode := os.Getenv("HTTPRR_MODE")
		defer os.Setenv("HTTPRR_MODE", oldMode)
		os.Setenv("HTTPRR_MODE", "replay")

		helper := NewAutoHelper(t, recordingsDir)
		defer helper.Cleanup()

		// Verify it's in replay mode
		assert.Equal(t, ModeReplay, helper.Recorder.Mode)

		// Make a request
		resp, err := helper.Client.Get(server.URL)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()

		// Verify we got the recorded response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "auto helper test", string(body))
	}
} 