package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
)

// LoggingClient is an [http.Client] that logs complete HTTP requests and responses
// using structured logging via [slog]. This client is useful for debugging API
// interactions, as it captures the full request and response including headers
// and bodies. The logs are emitted at DEBUG level.
//
// Example:
//
//	slog.SetLogLoggerLevel(slog.LevelDebug)
//	resp, err := httputil.LoggingClient.Get("https://api.example.com/data")
var LoggingClient = &http.Client{ //nolint:gochecknoglobals
	Transport: &Transport{
		Transport: &LoggingTransport{},
	},
}

// JSONDebugClient is an [http.Client] designed for debugging JSON APIs.
// It pretty-prints JSON request and response bodies to stdout with ANSI colors:
// requests are shown in blue, responses in green. This client is intended for
// development and debugging purposes only.
//
// Unlike [LoggingClient], this client writes directly to stdout rather than
// using structured logging.
var JSONDebugClient = &http.Client{ //nolint:gochecknoglobals
	Transport: &Transport{
		Transport: &jsonDebugTransport{},
	},
}

// DebugHTTPClient is a deprecated alias for [LoggingClient].
//
// Deprecated: Use [LoggingClient] instead.
var DebugHTTPClient = LoggingClient

// DebugHTTPColorJSON is a deprecated alias for [JSONDebugClient].
//
// Deprecated: Use [JSONDebugClient] instead.
var DebugHTTPColorJSON = JSONDebugClient //nolint:gochecknoglobals

// LoggingTransport is an [http.RoundTripper] that logs complete HTTP requests
// and responses using structured logging. It's designed for debugging and
// development purposes.
//
// The transport logs at DEBUG level, so ensure your logger is configured
// appropriately to see the output.
type LoggingTransport struct {
	// Transport is the underlying [http.RoundTripper] to use.
	// If nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper

	// Logger is the [slog.Logger] to use for logging.
	// If nil, [slog.Default] is used.
	Logger *slog.Logger
}

// RoundTrip implements the [http.RoundTripper] interface. It logs the complete
// HTTP request (including headers and body) before sending it, executes the
// request using the underlying transport, then logs the complete response.
// Both request and response are logged at DEBUG level.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := t.Logger
	if logger == nil {
		logger = slog.Default()
	}
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	// Dump the request
	requestDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Error("Failed to dump request", "error", err)
	} else {
		logger.Debug(string(requestDump))
	}
	// Perform the actual request
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("failed to round trip request: %w", err)
	}
	// Dump the response
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logger.Error("Failed to dump response", "error", err)
	} else {
		logger.Debug(string(responseDump))
	}
	return resp, nil
}

// ANSI color codes
const (
	colorBlue  = "\033[34m"
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

type jsonDebugTransport struct {
	Transport http.RoundTripper
}

func (t *jsonDebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Log JSON request if present
	if strings.Contains(req.Header.Get("Content-Type"), "application/json") && req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))

		// Pretty print request in blue
		var pretty bytes.Buffer
		if json.Indent(&pretty, body, "", "  ") == nil {
			fmt.Printf("%sRequest to %s\n%s%s\n", colorBlue, req.URL, pretty.String(), colorReset)
		}
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log JSON response if present
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") && resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = io.NopCloser(bytes.NewReader(body))

		// Pretty print response in green
		var pretty bytes.Buffer
		if json.Indent(&pretty, body, "", "  ") == nil {
			fmt.Printf("%sResponse %d\n%s%s\n", colorGreen, resp.StatusCode, pretty.String(), colorReset)
		}
	}

	return resp, nil
}
