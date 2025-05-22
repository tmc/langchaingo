package httprr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecorder(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from test server",
		})
	}))
	defer server.Close()

	// Create a recorder
	recorder := NewRecorder(http.DefaultTransport)

	// Create a request
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Execute the request using the recorder
	resp, err := recorder.RoundTrip(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	// Check that we recorded the request
	records := recorder.Records()
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// Check the recorded request
	record := records[0]
	if record.Request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", record.Request.Method)
	}
	if record.Request.URL.String() != server.URL {
		t.Errorf("Expected URL %s, got %s", server.URL, record.Request.URL)
	}

	// Check the recorded response
	if record.Response.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", record.Response.Status)
	}
	if !bytes.Contains(record.ResponseDump, []byte("Hello from test server")) {
		t.Errorf("Response dump doesn't contain expected content")
	}
}

func TestRecorderWithRequestBody(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"request_body": string(body),
		})
	}))
	defer server.Close()

	// Create a recorder
	recorder := NewRecorder(http.DefaultTransport)

	// Create a request with a body
	requestBody := `{"test":"value"}`
	req, err := http.NewRequest("POST", server.URL, strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request using the recorder
	resp, err := recorder.RoundTrip(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Verify response can be read after recording
	var responseData map[string]string
	if err := json.Unmarshal(respBody, &responseData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if responseData["request_body"] != requestBody {
		t.Errorf("Expected %q, got %q", requestBody, responseData["request_body"])
	}

	// Check the recorded request
	records := recorder.Records()
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// Verify request dump contains the body
	if !bytes.Contains(records[0].RequestDump, []byte(requestBody)) {
		t.Errorf("Request dump doesn't contain expected body")
	}
}

func TestRecorderSavesToDisk(t *testing.T) {
	// Create temporary directory for recordings
	dir, err := os.MkdirTemp("", "httprr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create a recorder that saves to disk
	recorder := NewRecorder(http.DefaultTransport)
	recorder.Dir = dir
	recorder.Pattern = "test-%s-%d.txt"

	// Create a request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Execute the request using the recorder
	resp, err := recorder.RoundTrip(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check that a file was created in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	// Read the file
	fileContent, err := os.ReadFile(filepath.Join(dir, files[0].Name()))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check file content
	if !bytes.Contains(fileContent, []byte("REQUEST: GET")) {
		t.Errorf("File doesn't contain request info")
	}
	if !bytes.Contains(fileContent, []byte("RESPONSE: 200 OK")) {
		t.Errorf("File doesn't contain response info")
	}
	if !bytes.Contains(fileContent, []byte("test response")) {
		t.Errorf("File doesn't contain response body")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient(http.DefaultTransport)
	
	// Check we got a client with our transport
	transport, ok := client.Transport.(*Recorder)
	if !ok {
		t.Fatalf("Expected client transport to be *Recorder, got %T", client.Transport)
	}
	
	// Check the transport is configured correctly
	if transport.Transport != http.DefaultTransport {
		t.Errorf("Transport not configured correctly")
	}
}

func TestRecorderReset(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a recorder
	recorder := NewRecorder(http.DefaultTransport)

	// Create and execute two requests
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := recorder.RoundTrip(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		resp.Body.Close()
	}

	// Check that we have two records
	if len(recorder.Records()) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(recorder.Records()))
	}

	// Reset the recorder
	recorder.Reset()

	// Check that we have no records
	if len(recorder.Records()) != 0 {
		t.Fatalf("Expected 0 records after reset, got %d", len(recorder.Records()))
	}
} 