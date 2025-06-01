// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httprr implements HTTP record and replay, mainly for use in tests.
//
// [Open] creates a new [RecordReplay]. Whether it is recording or replaying
// is controlled by the -httprecord flag, which is defined by this package
// only in test programs (built by “go test”).
// See the [Open] documentation for more details.
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
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	record      = new(string)
	recordDelay = new(time.Duration)
)

func init() {
	if testing.Testing() {
		record = flag.String("httprecord", "", "re-record traces for files matching `regexp`")
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
	reqWire, err := rr.reqWire(req)
	if err != nil {
		return nil, err
	}

	// If we're in replay mode, replay a response.
	if rr.replay != nil {
		return rr.replayRoundTrip(req, reqWire)
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

	// Add delay after request when recording (helps avoid rate limits)
	if rr.Recording() {
		delay := *recordDelay
		if delay == 0 {
			// Default to 1 second delay when recording
			delay = 1 * time.Second
		}
		time.Sleep(delay)
	}

	// Encode resp and decode to get a copy for our caller.
	respWire, err := rr.respWire(resp)
	if err != nil {
		return nil, err
	}
	if err := rr.writeLog(reqWire, respWire); err != nil {
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
	for _, scrub := range rr.reqScrub {
		if err := scrub(rkey); err != nil {
			return "", err
		}
	}

	// Now that scrubbers are done potentially modifying body, set length.
	if rkey.Body != nil {
		rkey.ContentLength = int64(len(rkey.Body.(*Body).Data))
	}

	// Serialize rkey to produce the log entry.
	// Use WriteProxy instead of Write to preserve the URL's scheme.
	var key strings.Builder
	if err := rkey.WriteProxy(&key); err != nil {
		return "", err
	}
	return key.String(), nil
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
	respLog, ok := rr.replay[reqLog]
	if !ok {
		return nil, fmt.Errorf("cached HTTP response not found for:\n%s\n\nHint: Re-run tests with -httprecord=. to record new HTTP interactions", reqLog)
	}
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(respLog)), req)
	if err != nil {
		return nil, fmt.Errorf("read %s: corrupt httprr trace: %w", rr.file, err)
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

// OpenForTest creates a [RecordReplay] for the given test using a filename
// derived from the test name. The recording will be stored in a "testdata"
// subdirectory with a ".httprr" extension.
//
// Example usage:
//
//	func TestMyAPI(t *testing.T) {
//	    rr, err := httprr.OpenForTest(t, http.DefaultTransport)
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    defer rr.Close()
//
//	    client := rr.Client()
//	    // use client for HTTP requests...
//	}
//
// This will create/use a file at "testdata/TestMyAPI.httprr".
func OpenForTest(t *testing.T, rt http.RoundTripper) *RecordReplay {
	t.Helper()
	testName := CleanFileName(t.Name())
	filename := filepath.Join("testdata", testName+".httprr")

	// Ensure testdata directory exists
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatalf("httprr: failed to create testdata directory: %v", err)
	}

	// Check if we're in recording mode
	recording, err := Recording(filename)
	if err != nil {
		t.Fatal(err)
	}

	if recording {
		// Recording mode: clean up existing files and create uncompressed
		cleanupExistingFiles(t, filename)
		rr, err := Open(filename, rt)
		if err != nil {
			t.Fatalf("httprr: failed to open recording file %s: %v", filename, err)
		}
		t.Cleanup(func() { rr.Close() })
		return rr
	}

	// Replay mode: find the best existing file
	filename = findBestReplayFile(t, filename)
	rr, err := Open(filename, rt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rr.Close() })
	return rr
}

// cleanupExistingFiles removes any existing files to avoid conflicts during recording
func cleanupExistingFiles(t *testing.T, baseFilename string) {
	t.Helper()
	filesToCheck := []string{baseFilename, baseFilename + ".gz"}

	for _, filename := range filesToCheck {
		if _, err := os.Stat(filename); err == nil {
			t.Logf("httprr: removing existing file %s", filename)
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
	if !hasRequiredCredentials(envVars) && !hasExistingRecording(t) {
		skipMessage := "no httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions"

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

// ConvertAllInDir processes all files in a directory, converting between
// compressed and uncompressed formats based on the compress parameter.
func ConvertAllInDir(dir string, compress bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if compress {
			if strings.HasSuffix(path, ".httprr") && !strings.HasSuffix(path, ".httprr.gz") {
				if err := CompressFile(path); err != nil {
					return fmt.Errorf("compressing %s: %w", path, err)
				}
			}
		} else {
			if strings.HasSuffix(path, ".httprr.gz") {
				if err := DecompressFile(path); err != nil {
					return fmt.Errorf("decompressing %s: %w", path, err)
				}
			}
		}
	}

	return nil
}

// CompressFile compresses a file to .httprr.gz format.
func CompressFile(path string) error {
	if !strings.HasSuffix(path, ".httprr") || strings.HasSuffix(path, ".httprr.gz") {
		return fmt.Errorf("invalid file for compression: %s", path)
	}

	// Read the original file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Write compressed file
	compressedPath := path + ".gz"
	f, err := os.Create(compressedPath)
	if err != nil {
		return fmt.Errorf("creating compressed file: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	if _, err := gz.Write(data); err != nil {
		os.Remove(compressedPath)
		return fmt.Errorf("writing compressed data: %w", err)
	}

	if err := gz.Close(); err != nil {
		os.Remove(compressedPath)
		return fmt.Errorf("finalizing compressed file: %w", err)
	}

	// Remove original file
	if err := os.Remove(path); err != nil {
		// Try to clean up the compressed file if we can't remove the original
		os.Remove(compressedPath)
		return fmt.Errorf("removing original file: %w", err)
	}

	return nil
}

// DecompressFile decompresses an .httprr.gz file to .httprr format.
func DecompressFile(path string) error {
	if !strings.HasSuffix(path, ".httprr.gz") {
		return fmt.Errorf("invalid file for decompression: %s", path)
	}

	// Open the compressed file
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening compressed file: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}
	defer gz.Close()

	// Read decompressed data
	data, err := io.ReadAll(gz)
	if err != nil {
		return fmt.Errorf("reading compressed data: %w", err)
	}

	// Write decompressed file
	decompressedPath := strings.TrimSuffix(path, ".gz")
	if err := os.WriteFile(decompressedPath, data, 0o644); err != nil {
		return fmt.Errorf("writing decompressed file: %w", err)
	}

	// Remove compressed file
	if err := os.Remove(path); err != nil {
		// Try to clean up the decompressed file if we can't remove the compressed one
		os.Remove(decompressedPath)
		return fmt.Errorf("removing compressed file: %w", err)
	}

	return nil
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

			// Munge Openai-Organization header to a test value
			if req.Header.Get("Openai-Organization") != "" {
				req.Header.Set("Openai-Organization", "lcgo-tst")
			}
			if req.Header.Get("openai-organization") != "" {
				req.Header.Set("openai-organization", "lcgo-tst")
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
