// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httprr implements HTTP record and replay, mainly for use in tests.
//
// [Open] creates a new [RecordReplay]. Whether it is recording or replaying
// is controlled by the -httprecord flag, which is defined by this package
// only in test programs (built by “go test”).
// See the [Open] documentation for more details.
//
// Note: This package has been adapted for use in the LangChainGo library with convienence
// functions for creating [RecordReplay] instances that are suitable for testing.
package httprr

import (
	"bufio"
	"bytes"
	"cmp"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	nethttputil "net/http/httputil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tmc/langchaingo/httputil"
)

var (
	record      = new(string)
	debug       = new(bool)
	httpDebug   = new(bool)
	recordDelay = new(time.Duration)
	recordMu    sync.Mutex
)

func init() {
	if testing.Testing() {
		record = flag.String("httprecord", "", "re-record traces for files matching `regexp`")
		debug = flag.Bool("httprecord-debug", false, "enable debug output for httprr recording details")
		httpDebug = flag.Bool("httpdebug", false, "enable HTTP request/response logging")
		recordDelay = flag.Duration("httprecord-delay", 0, "delay between HTTP requests during recording (helps avoid rate limits)")
	}
}

// A RecordReplay is an [http.RoundTripper] that can operate in two modes: record and replay.
//
// In record mode, the RecordReplay invokes another RoundTripper
// and logs the (request, response) pairs to a file.
//
// In replay mode, the RecordReplay responds to requests by finding
// an identical request in the log and sending the logged response.
type RecordReplay struct {
	file string            // file being read or written
	real http.RoundTripper // real HTTP connection

	mu        sync.Mutex
	reqScrub  []func(*http.Request) error // scrubbers for logging requests
	respScrub []func(*bytes.Buffer) error // scrubbers for logging responses
	replay    map[string]string           // if replaying, the log
	record    *os.File                    // if recording, the file being written
	writeErr  error                       // if recording, any write error encountered
	logger    *slog.Logger                // logger for debug output
}

// ScrubReq adds new request scrubbing functions to rr.
//
// Before using a request as a lookup key or saving it in the record/replay log,
// the [RecordReplay] calls each scrub function, in the order they were registered,
// to canonicalize non-deterministic parts of the request and remove secrets.
// Scrubbing only applies to a copy of the request used in the record/replay log;
// the unmodified original request is sent to the actual server in recording mode.
// A scrub function can assume that if req.Body is not nil, then it has type [*Body].
//
// Calling ScrubReq adds to the list of registered request scrubbing functions;
// it does not replace those registered by earlier calls.
func (rr *RecordReplay) ScrubReq(scrubs ...func(req *http.Request) error) {
	rr.reqScrub = append(rr.reqScrub, scrubs...)
}

// ScrubResp adds new response scrubbing functions to rr.
//
// Before using a response as a lookup key or saving it in the record/replay log,
// the [RecordReplay] calls each scrub function on a byte representation of the
// response, in the order they were registered, to canonicalize non-deterministic
// parts of the response and remove secrets.
//
// Calling ScrubResp adds to the list of registered response scrubbing functions;
// it does not replace those registered by earlier calls.
//
// Clients should be careful when loading the bytes into [*http.Response] using
// [http.ReadResponse]. This function can set [http.Response].Close to true even
// when the original response had it false. See code in go/src/net/http.Response.Write
// and go/src/net/http.Write for more info.
func (rr *RecordReplay) ScrubResp(scrubs ...func(*bytes.Buffer) error) {
	rr.respScrub = append(rr.respScrub, scrubs...)
}

// Recording reports whether the [RecordReplay] is in recording mode.
func (rr *RecordReplay) Recording() bool {
	return rr.record != nil
}

// Replaying reports whether the [RecordReplay] is in replaying mode.
func (rr *RecordReplay) Replaying() bool {
	return !rr.Recording()
}

// Open opens a new record/replay log in the named file and
// returns a [RecordReplay] backed by that file.
//
// By default Open expects the file to exist and contain a
// previously-recorded log of (request, response) pairs,
// which [RecordReplay.RoundTrip] consults to prepare its responses.
//
// If the command-line flag -httprecord is set to a non-empty
// regular expression that matches file, then Open creates
// the file as a new log. In that mode, [RecordReplay.RoundTrip]
// makes actual HTTP requests using rt but then logs the requests and
// responses to the file for replaying in a future run.
func Open(file string, rt http.RoundTripper) (*RecordReplay, error) {
	record, err := Recording(file)
	if err != nil {
		return nil, err
	}
	if record {
		return create(file, rt)
	}

	// Check if a compressed version exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if _, err := os.Stat(file + ".gz"); err == nil {
			file = file + ".gz"
		}
	}

	return open(file, rt)
}

// Recording reports whether the "-httprecord" flag is set
// for the given file.
// It returns an error if the flag is set to an invalid value.
func Recording(file string) (bool, error) {
	recordMu.Lock()
	defer recordMu.Unlock()
	if *record != "" {
		re, err := regexp.Compile(*record)
		if err != nil {
			return false, fmt.Errorf("invalid -httprecord flag: %w", err)
		}
		if re.MatchString(file) {
			return true, nil
		}
	}
	return false, nil
}

// setRecordForTesting sets the record flag value for testing purposes.
// It returns a function that restores the original value.
func setRecordForTesting(value string) func() {
	recordMu.Lock()
	defer recordMu.Unlock()
	old := *record
	*record = value
	return func() {
		recordMu.Lock()
		defer recordMu.Unlock()
		*record = old
	}
}

// creates a new record-mode RecordReplay in the file.
func create(file string, rt http.RoundTripper) (*RecordReplay, error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	// Write header line.
	// Each round-trip will write a new request-response record.
	if _, err := fmt.Fprintf(f, "httprr trace v1\n"); err != nil {
		// unreachable unless write error immediately after os.Create
		f.Close()
		return nil, err
	}
	rr := &RecordReplay{
		file:   file,
		real:   rt,
		record: f,
	}
	// Apply default scrubbing
	rr.ScrubReq(getDefaultRequestScrubbers()...)
	rr.ScrubResp(getDefaultResponseScrubbers()...)
	return rr, nil
}

// open opens a replay-mode RecordReplay using the data in the file.
func open(file string, rt http.RoundTripper) (*RecordReplay, error) {
	// Note: To handle larger traces without storing entirely in memory,
	// could instead read the file incrementally, storing a map[hash]offsets
	// and then reread the relevant part of the file during RoundTrip.

	var bdata []byte
	var err error

	// Check if file is compressed
	if strings.HasSuffix(file, ".gz") {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()

		bdata, err = io.ReadAll(gz)
		if err != nil {
			return nil, err
		}
	} else {
		bdata, err = os.ReadFile(file)
		if err != nil {
			return nil, err
		}
	}

	// Trace begins with header line.
	data := string(bdata)
	line, data, ok := strings.Cut(data, "\n")
	// Trim any trailing CR for compatibility with both LF and CRLF line endings
	line = strings.TrimSuffix(line, "\r")
	if !ok || line != "httprr trace v1" {
		return nil, fmt.Errorf("read %s: not an httprr trace", file)
	}

	replay := make(map[string]string)
	for data != "" {
		// Each record starts with a line of the form "n1 n2\n" (or "n1 n2\r\n")
		// followed by n1 bytes of request encoding and
		// n2 bytes of response encoding.
		line, data, ok = strings.Cut(data, "\n")
		line = strings.TrimSuffix(line, "\r")
		f1, f2, _ := strings.Cut(line, " ")
		n1, err1 := strconv.Atoi(f1)
		n2, err2 := strconv.Atoi(f2)
		if !ok || err1 != nil || err2 != nil || n1 > len(data) || n2 > len(data[n1:]) {
			return nil, fmt.Errorf("read %s: corrupt httprr trace", file)
		}
		var req, resp string
		req, resp, data = data[:n1], data[n1:n1+n2], data[n1+n2:]
		replay[req] = resp
	}

	rr := &RecordReplay{
		file:   file,
		real:   rt,
		replay: replay,
	}
	// Apply default scrubbing
	rr.ScrubReq(getDefaultRequestScrubbers()...)
	rr.ScrubResp(getDefaultResponseScrubbers()...)
	return rr, nil
}

// Client returns an http.Client using rr as its transport.
// It is a shorthand for:
//
//	return &http.Client{Transport: rr}
//
// For more complicated uses, use rr or the [RecordReplay.RoundTrip] method directly.
func (rr *RecordReplay) Client() *http.Client {
	return &http.Client{Transport: rr}
}

// A Body is an [io.ReadCloser] used as an HTTP request body.
// In a Scrubber, if req.Body != nil, then req.Body is guaranteed
// to have type [*Body], making it easy to access the body to change it.
type Body struct {
	Data       []byte
	ReadOffset int
}

// Read reads from the body, implementing [io.Reader].
func (b *Body) Read(p []byte) (int, error) {
	n := copy(p, b.Data[b.ReadOffset:])
	if n == 0 {
		return 0, io.EOF
	}
	b.ReadOffset += n
	return n, nil
}

// Close is a no-op, implementing [io.Closer].
func (b *Body) Close() error {
	return nil
}

// RoundTrip implements [http.RoundTripper].
//
// If rr has been opened in record mode, RoundTrip passes the requests on to
// the [http.RoundTripper] specified in the call to [Open] and then logs the
// (request, response) pair to the underlying file.
//
// If rr has been opened in replay mode, RoundTrip looks up the request in the log
// and then responds with the previously logged response.
// If the log does not contain req, RoundTrip returns an error.
func (rr *RecordReplay) RoundTrip(req *http.Request) (*http.Response, error) {
	// Debug: log headers at RoundTrip entry
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: RoundTrip entry",
			"User-Agent", req.Header.Get("User-Agent"),
			"x-goog-api-client", req.Header.Get("x-goog-api-client"))
	}

	// Log the request if httpdebug is enabled
	if rr.logger != nil && *httpDebug {
		if reqDump, err := nethttputil.DumpRequestOut(req, true); err == nil {
			rr.logger.Debug(string(reqDump))
		}
	}

	// Save the body before calling reqWire since it consumes it
	// This is needed because reqWire consumes the body and we need it for both
	// the replay lookup and the recording
	var bodyBytes []byte
	hasBody := false
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		hasBody = true
		// Set body as a Body type that can be reused
		req.Body = &Body{Data: bodyBytes}
	}

	reqWire, err := rr.reqWire(req)
	if err != nil {
		return nil, err
	}

	// Reset the body's read position for the actual request
	if hasBody {
		// Reset the read offset so the body can be read again
		if body, ok := req.Body.(*Body); ok {
			body.ReadOffset = 0
		}
	}

	// If we're in replay mode, replay a response.
	if rr.replay != nil {
		resp, err := rr.replayRoundTrip(req, reqWire)
		if err != nil {
			return nil, err
		}
		// Log the response if httpdebug is enabled
		if rr.logger != nil && *httpDebug {
			if respDump, err := nethttputil.DumpResponse(resp, true); err == nil {
				rr.logger.Debug(string(respDump))
			}
		}
		return resp, nil
	}

	// Otherwise run a real round trip and save the request-response pair.
	// But if we've had a log write error already, don't bother.
	if err := rr.writeError(); err != nil {
		return nil, err
	}
	if rr.real == nil {
		return nil, fmt.Errorf("httprr: no transport configured")
	}
	resp, err := rr.real.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log the response if httpdebug is enabled
	if rr.logger != nil && *httpDebug {
		if respDump, err := nethttputil.DumpResponse(resp, true); err == nil {
			rr.logger.Debug(string(respDump))
		}
	}

	// Add delay after request when recording (helps avoid rate limits)
	if rr.Recording() {
		delay := *recordDelay
		if delay == 0 {
			// Default to 1 second delay when recording
			delay = 1 * time.Second
		}
		time.Sleep(delay)
	}

	// Recompute reqWire after real RoundTrip to capture any headers that were added
	// The real transport may have modified the request (e.g., added headers)
	// Debug: Check if headers were added by real RoundTrip
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: After real RoundTrip",
			"User-Agent", req.Header.Get("User-Agent"),
			"x-goog-api-client", req.Header.Get("x-goog-api-client"))
	}

	// Reset body read position before second reqWire call
	if hasBody {
		if body, ok := req.Body.(*Body); ok {
			body.ReadOffset = 0
		}
	}

	reqWireForSaving, err := rr.reqWire(req)
	if err != nil {
		return nil, err
	}

	// Encode resp and decode to get a copy for our caller.
	respWire, err := rr.respWire(resp)
	if err != nil {
		return nil, err
	}
	if err := rr.writeLog(reqWireForSaving, respWire); err != nil {
		return nil, err
	}
	return resp, nil
}

// reqWire returns the wire-format HTTP request key to be
// used for request when saving to the log or looking up in a
// previously written log. It consumes the original req.Body
// but modifies req.Body to be an equivalent [*Body].
func (rr *RecordReplay) reqWire(req *http.Request) (string, error) {
	// rkey is the scrubbed request used as a lookup key.
	// Clone req including req.Body.
	rkey := req.Clone(context.Background())
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return "", err
		}
		req.Body = &Body{Data: body}
		rkey.Body = &Body{Data: bytes.Clone(body)}
	}

	// Canonicalize and scrub request key.
	// Debug: log the number of scrubbers and headers before scrubbing
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: before scrubbing",
			"scrubber_count", len(rr.reqScrub),
			"User-Agent", rkey.Header.Get("User-Agent"),
			"x-goog-api-client", rkey.Header.Get("x-goog-api-client"))
	}
	for _, scrub := range rr.reqScrub {
		if err := scrub(rkey); err != nil {
			return "", err
		}
	}
	// Debug: log headers after scrubbing
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: after scrubbing",
			"User-Agent", rkey.Header.Get("User-Agent"),
			"x-goog-api-client", rkey.Header.Get("x-goog-api-client"))
	}

	// Now that scrubbers are done potentially modifying body, set length.
	if rkey.Body != nil {
		rkey.ContentLength = int64(len(rkey.Body.(*Body).Data))
	}

	// Serialize rkey to produce the log entry.
	// Use WriteProxy to preserve the URL's scheme and format correctly
	var key strings.Builder
	if err := rkey.WriteProxy(&key); err != nil {
		return "", err
	}

	// Apply string-based scrubbing to normalize headers that may have been added during serialization
	result := key.String()

	// Normalize User-Agent header in the serialized request
	result = regexp.MustCompile(`(?m)^User-Agent: .*$`).ReplaceAllString(result, "User-Agent: langchaingo-httprr")

	// Remove OpenAI-Project header for consistency across recordings
	result = regexp.MustCompile(`(?m)^openai-project: .*\n`).ReplaceAllString(result, "")

	// Normalize x-goog-api-client header with version information
	result = regexp.MustCompile(`(?m)^x-goog-api-client: (.*)$`).ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.SplitN(match, ": ", 2)
		if len(parts) == 2 {
			normalized := normalizeGoogleAPIClientHeader(parts[1])
			return parts[0] + ": " + normalized
		}
		return match
	})

	return result, nil
}

// respWire returns the wire-format HTTP response log entry.
// It preserves the original response body while creating a copy for logging.
func (rr *RecordReplay) respWire(resp *http.Response) (string, error) {
	// Read the original body
	var bodyBytes []byte
	var err error
	if resp.Body != nil {
		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}
		// Replace the body with a fresh reader for the client
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Create a copy of the response for serialization
	respCopy := *resp
	if bodyBytes != nil {
		respCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		respCopy.ContentLength = int64(len(bodyBytes))
	}

	// Serialize the copy to produce the log entry
	var key bytes.Buffer
	if err := respCopy.Write(&key); err != nil {
		return "", err
	}

	// Close the copy's body since we're done with it
	if respCopy.Body != nil {
		respCopy.Body.Close()
	}

	// Apply scrubbers to the serialized data
	for _, scrub := range rr.respScrub {
		if err := scrub(&key); err != nil {
			return "", err
		}
	}
	return key.String(), nil
}

// replayRoundTrip implements RoundTrip using the replay log.
func (rr *RecordReplay) replayRoundTrip(req *http.Request, reqLog string) (*http.Response, error) {
	// Log the incoming request if debug is enabled
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: attempting to match request in replay cache",
			"method", req.Method,
			"url", req.URL.String(),
			"file", rr.file,
		)
		// Also dump the full request for detailed debugging
		if reqDump, err := nethttputil.DumpRequestOut(req, true); err == nil {
			rr.logger.Debug("httprr: request details\n" + string(reqDump))
		}
	}

	respLog, ok := rr.replay[reqLog]
	if !ok {
		if rr.logger != nil && *debug {
			rr.logger.Debug("httprr: request not found in replay cache",
				"method", req.Method,
				"url", req.URL.String(),
				"file", rr.file,
			)
		}
		return nil, fmt.Errorf("cached HTTP response not found for:\n%s\n\nHint: Re-run tests with -httprecord=. to record new HTTP interactions\nDebug flags: -httprecord-debug for recording details, -httpdebug for HTTP traffic", reqLog)
	}

	// Log that we found a match
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: found matching request in replay cache",
			"method", req.Method,
			"url", req.URL.String(),
			"file", rr.file,
			"response_size", len(respLog),
		)
	}

	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(respLog)), req)
	if err != nil {
		return nil, fmt.Errorf("read %s: corrupt httprr trace: %w", rr.file, err)
	}

	// Log the response being returned
	if rr.logger != nil && *debug {
		rr.logger.Debug("httprr: returning cached response",
			"status", resp.StatusCode,
			"content_length", resp.ContentLength,
		)
		// Also dump the full response for detailed debugging
		if respDump, err := nethttputil.DumpResponse(resp, true); err == nil {
			rr.logger.Debug("httprr: response details\n" + string(respDump))
		}
	}

	return resp, nil
}

// writeError reports any previous log write error.
func (rr *RecordReplay) writeError() error {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	return rr.writeErr
}

// writeLog writes the request-response pair to the log.
// If a write fails, writeLog arranges for rr.broken to return
// an error and deletes the underlying log.
func (rr *RecordReplay) writeLog(reqWire, respWire string) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if rr.writeErr != nil {
		// Unreachable unless concurrent I/O error.
		// Caller should have checked already.
		return rr.writeErr
	}

	_, err1 := fmt.Fprintf(rr.record, "%d %d\n", len(reqWire), len(respWire))
	_, err2 := rr.record.WriteString(reqWire)
	_, err3 := rr.record.WriteString(respWire)
	if err := cmp.Or(err1, err2, err3); err != nil {
		rr.writeErr = err
		rr.record.Close()
		os.Remove(rr.file)
		return err
	}

	return nil
}

// Close closes the [RecordReplay].
// It is a no-op in replay mode.
func (rr *RecordReplay) Close() error {
	if rr.writeErr != nil {
		return rr.writeErr
	}
	if rr.record != nil {
		return rr.record.Close()
	}
	return nil
}

// CleanFileName converts a test name to a clean filename suitable for recordings.
// It replaces path separators and other non-path-friendly characters with hyphens.
// For example:
//   - "TestMyFunction/subtest" becomes "TestMyFunction-subtest"
//   - "Test API/Complex_Case" becomes "Test-API-Complex_Case"
func CleanFileName(testName string) string {
	// Replace forward slashes (subtest separators) with hyphens
	clean := strings.ReplaceAll(testName, "/", "-")

	// Replace other potentially problematic characters
	clean = strings.ReplaceAll(clean, "\\", "-")
	clean = strings.ReplaceAll(clean, ":", "-")
	clean = strings.ReplaceAll(clean, "*", "-")
	clean = strings.ReplaceAll(clean, "?", "-")
	clean = strings.ReplaceAll(clean, "\"", "-")
	clean = strings.ReplaceAll(clean, "<", "-")
	clean = strings.ReplaceAll(clean, ">", "-")
	clean = strings.ReplaceAll(clean, "|", "-")
	clean = strings.ReplaceAll(clean, " ", "-")

	// Remove multiple consecutive hyphens
	re := regexp.MustCompile(`-+`)
	clean = re.ReplaceAllString(clean, "-")

	// Remove leading/trailing hyphens
	clean = strings.Trim(clean, "-")

	return clean
}

func logWriter(t *testing.T) io.Writer {
	t.Helper()
	return testWriter{t}
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(b []byte) (int, error) {
	w.t.Logf("%s", b)
	return len(b), nil
}

// OpenForTest creates a [RecordReplay] for the given test using a filename
// derived from the test name. The recording will be stored in a "testdata"
// subdirectory with a ".httprr" extension.
//
// The transport parameter is optional. If not provided (nil), it defaults to
// [httputil.DefaultTransport].
//
// Example usage:
//
//	func TestMyAPI(t *testing.T) {
//	    rr := httprr.OpenForTest(t, nil) // Uses httputil.DefaultTransport
//	    defer rr.Close()
//
//	    client := rr.Client()
//	    // use client for HTTP requests...
//	}
//
//	// Or with a custom transport:
//	func TestMyAPIWithCustomTransport(t *testing.T) {
//	    customTransport := &http.Transport{MaxIdleConns: 10}
//	    rr := httprr.OpenForTest(t, customTransport)
//	    defer rr.Close()
//
//	    client := rr.Client()
//	    // use client for HTTP requests...
//	}
//
// This will create/use a file at "testdata/TestMyAPI.httprr".
// OpenForEmbeddingTest creates a RecordReplay instance optimized for embedding tests.
// It automatically applies embedding JSON formatting to reduce file sizes.
func OpenForEmbeddingTest(t *testing.T, rt http.RoundTripper) *RecordReplay {
	rr := OpenForTest(t, rt)
	rr.ScrubResp(EmbeddingJSONFormatter())
	return rr
}

func OpenForTest(t *testing.T, rt http.RoundTripper) *RecordReplay {
	t.Helper()

	// Default to httputil.DefaultTransport if no transport provided
	if rt == nil {
		rt = httputil.DefaultTransport
	}

	testName := CleanFileName(t.Name())
	filename := filepath.Join("testdata", testName+".httprr")

	// Ensure testdata directory exists
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatalf("httprr: failed to create testdata directory: %v", err)
	}

	// Create logger for debug mode
	var logger *slog.Logger
	if *debug || *httpDebug {
		logger = slog.New(slog.NewTextHandler(logWriter(t), &slog.HandlerOptions{Level: slog.LevelDebug}))
		if *debug {
			rt = &httputil.LoggingTransport{
				Transport: rt,
				Logger:    logger,
			}
		}
	}

	// Check if we're in recording mode
	recording, err := Recording(filename)
	if err != nil {
		t.Fatal(err)
	}

	if recording && testing.Short() {
		t.Skipf("httprr: skipping recording for %s in short mode", filename)
	}

	if recording {
		// Recording mode: clean up existing files and create uncompressed
		cleanupExistingFiles(t, filename)
		rr, err := Open(filename, rt)
		if err != nil {
			t.Fatalf("httprr: failed to open recording file %s: %v", filename, err)
		}
		rr.logger = logger

		// Add selective scrubber for embedding responses only
		rr.ScrubResp(conditionalEmbeddingFormatter())

		t.Cleanup(func() { rr.Close() })
		return rr
	}

	// Replay mode: find the best existing file
	filename = findBestReplayFile(t, filename)
	rr, err := Open(filename, rt)
	if err != nil {
		t.Fatal(err)
	}
	rr.logger = logger
	return rr
}

// cleanupExistingFiles removes any existing files to avoid conflicts during recording
func cleanupExistingFiles(t *testing.T, baseFilename string) {
	t.Helper()
	filesToCheck := []string{baseFilename, baseFilename + ".gz"}

	for _, filename := range filesToCheck {
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				t.Logf("httprr: warning - failed to remove %s: %v", filename, err)
			}
		}
	}
}

// findBestReplayFile finds the best existing file for replay mode
func findBestReplayFile(t *testing.T, baseFilename string) string {
	t.Helper()
	compressedFilename := baseFilename + ".gz"

	uncompressedStat, uncompressedErr := os.Stat(baseFilename)
	compressedStat, compressedErr := os.Stat(compressedFilename)

	// Both files exist - use the newer one and warn
	if uncompressedErr == nil && compressedErr == nil {
		if uncompressedStat.ModTime().After(compressedStat.ModTime()) {
			t.Logf("httprr: found both files, using newer uncompressed version")
			return baseFilename
		} else {
			t.Logf("httprr: found both files, using newer compressed version")
			return compressedFilename
		}
	}

	// Prefer compressed file if only it exists
	if compressedErr == nil {
		return compressedFilename
	}

	// Return base filename (may or may not exist)
	return baseFilename
}

// SkipIfNoCredentialsAndRecordingMissing skips the test if required environment variables
// are not set and no httprr recording exists. This allows tests to gracefully
// skip when they cannot run.
//
// Example usage:
//
//	func TestMyAPI(t *testing.T) {
//	    httprr.SkipIfNoCredentialsAndRecordingMissing(t, "API_KEY", "API_URL")
//
//	    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    defer rr.Close()
//	    // use rr.Client() for HTTP requests...
//	}
func SkipIfNoCredentialsAndRecordingMissing(t *testing.T, envVars ...string) {
	t.Helper()
	if !hasExistingRecording(t) && !hasRequiredCredentials(envVars) {
		skipMessage := "no httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions\nDebug flags: -httprecord-debug for recording details, -httpdebug for HTTP traffic"

		if len(envVars) > 0 {
			missingEnvVars := []string{}
			for _, envVar := range envVars {
				if os.Getenv(envVar) == "" {
					missingEnvVars = append(missingEnvVars, envVar)
				}
			}
			skipMessage = fmt.Sprintf("%s not set and %s", strings.Join(missingEnvVars, ","), skipMessage)
		}

		t.Skip(skipMessage)
	}
}

// hasRequiredCredentials checks if any of the required environment variables are set
func hasRequiredCredentials(envVars []string) bool {
	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}

// hasExistingRecording checks if any recording exists for the current test
func hasExistingRecording(t *testing.T) bool {
	t.Helper()
	testName := CleanFileName(t.Name())
	baseFilename := filepath.Join("testdata", testName+".httprr")

	_, uncompressedErr := os.Stat(baseFilename)
	_, compressedErr := os.Stat(baseFilename + ".gz")

	return uncompressedErr == nil || compressedErr == nil
}

// normalizeGoogleAPIClientHeader normalizes version information in the x-goog-api-client header
// to avoid test failures when dependencies are updated. It preserves the exact byte count
// by padding with spaces to maintain httprr recording integrity.
//
// Example input:  "gl-go/1.24.4 gccl/v0.15.1 genai-go/0.15.1 gapic/0.7.0 gax/2.14.1 rest/UNKNOWN"
// Example output: "gl-go/X.XX.X gccl/vX.XX.X genai-go/X.XX.X gapic/X.X.X gax/X.XX.X rest/UNKNOWN"
func normalizeGoogleAPIClientHeader(header string) string {
	originalLen := len(header)

	// Replace each version segment while preserving its length
	versionPattern := regexp.MustCompile(`(/v?)(\d+\.\d+(?:\.\d+)?)`)

	normalized := versionPattern.ReplaceAllStringFunc(header, func(match string) string {
		// Find the slash and optional 'v'
		slashIdx := strings.Index(match, "/")
		prefix := match[:slashIdx+1]
		if strings.HasPrefix(match[slashIdx+1:], "v") {
			prefix += "v"
		}

		// Calculate how many characters we need for the version part
		versionPart := match[len(prefix):]
		versionLen := len(versionPart)

		// Create a normalized version string that matches the exact length
		// Count the dots in the original to preserve structure
		dotCount := strings.Count(versionPart, ".")

		var replacement string
		if dotCount == 0 {
			// No dots, just replace with X's
			replacement = strings.Repeat("X", versionLen)
		} else if dotCount == 1 {
			// Format: X.X or XX.X etc
			if versionLen == 3 {
				replacement = "X.X"
			} else {
				// Distribute X's around the dot
				xBefore := (versionLen - 1) / 2
				xAfter := versionLen - 1 - xBefore
				replacement = strings.Repeat("X", xBefore) + "." + strings.Repeat("X", xAfter)
			}
		} else if dotCount == 2 {
			// Format: X.X.X, X.XX.X, etc.
			// Distribute X's around the dots fairly
			switch versionLen {
			case 5:
				replacement = "X.X.X"
			case 6:
				replacement = "X.XX.X"
			case 7:
				replacement = "X.XX.XX"
			default:
				// Generic case: distribute evenly
				segLen := (versionLen - 2) / 3
				remainder := (versionLen - 2) % 3
				seg1 := segLen + min(1, remainder)
				seg2 := segLen + min(1, max(0, remainder-1))
				seg3 := segLen
				replacement = strings.Repeat("X", seg1) + "." + strings.Repeat("X", seg2) + "." + strings.Repeat("X", seg3)
			}
		} else {
			// More than 2 dots or other format, just preserve length with X's
			replacement = strings.Repeat("X", versionLen)
		}

		return prefix + replacement
	})

	// Ensure the result has the exact same length
	if len(normalized) < originalLen {
		normalized += strings.Repeat(" ", originalLen-len(normalized))
	} else if len(normalized) > originalLen {
		normalized = normalized[:originalLen]
	}

	return normalized
}

// normalizeVersionHeader is a general-purpose version normalizer for headers containing
// version information in various formats.
func normalizeVersionHeader(header string) string {
	normalized := header

	// Pattern 1: Go version format (go1.21.0) - handle first to avoid conflict with semver
	goVersionPattern := regexp.MustCompile(`\bgo\d+\.\d+(\.\d+)?\b`)
	normalized = goVersionPattern.ReplaceAllString(normalized, "goX.X.X")

	// Pattern 2: Date-based versions with dots (2024.08.15)
	dotDatePattern := regexp.MustCompile(`\b20\d{2}\.\d{2}\.\d{2}\b`)
	normalized = dotDatePattern.ReplaceAllString(normalized, "XXXX.XX.XX")

	// Pattern 3: Date-based versions with dashes (2024-08-15)
	dashDatePattern := regexp.MustCompile(`\b20\d{2}-\d{2}-\d{2}\b`)
	normalized = dashDatePattern.ReplaceAllString(normalized, "XXXX.XX.XX")

	// Pattern 4: Compact date versions (20240815)
	compactDatePattern := regexp.MustCompile(`\b20\d{6}\b`)
	normalized = compactDatePattern.ReplaceAllString(normalized, "XXXX.XX.XX")

	// Pattern 5: Semantic versions (1.2.3, v1.2.3, 1.2, etc.) - do this last
	semverPattern := regexp.MustCompile(`\bv?\d+\.\d+(\.\d+)?(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?\b`)
	normalized = semverPattern.ReplaceAllString(normalized, "X.X.X")

	return normalized
}

// getDefaultRequestScrubbers returns the default request scrubbing functions to remove
// sensitive headers and API keys from request recordings.
func getDefaultRequestScrubbers() []func(*http.Request) error {
	return []func(*http.Request) error{
		func(req *http.Request) error {
			// Iterate through all headers to find any containing api-key, api-token, token, or authorization (case insensitive)
			for header, values := range req.Header {
				headerLower := strings.ToLower(header)
				if strings.Contains(headerLower, "api-key") ||
					strings.Contains(headerLower, "api-token") ||
					strings.Contains(headerLower, "token") ||
					headerLower == "authorization" {

					// Special handling for Authorization header
					if headerLower == "authorization" && len(values) > 0 {
						// Preserve the auth type (Bearer, Basic, etc.) but scrub the token
						authValue := values[0]
						parts := strings.SplitN(authValue, " ", 2)
						if len(parts) == 2 {
							req.Header.Set(header, parts[0]+" test-api-key")
						} else {
							req.Header.Set(header, "test-api-key")
						}
					} else {
						req.Header.Set(header, "test-api-key")
					}
				}
			}

			// Scrub sensitive query parameters
			q := req.URL.Query()
			for param := range q {
				paramLower := strings.ToLower(param)
				if strings.Contains(paramLower, "api_key") ||
					strings.Contains(paramLower, "api-key") ||
					strings.Contains(paramLower, "api-token") ||
					strings.Contains(paramLower, "token") ||
					strings.Contains(paramLower, "key") {
					q.Set(param, "test-api-key")
				}
			}
			req.URL.RawQuery = q.Encode()

			// Munge Openai-Organization header to a test value
			if req.Header.Get("Openai-Organization") != "" {
				req.Header.Set("Openai-Organization", "lcgo-tst")
			}
			if req.Header.Get("openai-organization") != "" {
				req.Header.Set("openai-organization", "lcgo-tst")
			}

			// Normalize User-Agent to avoid version-specific differences
			// Many Go libraries include version information in User-Agent
			if ua := req.Header.Get("User-Agent"); ua != "" {
				// Set to a consistent value, removing all version information
				req.Header.Set("User-Agent", "langchaingo-httprr")
			}

			// Normalize version information in x-goog-api-client header
			// This header contains library version information that changes with dependency updates
			if googClient := req.Header.Get("x-goog-api-client"); googClient != "" {
				normalized := normalizeGoogleAPIClientHeader(googClient)
				req.Header.Set("x-goog-api-client", normalized)
			}

			// Normalize other potential version headers
			// AWS SDK version headers
			if amzSdk := req.Header.Get("x-amz-user-agent"); amzSdk != "" {
				normalized := normalizeVersionHeader(amzSdk)
				req.Header.Set("x-amz-user-agent", normalized)
			}

			// Azure SDK version headers
			if azureSdk := req.Header.Get("x-ms-client-request-id"); azureSdk != "" {
				// Azure uses UUIDs, just set to a consistent value
				req.Header.Set("x-ms-client-request-id", "test-request-id")
			}

			return nil
		},
	}
}

// getDefaultResponseScrubbers returns the default response scrubbing functions to remove
// sensitive headers and tracing information from response recordings.
func getDefaultResponseScrubbers() []func(*bytes.Buffer) error {
	return []func(*bytes.Buffer) error{
		func(buf *bytes.Buffer) error {
			// Parse the response from the buffer
			resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(buf.Bytes())), nil)
			if err != nil {
				return nil // Ignore parse errors, just return the buffer as-is
			}

			// Remove Cf-Ray header (Cloudflare tracing)
			resp.Header.Del("Cf-Ray")
			resp.Header.Del("cf-ray")

			// Remove Set-Cookie headers (session tokens, etc.)
			resp.Header.Del("Set-Cookie")
			resp.Header.Del("set-cookie")

			// Munge Openai-Organization header in responses too
			if resp.Header.Get("Openai-Organization") != "" {
				resp.Header.Set("Openai-Organization", "lcgo-tst")
			}
			if resp.Header.Get("openai-organization") != "" {
				resp.Header.Set("openai-organization", "lcgo-tst")
			}

			// Re-serialize the response
			buf.Reset()
			if err := resp.Write(buf); err != nil {
				return err
			}

			return nil
		},
	}
}

// conditionalEmbeddingFormatter returns a scrubber that only applies embedding
// formatting to responses from specific embedding endpoints.
func conditionalEmbeddingFormatter() func(*bytes.Buffer) error {
	embeddingFormatter := EmbeddingJSONFormatter()
	return func(buf *bytes.Buffer) error {
		content := buf.String()

		// Only apply to known embedding endpoints (be very specific)
		if strings.Contains(content, "POST https://api.openai.com/v1/embeddings") ||
			strings.Contains(content, "batchEmbedContents") ||
			strings.Contains(content, "models/embedding-") {
			return embeddingFormatter(buf)
		}

		return nil // Not an embedding endpoint, skip formatting
	}
}

// EmbeddingJSONFormatter returns a response scrubber that formats JSON responses
// with special handling for number arrays (displays them on single lines).
// This is particularly useful for embedding API responses which often contain
// large arrays of floating-point numbers.
//
// Usage in tests:
//
//	rr.ScrubResp(httprr.EmbeddingJSONFormatter())
func EmbeddingJSONFormatter() func(*bytes.Buffer) error {
	return func(buf *bytes.Buffer) error {
		// Parse the response to get headers and body
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(buf.Bytes())), nil)
		if err != nil {
			return nil // Not an HTTP response, skip formatting
		}

		// Read the body
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil // Can't read body, skip formatting
		}

		// Check if content-type suggests JSON
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return nil // Not JSON, skip formatting
		}

		// Try to format the JSON body
		formattedBody := formatJSONBody(bodyBytes)

		// Reconstruct the response with formatted body
		resp.Body = io.NopCloser(bytes.NewReader(formattedBody))
		resp.ContentLength = int64(len(formattedBody))

		// Re-serialize the response
		buf.Reset()
		if err := resp.Write(buf); err != nil {
			return err
		}

		return nil
	}
}

// formatJSONBody formats JSON with special handling for number arrays
func formatJSONBody(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	// Try to detect if this might be JSON
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return data
	}

	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return data // Not valid JSON, return original
	}

	// Format with custom handling for number arrays
	formatted := formatJSONValue(jsonData, 0)
	return []byte(formatted)
}

// formatJSONValue formats a JSON value with special handling for arrays of numbers
func formatJSONValue(v interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	nextIndentStr := strings.Repeat("  ", indent+1)

	switch val := v.(type) {
	case map[string]interface{}:
		if len(val) == 0 {
			return "{}"
		}
		var parts []string
		parts = append(parts, "{")

		// Preserve original key order by iterating over the map directly
		i := 0
		for k, v := range val {
			formatted := formatJSONValue(v, indent+1)
			parts = append(parts, fmt.Sprintf("%s\"%s\": %s", nextIndentStr, k, formatted))
			if i < len(val)-1 {
				parts[len(parts)-1] += ","
			}
			i++
		}
		parts = append(parts, indentStr+"}")
		return strings.Join(parts, "\n")

	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}

		// Check if this is an array of numbers
		allNumbers := true
		for _, item := range val {
			switch item.(type) {
			case float64, int, int64:
				// continue
			default:
				allNumbers = false
				break
			}
		}

		if allNumbers && len(val) > 0 {
			// Format number arrays on a single line
			var nums []string
			for _, item := range val {
				switch n := item.(type) {
				case float64:
					// Format float64 with appropriate precision
					if n == float64(int64(n)) {
						nums = append(nums, fmt.Sprintf("%d", int64(n)))
					} else {
						nums = append(nums, fmt.Sprintf("%g", n))
					}
				default:
					nums = append(nums, fmt.Sprintf("%v", item))
				}
			}
			return "[" + strings.Join(nums, ", ") + "]"
		}

		// Format other arrays with one item per line
		var parts []string
		parts = append(parts, "[")
		for i, item := range val {
			formatted := formatJSONValue(item, indent+1)
			parts = append(parts, nextIndentStr+formatted)
			if i < len(val)-1 {
				parts[len(parts)-1] += ","
			}
		}
		parts = append(parts, indentStr+"]")
		return strings.Join(parts, "\n")

	case string:
		// Marshal string to get proper escaping
		b, _ := json.Marshal(val)
		return string(b)

	case float64:
		// Format float64 with appropriate precision
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)

	case bool:
		return fmt.Sprintf("%v", val)

	case nil:
		return "null"

	default:
		// Fallback to standard JSON marshaling
		b, _ := json.Marshal(val)
		return string(b)
	}
}
