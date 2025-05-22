package httprr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestingT is an interface that implements the minimal subset of testing.T methods we need.
type TestingT interface {
	Logf(format string, args ...interface{})
	Log(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// TestHelper provides utilities for using the recorder in tests.
type TestHelper struct {
	// T is the testing context.
	T TestingT

	// Recorder is the HTTP recorder.
	Recorder *Recorder

	// Client is a HTTP client that uses the Recorder.
	Client *http.Client

	// RecordingsDir is the directory where recordings are saved.
	RecordingsDir string
}

// NewTestHelper creates a new TestHelper with a configured Recorder and Client.
// By default, it creates a recorder in record mode with a temporary directory.
func NewTestHelper(t TestingT) *TestHelper {
	// Create a temporary directory for recordings
	recordingsDir, err := os.MkdirTemp("", "httprr-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Create the recorder
	recorder := NewRecorder(http.DefaultTransport)
	recorder.Dir = recordingsDir
	recorder.Mode = ModeRecord

	// Create the client
	client := &http.Client{
		Transport: recorder,
	}

	return &TestHelper{
		T:             t,
		Recorder:      recorder,
		Client:        client,
		RecordingsDir: recordingsDir,
	}
}

// NewReplayHelper creates a new TestHelper that replays HTTP interactions from a directory.
func NewReplayHelper(t TestingT, recordingsDir string) *TestHelper {
	// Create the recorder in replay mode
	recorder := NewRecorder(http.DefaultTransport)
	recorder.Dir = recordingsDir
	recorder.Mode = ModeReplay
	
	// Load existing recordings
	if err := recorder.loadRecordings(); err != nil {
		t.Logf("Warning: Failed to load recordings: %v", err)
	}

	// Create the client
	client := &http.Client{
		Transport: recorder,
	}

	return &TestHelper{
		T:             t,
		Recorder:      recorder,
		Client:        client,
		RecordingsDir: recordingsDir,
	}
}

// NewAutoHelper creates a new TestHelper that automatically detects whether to record or replay.
// It checks environment variables and the existence of the recordings directory to determine mode.
// - If HTTPRR_MODE is set to "record", it will record.
// - If HTTPRR_MODE is set to "replay", it will replay.
// - If recordingsDir exists and HTTPRR_MODE is not set, it will replay.
// - Otherwise, it will record.
func NewAutoHelper(t TestingT, recordingsDir string) *TestHelper {
	mode := os.Getenv("HTTPRR_MODE")
	
	// If mode is not set, check if recordings directory exists
	if mode == "" {
		if _, err := os.Stat(recordingsDir); err == nil {
			mode = "replay"
		} else {
			mode = "record"
		}
	}

	switch strings.ToLower(mode) {
	case "replay":
		return NewReplayHelper(t, recordingsDir)
	default:
		// Create a recorder in record mode
		recorder := NewRecorder(http.DefaultTransport)
		recorder.Dir = recordingsDir
		recorder.Mode = ModeRecord

		// Create the client
		client := &http.Client{
			Transport: recorder,
		}

		// Ensure the recordings directory exists
		if err := os.MkdirAll(recordingsDir, 0755); err != nil {
			t.Logf("Warning: Failed to create recordings directory: %v", err)
		}

		return &TestHelper{
			T:             t,
			Recorder:      recorder,
			Client:        client,
			RecordingsDir: recordingsDir,
		}
	}
}

// Cleanup removes temporary directories and other resources.
// This should not delete permanent recording directories.
func (h *TestHelper) Cleanup() {
	// Only clean up temporary directories, not permanent ones
	if h.RecordingsDir != "" && strings.Contains(h.RecordingsDir, "httprr-") {
		os.RemoveAll(h.RecordingsDir)
	}
}

// DumpRecordings prints all recorded HTTP interactions.
func (h *TestHelper) DumpRecordings() {
	records := h.Recorder.Records()
	if len(records) == 0 {
		h.T.Log("No HTTP interactions recorded")
		return
	}

	h.T.Logf("=== Recorded %d HTTP interactions ===", len(records))
	for i, record := range records {
		h.T.Logf("--- Interaction %d ---", i+1)
		h.T.Logf("Request: %s %s", record.Request.Method, record.Request.URL)
		
		// Print request headers and body
		reqDump := string(record.RequestDump)
		h.T.Logf("Request dump:\n%s", reqDump)
		
		if record.Error != nil {
			h.T.Logf("Error: %v", record.Error)
			continue
		}
		
		// Print response status, headers and body
		h.T.Logf("Response: %d %s", record.Response.StatusCode, record.Response.Status)
		respDump := string(record.ResponseDump)
		h.T.Logf("Response dump:\n%s", respDump)
	}
}

// AssertRequestCount ensures the expected number of requests were made.
func (h *TestHelper) AssertRequestCount(expected int) {
	actual := len(h.Recorder.Records())
	if actual != expected {
		h.T.Errorf("Expected %d HTTP requests, got %d", expected, actual)
	}
}

// GetRequestURLs returns a list of URLs that were requested.
func (h *TestHelper) GetRequestURLs() []string {
	var urls []string
	for _, record := range h.Recorder.Records() {
		urls = append(urls, record.Request.URL.String())
	}
	return urls
}

// AssertURLCalled checks if a specific URL (or part of a URL) was called.
func (h *TestHelper) AssertURLCalled(urlPart string) {
	for _, record := range h.Recorder.Records() {
		if strings.Contains(record.Request.URL.String(), urlPart) {
			return
		}
	}
	h.T.Errorf("URL containing %q was not called", urlPart)
}

// SaveRecordingsToDir saves all recordings to a specific directory.
func (h *TestHelper) SaveRecordingsToDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for i, record := range h.Recorder.Records() {
		// Create a filename
		method := record.Request.Method
		if method == "" {
			method = "UNKNOWN"
		}
		filename := fmt.Sprintf("%s-%d.txt", method, i)
		path := filepath.Join(dir, filename)

		// Open the file
		f, err := os.Create(path)
		if err != nil {
			return err
		}

		// Write the request
		if _, err := fmt.Fprintf(f, "REQUEST: %s %s\n", record.Request.Method, record.Request.URL); err != nil {
			f.Close()
			return err
		}
		if _, err := f.Write(record.RequestDump); err != nil {
			f.Close()
			return err
		}
		if _, err := fmt.Fprintf(f, "\n\n"); err != nil {
			f.Close()
			return err
		}

		// Write the response
		if record.Error != nil {
			if _, err := fmt.Fprintf(f, "ERROR: %v\n", record.Error); err != nil {
				f.Close()
				return err
			}
		} else {
			if _, err := fmt.Fprintf(f, "RESPONSE: %d %s\n", record.Response.StatusCode, record.Response.Status); err != nil {
				f.Close()
				return err
			}
			if _, err := f.Write(record.ResponseDump); err != nil {
				f.Close()
				return err
			}
		}

		f.Close()
	}

	return nil
}

// FindResponse finds a recorded response that matches the given URL pattern.
func (h *TestHelper) FindResponse(urlPattern string) (*http.Response, []byte, error) {
	for _, record := range h.Recorder.Records() {
		if strings.Contains(record.Request.URL.String(), urlPattern) {
			// If we found a match, make a deep copy of the response body
			var respBody []byte
			if record.Response.Body != nil {
				respBody, _ = io.ReadAll(record.Response.Body)
				record.Response.Body = io.NopCloser(bytes.NewBuffer(respBody))
			}
			return record.Response, respBody, nil
		}
	}
	return nil, nil, fmt.Errorf("no response found matching URL pattern: %s", urlPattern)
} 