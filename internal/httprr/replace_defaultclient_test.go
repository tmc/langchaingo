package httprr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAndRestoreDefaultClient(t *testing.T) {
	// Save the original client to restore after the test
	originalClient := http.DefaultClient

	// Replace the default client
	recorder := ReplaceDefaultClient()
	assert.NotNil(t, recorder, "Recorder should not be nil")
	
	// Check that http.DefaultClient has changed
	assert.NotEqual(t, originalClient, http.DefaultClient, "http.DefaultClient should be replaced")
	
	// Check that the transport is our recorder
	transport, ok := http.DefaultClient.Transport.(*Recorder)
	assert.True(t, ok, "http.DefaultClient.Transport should be a *Recorder")
	assert.Equal(t, recorder, transport, "The recorder returned should be the one used in DefaultClient")
	
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()
	
	// Use the default client to make a request
	resp, err := http.Get(server.URL)
	assert.NoError(t, err, "Should successfully make a request")
	if resp != nil {
		resp.Body.Close()
	}
	
	// Check that the recorder recorded the request
	assert.Equal(t, 1, len(recorder.Records()), "Request should be recorded")
	assert.Equal(t, server.URL, recorder.Records()[0].Request.URL.String(), "Recorded URL should match")
	
	// Restore the default client
	RestoreDefaultClient()
	
	// Verify the original client was restored
	assert.Equal(t, originalClient, http.DefaultClient, "http.DefaultClient should be restored")
}

func TestCannotReplaceClientTwice(t *testing.T) {
	// Save the original client to restore after the test
	originalClient := http.DefaultClient
	defer func() {
		http.DefaultClient = originalClient
		replaced = false // Reset the replaced flag
	}()
	
	// Replace the client once
	_ = ReplaceDefaultClient()
	
	// Try to replace it again, should panic
	assert.Panics(t, func() {
		ReplaceDefaultClient()
	}, "Replacing client twice should panic")
}

func TestRestoreWithoutReplace(t *testing.T) {
	// Save the original client to restore after the test
	originalClient := http.DefaultClient
	
	// Make sure the client is not already replaced
	replaced = false
	
	// Restore the client without replacing it first, should not panic
	assert.NotPanics(t, func() {
		RestoreDefaultClient()
	}, "Restoring without replacing should not panic")
	
	// Verify the client didn't change
	assert.Equal(t, originalClient, http.DefaultClient, "http.DefaultClient should not change")
} 