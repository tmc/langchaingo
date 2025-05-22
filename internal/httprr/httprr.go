// Package httprr provides HTTP request/response recording functionality
// for testing and debugging purposes. It allows recording HTTP interactions
// and replaying them later for consistent testing.
package httprr

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Recording represents a recorded HTTP interaction
type Recording struct {
	Request  RecordedRequest  `json:"request"`
	Response RecordedResponse `json:"response"`
	ID       string           `json:"id"`
	Created  time.Time        `json:"created"`
}

// RecordedRequest represents the HTTP request part of a recording
type RecordedRequest struct {
	Method string            `json:"method"`
	URL    string            `json:"url"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
}

// RecordedResponse represents the HTTP response part of a recording
type RecordedResponse struct {
	StatusCode int               `json:"status_code"`
	Header     map[string]string `json:"header"`
	Body       string            `json:"body"`
}

// Recorder handles recording and replaying HTTP interactions
type Recorder struct {
	recordingsDir string
	mode          Mode
	recordings    map[string]*Recording
}

// Mode defines the recording mode
type Mode int

const (
	// ModeRecord records all HTTP interactions
	ModeRecord Mode = iota
	// ModeReplay replays recorded interactions
	ModeReplay
	// ModeRecordOnce records if no recording exists, otherwise replays
	ModeRecordOnce
)

// New creates a new Recorder instance
func New(recordingsDir string, mode Mode) *Recorder {
	r := &Recorder{
		recordingsDir: recordingsDir,
		mode:          mode,
		recordings:    make(map[string]*Recording),
	}
	
	// Load existing recordings if in replay mode
	if mode == ModeReplay || mode == ModeRecordOnce {
		r.loadRecordings()
	}
	
	return r
}

// Transport returns an http.RoundTripper that records or replays HTTP interactions
func (r *Recorder) Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &recordingTransport{
		recorder: r,
		base:     base,
	}
}

// recordingTransport implements http.RoundTripper
type recordingTransport struct {
	recorder *Recorder
	base     http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	id := rt.recorder.generateID(req)
	
	// Check if we should replay
	if rt.recorder.mode == ModeReplay || 
	   (rt.recorder.mode == ModeRecordOnce && rt.recorder.hasRecording(id)) {
		return rt.recorder.replay(req, id)
	}
	
	// Record the interaction
	return rt.recorder.record(req, id, rt.base)
}

// generateID creates a unique identifier for the request
func (r *Recorder) generateID(req *http.Request) string {
	// Create a deterministic ID based on method, URL, and sanitized body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	
	// Sort headers for consistent ID generation
	var headerKeys []string
	for k := range req.Header {
		// Skip headers that might vary between requests
		if !isVariableHeader(k) {
			headerKeys = append(headerKeys, k)
		}
	}
	sort.Strings(headerKeys)
	
	var headerStr strings.Builder
	for _, k := range headerKeys {
		headerStr.WriteString(k)
		headerStr.WriteString(":")
		headerStr.WriteString(strings.Join(req.Header[k], ","))
		headerStr.WriteString(";")
	}
	
	content := fmt.Sprintf("%s|%s|%s|%s", 
		req.Method, 
		req.URL.String(), 
		headerStr.String(),
		sanitizeBody(string(bodyBytes)))
	
	return fmt.Sprintf("%x", md5.Sum([]byte(content)))
}

// isVariableHeader returns true for headers that should be ignored when generating IDs
func isVariableHeader(header string) bool {
	variableHeaders := []string{
		"Date", "User-Agent", "Authorization", "X-Request-Id", 
		"X-Trace-Id", "X-Span-Id", "Request-Id", "Correlation-Id",
	}
	
	header = strings.ToLower(header)
	for _, vh := range variableHeaders {
		if strings.ToLower(vh) == header {
			return true
		}
	}
	return false
}

// sanitizeBody removes dynamic content from request bodies for consistent ID generation
func sanitizeBody(body string) string {
	// For JSON bodies, we might want to remove timestamp fields or other dynamic content
	if strings.TrimSpace(body) == "" {
		return ""
	}
	
	// Simple sanitization - remove common dynamic fields
	// In a real implementation, this could be more sophisticated
	body = strings.ReplaceAll(body, `"timestamp":"[^"]*"`, `"timestamp":"REDACTED"`)
	body = strings.ReplaceAll(body, `"id":"[^"]*"`, `"id":"REDACTED"`)
	
	return body
}

// hasRecording checks if a recording exists for the given ID
func (r *Recorder) hasRecording(id string) bool {
	_, exists := r.recordings[id]
	return exists
}

// record performs the actual HTTP request and records the interaction
func (r *Recorder) record(req *http.Request, id string, base http.RoundTripper) (*http.Response, error) {
	// Make the actual request
	resp, err := base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	
	// Record the interaction
	recording := &Recording{
		ID:      id,
		Created: time.Now(),
	}
	
	// Record request
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}
	
	recording.Request = RecordedRequest{
		Method: req.Method,
		URL:    req.URL.String(),
		Header: flattenHeaders(req.Header),
		Body:   string(reqBody),
	}
	
	// Record response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	
	recording.Response = RecordedResponse{
		StatusCode: resp.StatusCode,
		Header:     flattenHeaders(resp.Header),
		Body:       string(respBody),
	}
	
	// Save recording
	r.recordings[id] = recording
	if err := r.saveRecording(recording); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to save recording %s: %v\n", id, err)
	}
	
	return resp, nil
}

// replay returns a recorded response for the given request
func (r *Recorder) replay(req *http.Request, id string) (*http.Response, error) {
	recording, exists := r.recordings[id]
	if !exists {
		return nil, fmt.Errorf("no recording found for request ID: %s", id)
	}
	
	// Create response from recording
	resp := &http.Response{
		StatusCode: recording.Response.StatusCode,
		Header:     expandHeaders(recording.Response.Header),
		Body:       io.NopCloser(strings.NewReader(recording.Response.Body)),
		Request:    req,
	}
	
	return resp, nil
}

// flattenHeaders converts http.Header to map[string]string for JSON serialization
func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = strings.Join(v, ", ")
		}
	}
	return result
}

// expandHeaders converts map[string]string back to http.Header
func expandHeaders(headers map[string]string) http.Header {
	result := make(http.Header)
	for k, v := range headers {
		result[k] = strings.Split(v, ", ")
	}
	return result
}

// loadRecordings loads all recordings from the recordings directory
func (r *Recorder) loadRecordings() error {
	if _, err := os.Stat(r.recordingsDir); os.IsNotExist(err) {
		return nil // No recordings directory, nothing to load
	}
	
	files, err := filepath.Glob(filepath.Join(r.recordingsDir, "*.json"))
	if err != nil {
		return err
	}
	
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files we can't read
		}
		
		var recording Recording
		if err := json.Unmarshal(data, &recording); err != nil {
			continue // Skip invalid recordings
		}
		
		r.recordings[recording.ID] = &recording
	}
	
	return nil
}

// saveRecording saves a recording to disk
func (r *Recorder) saveRecording(recording *Recording) error {
	if err := os.MkdirAll(r.recordingsDir, 0755); err != nil {
		return err
	}
	
	filename := filepath.Join(r.recordingsDir, recording.ID+".json")
	data, err := json.MarshalIndent(recording, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// Client returns an http.Client configured with the recording transport
func (r *Recorder) Client() *http.Client {
	return &http.Client{
		Transport: r.Transport(nil),
	}
}

// Reset clears all recordings and optionally removes recording files
func (r *Recorder) Reset(removeFiles bool) error {
	r.recordings = make(map[string]*Recording)
	
	if removeFiles {
		return os.RemoveAll(r.recordingsDir)
	}
	
	return nil
}