package httprr

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestHelper(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create a test helper
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Make a request using the helper's client
	resp, err := helper.Client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()

	// Test AssertRequestCount
	helper.AssertRequestCount(1)

	// Test GetRequestURLs
	urls := helper.GetRequestURLs()
	if len(urls) != 1 || urls[0] != server.URL {
		t.Errorf("GetRequestURLs() = %v, want [%s]", urls, server.URL)
	}

	// Test AssertURLCalled
	helper.AssertURLCalled(server.URL)

	// Test FindResponse
	foundResp, body, err := helper.FindResponse(server.URL)
	if err != nil {
		t.Errorf("FindResponse() error = %v", err)
	}
	if foundResp.StatusCode != http.StatusOK {
		t.Errorf("FindResponse().StatusCode = %v, want %v", foundResp.StatusCode, http.StatusOK)
	}
	if string(body) != "test response" {
		t.Errorf("FindResponse() body = %q, want %q", string(body), "test response")
	}
}

func TestTestHelperSaveRecordingsToDir(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create a test helper
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Make a request using the helper's client
	resp, err := helper.Client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()

	// Create a temporary directory to save recordings
	dir, err := os.MkdirTemp("", "httprr-save-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Save recordings to the directory
	err = helper.SaveRecordingsToDir(dir)
	if err != nil {
		t.Fatalf("SaveRecordingsToDir() error = %v", err)
	}

	// Check that a file was created
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	// Read the file
	content, err := os.ReadFile(filepath.Join(dir, files[0].Name()))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check file content
	if !strings.Contains(string(content), "REQUEST: GET") {
		t.Errorf("File doesn't contain request method")
	}
	if !strings.Contains(string(content), server.URL) {
		t.Errorf("File doesn't contain request URL")
	}
	if !strings.Contains(string(content), "RESPONSE: 200 OK") {
		t.Errorf("File doesn't contain response status")
	}
	if !strings.Contains(string(content), "test response") {
		t.Errorf("File doesn't contain response body")
	}
}

func TestTestHelperDumpRecordings(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// This is hard to test since it just logs output, 
	// but we can verify it doesn't crash
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Make a request
	resp, err := helper.Client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()

	// Test that DumpRecordings doesn't crash
	helper.DumpRecordings()

	// Reset and test that it handles no recordings
	helper.Recorder.Reset()
	helper.DumpRecordings()
}

func TestTestHelperAssertURLCalledFailure(t *testing.T) {
	// Create a test that should fail (but we're going to catch the failure)
	mockT := new(mockTestingT)
	helper := NewTestHelper(mockT)
	defer helper.Cleanup()

	// Assert a URL that wasn't called
	helper.AssertURLCalled("http://example.com")

	// Verify that the test failed
	assert.True(t, mockT.failed, "Expected AssertURLCalled to fail the test")
}

func TestTestHelperAssertRequestCountFailure(t *testing.T) {
	// Create a test that should fail (but we're going to catch the failure)
	mockT := new(mockTestingT)
	helper := NewTestHelper(mockT)
	defer helper.Cleanup()

	// Assert wrong request count
	helper.AssertRequestCount(1)

	// Verify that the test failed
	assert.True(t, mockT.failed, "Expected AssertRequestCount to fail the test")
}

// mockTestingT implements a subset of testing.T for testing assertions
type mockTestingT struct {
	failed bool
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.failed = true
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failed = true
}

func (m *mockTestingT) FailNow() {
	m.failed = true
}

func (m *mockTestingT) Logf(format string, args ...interface{}) {
	// Do nothing
}

func (m *mockTestingT) Log(args ...interface{}) {
	// Do nothing
} 