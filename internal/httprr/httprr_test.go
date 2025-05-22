package httprr

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecording(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "test response", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
	}))
	defer server.Close()

	// Create temporary directory for recordings
	tempDir, err := os.MkdirTemp("", "httprr_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test recording mode
	t.Run("Record", func(t *testing.T) {
		recorder := New(tempDir, ModeRecord)
		client := recorder.Client()

		resp, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if !strings.Contains(string(body), "test response") {
			t.Errorf("Expected response to contain 'test response', got: %s", string(body))
		}

		// Check if recording was saved
		files, err := filepath.Glob(filepath.Join(tempDir, "*.json"))
		if err != nil {
			t.Fatalf("Failed to list recordings: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 recording file, got %d", len(files))
		}
	})

	// Test replay mode
	t.Run("Replay", func(t *testing.T) {
		recorder := New(tempDir, ModeReplay)
		client := recorder.Client()

		resp, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("Replay request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if !strings.Contains(string(body), "test response") {
			t.Errorf("Expected response to contain 'test response', got: %s", string(body))
		}
	})

	// Test record once mode
	t.Run("RecordOnce", func(t *testing.T) {
		recorder := New(tempDir, ModeRecordOnce)
		client := recorder.Client()

		// First request should use existing recording
		resp1, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("RecordOnce request failed: %v", err)
		}
		resp1.Body.Close()

		// Second request should also use existing recording
		resp2, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("RecordOnce second request failed: %v", err)
		}
		resp2.Body.Close()

		// Should still have only one recording file
		files, err := filepath.Glob(filepath.Join(tempDir, "*.json"))
		if err != nil {
			t.Fatalf("Failed to list recordings: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 recording file, got %d", len(files))
		}
	})
}

func TestIDGeneration(t *testing.T) {
	recorder := New("", ModeRecord)

	// Create test requests
	req1, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req2, _ := http.NewRequest("GET", "http://example.com/test", nil)
	req3, _ := http.NewRequest("POST", "http://example.com/test", nil)

	id1 := recorder.generateID(req1)
	id2 := recorder.generateID(req2)
	id3 := recorder.generateID(req3)

	// Same requests should generate same ID
	if id1 != id2 {
		t.Errorf("Expected same ID for identical requests, got %s and %s", id1, id2)
	}

	// Different requests should generate different IDs
	if id1 == id3 {
		t.Errorf("Expected different IDs for different requests, got %s", id1)
	}
}

func TestVariableHeaders(t *testing.T) {
	testCases := []struct {
		header   string
		variable bool
	}{
		{"Authorization", true},
		{"User-Agent", true},
		{"Content-Type", false},
		{"Accept", false},
		{"X-Request-Id", true},
	}

	for _, tc := range testCases {
		result := isVariableHeader(tc.header)
		if result != tc.variable {
			t.Errorf("Header %s: expected %v, got %v", tc.header, tc.variable, result)
		}
	}
}

func TestSanitizeBody(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{`{"message": "hello"}`, `{"message": "hello"}`},
		{`{"timestamp": "2023-01-01T00:00:00Z"}`, `{"timestamp": "REDACTED"}`},
		{`{"id": "123", "data": "test"}`, `{"id": "REDACTED", "data": "test"}`},
	}

	for _, tc := range testCases {
		result := sanitizeBody(tc.input)
		if result != tc.expected {
			t.Errorf("Input %s: expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}

func TestReset(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "httprr_reset_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	recorder := New(tempDir, ModeRecord)

	// Add a mock recording
	recorder.recordings["test"] = &Recording{
		ID:      "test",
		Created: time.Now(),
	}

	// Reset without removing files
	err = recorder.Reset(false)
	if err != nil {
		t.Errorf("Reset failed: %v", err)
	}

	if len(recorder.recordings) != 0 {
		t.Errorf("Expected recordings to be cleared, got %d recordings", len(recorder.recordings))
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.json")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Reset with removing files
	err = recorder.Reset(true)
	if err != nil {
		t.Errorf("Reset with file removal failed: %v", err)
	}

	// Check if directory was removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("Expected directory to be removed")
	}
}