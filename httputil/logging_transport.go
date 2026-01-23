package httputil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
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

// JSONDebugClient is an [http.Client] designed for debugging JSON APIs and Server-Sent Events.
// It provides comprehensive debugging output including HTTP headers, JSON payloads, and real-time
// SSE event parsing. All debug output is written to stderr with ANSI colors:
// requests in blue, responses in green, SSE events in green, and parsed data in purple/yellow.
//
// Key features:
//   - Pretty-prints JSON request and response bodies
//   - Displays HTTP headers with sensitive values scrubbed
//   - Streams SSE events in real-time as they arrive
//   - Parses token usage from streaming APIs
//
// Unlike [LoggingClient], this client writes directly to stderr rather than
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
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorPurple = "\033[35m"
	colorGrey   = "\033[90m"
	colorReset  = "\033[0m"
)

type jsonDebugTransport struct {
	Transport http.RoundTripper
}

func scrubSensitive(key, value string) string {
	key = strings.ToLower(key)
	if strings.Contains(key, "auth") || strings.Contains(key, "key") || strings.Contains(key, "token") ||
		strings.Contains(key, "cookie") || strings.Contains(key, "secret") {
		if len(value) > 8 {
			return value[:4] + "..." + value[len(value)-4:]
		}
		return "***"
	}
	return value
}

func (t *jsonDebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Print request headers
	fmt.Fprintf(os.Stderr, "%sRequest %s %s%s\n", colorBlue, req.Method, req.URL, colorReset)
	fmt.Fprintf(os.Stderr, "%sRequest Headers:%s\n", colorGrey, colorReset)
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "%s  %s: %s%s\n", colorGrey, key, scrubSensitive(key, value), colorReset)
		}
	}

	if err := t.logJSON(req.Header.Get("Content-Type"), &req.Body, colorBlue, "Body:"); err != nil {
		return nil, err
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Print response headers
	fmt.Fprintf(os.Stderr, "%sResponse %d %s%s\n", colorGreen, resp.StatusCode, resp.Status, colorReset)
	fmt.Fprintf(os.Stderr, "%sResponse Headers:%s\n", colorGrey, colorReset)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "%s  %s: %s%s\n", colorGrey, key, scrubSensitive(key, value), colorReset)
		}
	}

	// Handle SSE streams
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") && resp.Body != nil {
		return t.wrapSSEResponse(resp), nil
	}

	if err := t.logJSON(contentType, &resp.Body, colorGreen, "Body:"); err != nil {
		return resp, err
	}

	return resp, nil
}

func (t *jsonDebugTransport) wrapSSEResponse(resp *http.Response) *http.Response {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer resp.Body.Close()

		fmt.Fprintf(os.Stderr, "%sSSE stream starting...%s\n", colorGreen, colorReset)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Print raw SSE frame
			fmt.Fprintf(os.Stderr, "%s%s%s\n", colorGreen, line, colorReset)

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if !strings.Contains(data, `"type": "ping"`) {
					t.parseEvent(data)
				}
			}

			// Write to pipe for normal consumption
			if _, err := pw.Write(append([]byte(line), '\n')); err != nil {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "%sSSE stream error: %v%s\n", colorGreen, err, colorReset)
		}

		fmt.Fprintf(os.Stderr, "%sSSE stream ended.%s\n", colorGreen, colorReset)
	}()

	newResp := *resp
	newResp.Body = pr
	return &newResp
}

func (t *jsonDebugTransport) parseEvent(text string) {
	var data map[string]interface{}
	if json.Unmarshal([]byte(text), &data) != nil {
		return
	}

	switch data["type"] {
	case "message_start":
		if msg, ok := data["message"].(map[string]interface{}); ok {
			if usage, ok := msg["usage"].(map[string]interface{}); ok {
				fmt.Fprintf(os.Stderr, "%s[message_start] Usage: ", colorPurple)
				t.printUsage(usage)
				fmt.Fprintf(os.Stderr, "%s\n", colorReset)
			}
		}
	case "message_delta":
		if usage, ok := data["usage"].(map[string]interface{}); ok {
			fmt.Fprintf(os.Stderr, "%s[message_delta] Final usage: ", colorPurple)
			t.printUsage(usage)
			fmt.Fprintf(os.Stderr, "%s\n", colorReset)
		}
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			if sig := delta["signature"]; sig != nil {
				fmt.Fprintf(os.Stderr, "%s[signature] %+v%s\n", colorYellow, sig, colorReset)
			}
		}
	case "content_block_delta":
		if delta, ok := data["delta"].(map[string]interface{}); ok {
			if delta["type"] == "thinking_delta" {
				if text, ok := delta["text"].(string); ok {
					fmt.Fprintf(os.Stderr, "%s[thinking] %s%s\n", colorYellow, text, colorReset)
				}
			}
		}
	}
}

func (t *jsonDebugTransport) logJSON(contentType string, body *io.ReadCloser, color, label string) error {
	if !strings.Contains(contentType, "application/json") || *body == nil {
		return nil
	}
	data, err := io.ReadAll(*body)
	if err != nil {
		return err
	}
	*body = io.NopCloser(bytes.NewReader(data))

	var pretty bytes.Buffer
	if json.Indent(&pretty, data, "", "  ") == nil {
		fmt.Fprintf(os.Stderr, "%s%s%s\n%s%s%s\n", color, label, colorReset, color, pretty.String(), colorReset)
	}
	return nil
}

func (t *jsonDebugTransport) printUsage(usage map[string]interface{}) {
	if n, ok := usage["input_tokens"].(float64); ok {
		fmt.Fprintf(os.Stderr, "input=%d ", int(n))
	}
	if n, ok := usage["output_tokens"].(float64); ok {
		fmt.Fprintf(os.Stderr, "output=%d ", int(n))
	}
	if n, ok := usage["cache_creation_input_tokens"].(float64); ok {
		fmt.Fprintf(os.Stderr, "cache_create=%d ", int(n))
	}
	if n, ok := usage["cache_read_input_tokens"].(float64); ok {
		fmt.Fprintf(os.Stderr, "cache_read=%d ", int(n))
	}
}
