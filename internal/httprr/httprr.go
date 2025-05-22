// Package httprr provides HTTP recording and replay functionality for tests.
package httprr

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Mode represents the recording/replay mode.
type Mode int

const (
	// ModeDisabled disables recording and replay.
	ModeDisabled Mode = iota
	// ModeRecord records HTTP interactions to files.
	ModeRecord
	// ModeReplay replays HTTP interactions from files.
	ModeReplay
)

// Transport is an HTTP transport that can record and replay HTTP interactions.
type Transport struct {
	Transport http.RoundTripper
	Mode      Mode
	CassettePath string
	cassette  *Cassette
}

// Cassette represents a recorded HTTP interaction session.
type Cassette struct {
	Name         string        `json:"name"`
	Interactions []Interaction `json:"interactions"`
	RecordedAt   time.Time     `json:"recorded_at"`
}

// Interaction represents a single HTTP request/response pair.
type Interaction struct {
	ID       string   `json:"id"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Request represents the recorded HTTP request.
type Request struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Headers http.Header `json:"headers"`
	Body   string      `json:"body"`
}

// Response represents the recorded HTTP response.
type Response struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
}

// NewTransport creates a new httprr transport.
func NewTransport(cassettePath string, mode Mode) *Transport {
	return &Transport{
		Transport:    http.DefaultTransport,
		Mode:         mode,
		CassettePath: cassettePath,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch t.Mode {
	case ModeDisabled:
		return t.Transport.RoundTrip(req)
	case ModeRecord:
		return t.record(req)
	case ModeReplay:
		return t.replay(req)
	default:
		return t.Transport.RoundTrip(req)
	}
}

func (t *Transport) record(req *http.Request) (*http.Response, error) {
	// Make the actual HTTP request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Create interaction record
	interaction, err := t.createInteraction(req, resp)
	if err != nil {
		return resp, err // Return the response even if recording fails
	}

	// Load or create cassette
	if t.cassette == nil {
		t.cassette, err = t.loadOrCreateCassette()
		if err != nil {
			return resp, err
		}
	}

	// Add interaction to cassette
	t.cassette.Interactions = append(t.cassette.Interactions, *interaction)

	// Save cassette
	err = t.saveCassette()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (t *Transport) replay(req *http.Request) (*http.Response, error) {
	// Load cassette if not already loaded
	if t.cassette == nil {
		var err error
		t.cassette, err = t.loadCassette()
		if err != nil {
			return nil, fmt.Errorf("failed to load cassette: %w", err)
		}
	}

	// Find matching interaction
	reqID := t.generateRequestID(req)
	for _, interaction := range t.cassette.Interactions {
		if interaction.ID == reqID {
			return t.createResponseFromInteraction(&interaction), nil
		}
	}

	return nil, fmt.Errorf("no matching interaction found for request: %s %s", req.Method, req.URL.String())
}

func (t *Transport) createInteraction(req *http.Request, resp *http.Response) (*Interaction, error) {
	// Read request body
	var reqBody string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		reqBody = string(bodyBytes)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Read response body
	var respBody string
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		respBody = string(bodyBytes)
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create interaction
	interaction := &Interaction{
		ID: t.generateRequestID(req),
		Request: Request{
			Method:  req.Method,
			URL:     req.URL.String(),
			Headers: req.Header.Clone(),
			Body:    reqBody,
		},
		Response: Response{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       respBody,
		},
	}

	return interaction, nil
}

func (t *Transport) generateRequestID(req *http.Request) string {
	var bodyHash string
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		hash := sha256.Sum256(bodyBytes)
		bodyHash = fmt.Sprintf("%x", hash)[:8]
	}

	// Create a unique ID based on method, URL, and body hash
	id := fmt.Sprintf("%s-%s-%s", req.Method, normalizeURL(req.URL), bodyHash)
	hash := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%x", hash)[:16]
}

func normalizeURL(u *url.URL) string {
	// Normalize URL for consistent matching
	normalized := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
	}
	return normalized.String()
}

func (t *Transport) createResponseFromInteraction(interaction *Interaction) *http.Response {
	header := make(http.Header)
	for k, v := range interaction.Response.Headers {
		header[k] = v
	}

	return &http.Response{
		Status:        interaction.Response.Status,
		StatusCode:    interaction.Response.StatusCode,
		Header:        header,
		Body:          io.NopCloser(strings.NewReader(interaction.Response.Body)),
		ContentLength: int64(len(interaction.Response.Body)),
	}
}

func (t *Transport) loadOrCreateCassette() (*Cassette, error) {
	cassette, err := t.loadCassette()
	if err != nil {
		// Create new cassette if loading fails
		return &Cassette{
			Name:         filepath.Base(t.CassettePath),
			Interactions: []Interaction{},
			RecordedAt:   time.Now(),
		}, nil
	}
	return cassette, nil
}

func (t *Transport) loadCassette() (*Cassette, error) {
	data, err := os.ReadFile(t.CassettePath)
	if err != nil {
		return nil, err
	}

	var cassette Cassette
	err = json.Unmarshal(data, &cassette)
	if err != nil {
		return nil, err
	}

	return &cassette, nil
}

func (t *Transport) saveCassette() error {
	// Ensure directory exists
	dir := filepath.Dir(t.CassettePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Marshal cassette to JSON
	data, err := json.MarshalIndent(t.cassette, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(t.CassettePath, data, 0644)
}

// Client creates an HTTP client with httprr transport.
func Client(cassettePath string, mode Mode) *http.Client {
	return &http.Client{
		Transport: NewTransport(cassettePath, mode),
	}
}

// RecordingClient creates an HTTP client in recording mode.
func RecordingClient(cassettePath string) *http.Client {
	return Client(cassettePath, ModeRecord)
}

// ReplayClient creates an HTTP client in replay mode.
func ReplayClient(cassettePath string) *http.Client {
	return Client(cassettePath, ModeReplay)
}