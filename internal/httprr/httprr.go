// Package httprr provides a way to record and replay HTTP requests and responses.
// The name is inspired by "http recorder/replayer" similar to the way Russ Cox would name it.
package httprr

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Mode defines whether the recorder is recording real HTTP calls or replaying from stored data.
type Mode string

const (
	// ModeRecord indicates that real HTTP requests should be made and recorded.
	ModeRecord Mode = "record"

	// ModeReplay indicates that responses should be replayed from previous recordings.
	ModeReplay Mode = "replay"

	// ModePassthrough indicates that real HTTP requests should be made but not recorded.
	ModePassthrough Mode = "passthrough"
)

// Recorder is a RoundTripper that records HTTP requests and responses.
type Recorder struct {
	// Transport is the underlying RoundTripper to use for the actual requests.
	Transport http.RoundTripper

	// Dir is the directory where recordings will be saved.
	// If empty, recordings are not saved to disk.
	Dir string

	// Mode determines whether to replay recorded responses instead of making actual requests.
	// If set to ModeReplay and a matching recording exists, it will be used instead of making a real request.
	Mode Mode

	// Pattern is the pattern to use for recording filenames.
	// Default is "%s-%d.txt" where %s is the request method and %d is a timestamp.
	Pattern string

	// MatchHeaders is a list of headers to use when matching requests for replay.
	// If empty, only the URL and method are used.
	MatchHeaders []string

	// StrictMatching requires that the request match exactly for replay.
	// If false, a less strict matching algorithm is used (e.g., ignoring some headers).
	StrictMatching bool

	mu      sync.Mutex
	records []*Record
	cache   map[string]*Record // Used for replay mode
}

// Record represents a recorded HTTP request and response.
type Record struct {
	// When is when the request was made.
	When time.Time

	// Request is the original request.
	Request *http.Request

	// RequestDump is the dumped request.
	RequestDump []byte

	// Response is the response.
	Response *http.Response

	// ResponseDump is the dumped response.
	ResponseDump []byte

	// Error is any error that occurred during the request.
	Error error
}

// NewRecorder creates a new Recorder.
func NewRecorder(transport http.RoundTripper) *Recorder {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Recorder{
		Transport: transport,
		Pattern:   "%s-%d.txt",
		Mode:      ModeRecord,
		cache:     make(map[string]*Record),
	}
}

// NewClient creates a new http.Client with a Recorder as its Transport.
func NewClient(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: NewRecorder(transport),
	}
}

// ReplayClient creates a new http.Client with a Recorder in replay mode.
func ReplayClient(dir string, transport http.RoundTripper) *http.Client {
	recorder := NewRecorder(transport)
	recorder.Mode = ModeReplay
	recorder.Dir = dir

	// Load the recordings
	if err := recorder.loadRecordings(); err != nil {
		fmt.Fprintf(os.Stderr, "httprr: failed to load recordings: %v\n", err)
	}

	return &http.Client{
		Transport: recorder,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.Mode == ModeReplay {
		// Try to find a matching recording
		if resp, err := r.replayResponse(req); resp != nil || err != nil {
			return resp, err
		}
		// If no matching recording is found and we're in strict replay mode, return an error
		if r.StrictMatching {
			return nil, fmt.Errorf("httprr: no matching recording found for %s %s", req.Method, req.URL.String())
		}
		// If not strict, fall through to make a real request
		fmt.Fprintf(os.Stderr, "httprr: no recording found for %s %s, making real request\n", req.Method, req.URL.String())
	}

	// Create a copy of the request body so we can read it multiple times
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// Dump the request before sending
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, fmt.Errorf("httprr: failed to dump request: %w", err)
	}

	// Reset the request body for the actual request
	if req.Body != nil {
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// Make the actual request if we're not in replay mode
	var resp *http.Response
	var reqErr error
	if r.Mode != ModeReplay || !r.StrictMatching {
		resp, reqErr = r.Transport.RoundTrip(req)
	}

	// Create a record
	record := &Record{
		When:        time.Now(),
		Request:     req,
		RequestDump: reqDump,
		Response:    resp,
		Error:       reqErr,
	}

	// If the response is successful, dump it
	if reqErr == nil && resp != nil {
		// Read the response body
		var respBody []byte
		if resp.Body != nil {
			respBody, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		}

		// Dump the response
		respDump, dumpErr := httputil.DumpResponse(resp, true)
		if dumpErr != nil {
			return resp, fmt.Errorf("httprr: failed to dump response: %w", dumpErr)
		}
		record.ResponseDump = respDump
	}

	// Store the record
	r.mu.Lock()
	r.records = append(r.records, record)
	// Also add to the cache for replay
	key := r.requestKey(req)
	r.cache[key] = record
	r.mu.Unlock()

	// Write the record to disk if Dir is set and we're in record mode
	if r.Dir != "" && r.Mode == ModeRecord {
		err := r.SaveRecord(record)
		if err != nil {
			fmt.Fprintf(os.Stderr, "httprr: failed to save record: %v\n", err)
		}
	}

	return resp, reqErr
}

// Records returns all recorded request/response pairs.
func (r *Recorder) Records() []*Record {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*Record{}, r.records...)
}

// Reset clears all recorded request/response pairs.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = nil
	r.cache = make(map[string]*Record)
}

// SaveRecord writes a record to disk.
func (r *Recorder) SaveRecord(record *Record) error {
	if r.Dir == "" {
		return nil
	}

	// Ensure the directory exists
	if err := os.MkdirAll(r.Dir, 0755); err != nil {
		return err
	}

	// Create a filename
	method := record.Request.Method
	if method == "" {
		method = "UNKNOWN"
	}
	timestamp := record.When.UnixNano()
	filename := fmt.Sprintf(r.Pattern, method, timestamp)
	path := filepath.Join(r.Dir, filename)

	// Open the file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the request
	if _, err := fmt.Fprintf(f, "REQUEST: %s %s\n", record.Request.Method, record.Request.URL); err != nil {
		return err
	}
	if _, err := f.Write(record.RequestDump); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "\n\n"); err != nil {
		return err
	}

	// Write the response
	if record.Error != nil {
		if _, err := fmt.Fprintf(f, "ERROR: %v\n", record.Error); err != nil {
			return err
		}
	} else if record.Response != nil {
		if _, err := fmt.Fprintf(f, "RESPONSE: %s\n", record.Response.Status); err != nil {
			return err
		}
		if _, err := f.Write(record.ResponseDump); err != nil {
			return err
		}
	}

	return nil
}

// loadRecordings loads all recordings from the directory.
func (r *Recorder) loadRecordings() error {
	if r.Dir == "" {
		return fmt.Errorf("httprr: no directory specified for recordings")
	}

	// Ensure the directory exists
	if _, err := os.Stat(r.Dir); os.IsNotExist(err) {
		return fmt.Errorf("httprr: recordings directory %s does not exist", r.Dir)
	}

	// List all files in the directory
	files, err := os.ReadDir(r.Dir)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Read file
		data, err := os.ReadFile(filepath.Join(r.Dir, file.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "httprr: failed to read file %s: %v\n", file.Name(), err)
			continue
		}

		// Parse file to extract request and response
		record, err := r.parseRecordingFile(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "httprr: failed to parse file %s: %v\n", file.Name(), err)
			continue
		}

		// Add to cache and records
		key := r.requestKey(record.Request)
		r.cache[key] = record
		r.records = append(r.records, record)
	}

	return nil
}

// parseRecordingFile parses a recording file and returns a record.
func (r *Recorder) parseRecordingFile(data []byte) (*Record, error) {
	// Split into request and response parts
	parts := bytes.Split(data, []byte("\n\n"))
	if len(parts) < 2 {
		return nil, fmt.Errorf("httprr: invalid recording format")
	}

	requestData := parts[0]
	responseData := parts[1]

	// Parse request - use bufio.NewReader instead of bytes.NewReader
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(requestData)))
	if err != nil {
		// Try a different approach - the file might not be directly readable as an HTTP request
		// Create a fake request from the available data
		urlLine := string(bytes.Split(requestData, []byte("\n"))[0])
		parts := strings.Split(urlLine, " ")
		if len(parts) < 3 {
			return nil, fmt.Errorf("httprr: cannot parse request: %w", err)
		}
		method := parts[1]
		urlStr := parts[2]

		req, err = http.NewRequest(method, urlStr, bytes.NewReader([]byte{}))
		if err != nil {
			return nil, fmt.Errorf("httprr: cannot create request: %w", err)
		}
	}

	// Parse response - use bufio.NewReader instead of bytes.NewReader
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(responseData)), req)
	if err != nil {
		// Handle files with ERROR: lines instead of responses
		if bytes.HasPrefix(responseData, []byte("ERROR:")) {
			errText := string(bytes.TrimPrefix(responseData, []byte("ERROR:")))
			return &Record{
				Request:     req,
				RequestDump: requestData,
				Error:       fmt.Errorf(strings.TrimSpace(errText)),
			}, nil
		}
		return nil, fmt.Errorf("httprr: cannot parse response: %w", err)
	}

	return &Record{
		When:         time.Now(),
		Request:      req,
		RequestDump:  requestData,
		Response:     resp,
		ResponseDump: responseData,
	}, nil
}

// replayResponse finds a matching response for the given request.
func (r *Recorder) replayResponse(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.requestKey(req)
	if record, ok := r.cache[key]; ok {
		if record.Error != nil {
			return nil, record.Error
		}

		// Clone the response to avoid modifying the original
		respBody := bytes.NewBuffer(nil)
		if record.Response.Body != nil {
			body, _ := io.ReadAll(record.Response.Body)
			record.Response.Body = io.NopCloser(bytes.NewBuffer(body))
			respBody.Write(body)
		}

		// Create a new response with the same data
		return &http.Response{
			Status:           record.Response.Status,
			StatusCode:       record.Response.StatusCode,
			Proto:            record.Response.Proto,
			ProtoMajor:       record.Response.ProtoMajor,
			ProtoMinor:       record.Response.ProtoMinor,
			Header:           record.Response.Header.Clone(),
			Body:             io.NopCloser(respBody),
			ContentLength:    record.Response.ContentLength,
			TransferEncoding: record.Response.TransferEncoding,
			Close:            record.Response.Close,
			Uncompressed:     record.Response.Uncompressed,
			Trailer:          record.Response.Trailer,
			Request:          req,
		}, nil
	}

	return nil, nil
}

// requestKey generates a unique key for a request used for replay matching.
func (r *Recorder) requestKey(req *http.Request) string {
	if req == nil {
		return ""
	}

	h := sha256.New()

	// Always include method and URL
	fmt.Fprintf(h, "%s %s\n", req.Method, req.URL.String())

	// Include specified headers for matching if required
	if len(r.MatchHeaders) > 0 {
		for _, header := range r.MatchHeaders {
			if values, ok := req.Header[header]; ok {
				for _, value := range values {
					fmt.Fprintf(h, "%s: %s\n", header, value)
				}
			}
		}
	}

	// In strict mode, include the entire request body
	if r.StrictMatching && req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		h.Write(body)
	}

	return hex.EncodeToString(h.Sum(nil))
}
