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

// LoggingClient is an [http.Client] that logs the request and response with full contents to the default [slog.Logger].
var LoggingClient = &http.Client{ //nolint:gochecknoglobals
	Transport: &Transport{
		Transport: &LoggingTransport{},
	},
}

// JSONDebugClient is an HTTP client that pretty-prints JSON request/response bodies with color.
// Use this for debugging JSON APIs during development.
var JSONDebugClient = &http.Client{ //nolint:gochecknoglobals
	Transport: &Transport{
		Transport: &jsonDebugTransport{},
	},
}

// Deprecated: Use [LoggingClient] instead.
var DebugHTTPClient = LoggingClient

// Deprecated: Use JSONDebugClient instead.
var DebugHTTPColorJSON = JSONDebugClient //nolint:gochecknoglobals

// LoggingTransport is an [http.RoundTripper] that logs the request and response with full contents.
// It uses [http.DefaultTransport] if the Transport field is nil, and the default [slog.Logger] if the Logger field is nil.
type LoggingTransport struct {
	Transport http.RoundTripper
	Logger    *slog.Logger
}

// RoundTrip implements the [http.RoundTripper] interface and logs the request and response to the Logger field.
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