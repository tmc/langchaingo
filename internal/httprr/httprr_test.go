package httprr

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransport_Record(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "test response"}`))
	}))
	defer server.Close()

	// Create temp cassette file
	tmpDir := t.TempDir()
	cassettePath := filepath.Join(tmpDir, "test.json")

	// Create recording transport
	transport := NewTransport(cassettePath, ModeRecord)
	transport.Transport = http.DefaultTransport

	client := &http.Client{Transport: transport}

	// Make request
	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify cassette was created
	if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
		t.Fatalf("Cassette file was not created: %s", cassettePath)
	}

	// Load and verify cassette content
	data, err := os.ReadFile(cassettePath)
	if err != nil {
		t.Fatalf("Failed to read cassette: %v", err)
	}

	var cassette Cassette
	err = json.Unmarshal(data, &cassette)
	if err != nil {
		t.Fatalf("Failed to unmarshal cassette: %v", err)
	}

	if len(cassette.Interactions) != 1 {
		t.Errorf("Expected 1 interaction, got %d", len(cassette.Interactions))
	}

	interaction := cassette.Interactions[0]
	if interaction.Request.Method != "GET" {
		t.Errorf("Expected GET method, got %s", interaction.Request.Method)
	}

	if !strings.Contains(interaction.Request.URL, "/test") {
		t.Errorf("Expected URL to contain /test, got %s", interaction.Request.URL)
	}

	if interaction.Response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", interaction.Response.StatusCode)
	}

	if !strings.Contains(interaction.Response.Body, "test response") {
		t.Errorf("Expected response body to contain 'test response', got %s", interaction.Response.Body)
	}
}

func TestTransport_Replay(t *testing.T) {
	// Create temp cassette file with test data
	tmpDir := t.TempDir()
	cassettePath := filepath.Join(tmpDir, "test.json")

	cassette := Cassette{
		Name:         "test",
		Interactions: []Interaction{},
	}

	// Create mock interaction
	req := &http.Request{
		Method: "GET",
		URL:    mustParseURL("http://example.com/test"),
		Header: make(http.Header),
		Body:   nil,
	}

	interaction := Interaction{
		ID: generateTestRequestID(req),
		Request: Request{
			Method:  "GET",
			URL:     "http://example.com/test",
			Headers: make(http.Header),
			Body:    "",
		},
		Response: Response{
			Status:     "200 OK",
			StatusCode: 200,
			Headers:    make(http.Header),
			Body:       `{"message": "replayed response"}`,
		},
	}
	cassette.Interactions = append(cassette.Interactions, interaction)

	// Save cassette
	data, err := json.MarshalIndent(cassette, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal cassette: %v", err)
	}

	err = os.WriteFile(cassettePath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write cassette: %v", err)
	}

	// Create replay transport
	transport := NewTransport(cassettePath, ModeReplay)
	client := &http.Client{Transport: transport}

	// Make request
	resp, err := client.Get("http://example.com/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := new(bytes.Buffer)
	body.ReadFrom(resp.Body)
	if !strings.Contains(body.String(), "replayed response") {
		t.Errorf("Expected response body to contain 'replayed response', got %s", body.String())
	}
}

func TestTransport_Disabled(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	// Create disabled transport
	tmpDir := t.TempDir()
	cassettePath := filepath.Join(tmpDir, "test.json")
	transport := NewTransport(cassettePath, ModeDisabled)
	transport.Transport = http.DefaultTransport

	client := &http.Client{Transport: transport}

	// Make request
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify no cassette was created
	if _, err := os.Stat(cassettePath); !os.IsNotExist(err) {
		t.Errorf("Cassette file should not have been created in disabled mode")
	}
}

func TestClient(t *testing.T) {
	tmpDir := t.TempDir()
	cassettePath := filepath.Join(tmpDir, "test.json")

	client := Client(cassettePath, ModeRecord)
	if client == nil {
		t.Fatal("Client should not be nil")
	}

	transport, ok := client.Transport.(*Transport)
	if !ok {
		t.Fatal("Transport should be *httprr.Transport")
	}

	if transport.Mode != ModeRecord {
		t.Errorf("Expected ModeRecord, got %v", transport.Mode)
	}

	if transport.CassettePath != cassettePath {
		t.Errorf("Expected cassette path %s, got %s", cassettePath, transport.CassettePath)
	}
}

func TestRecordingClient(t *testing.T) {
	cassettePath := "test.json"
	client := RecordingClient(cassettePath)

	transport, ok := client.Transport.(*Transport)
	if !ok {
		t.Fatal("Transport should be *httprr.Transport")
	}

	if transport.Mode != ModeRecord {
		t.Errorf("Expected ModeRecord, got %v", transport.Mode)
	}
}

func TestReplayClient(t *testing.T) {
	cassettePath := "test.json"
	client := ReplayClient(cassettePath)

	transport, ok := client.Transport.(*Transport)
	if !ok {
		t.Fatal("Transport should be *httprr.Transport")
	}

	if transport.Mode != ModeReplay {
		t.Errorf("Expected ModeReplay, got %v", transport.Mode)
	}
}

// Helper functions for tests

func mustParseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u
}

func generateTestRequestID(req *http.Request) string {
	transport := &Transport{}
	return transport.generateRequestID(req)
}