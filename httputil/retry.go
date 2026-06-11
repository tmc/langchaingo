package httputil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryConfig controls the retry behavior for HTTP requests.
//
// Retry is performed at the transport layer via [RetryTransport], which wraps
// an underlying [http.RoundTripper]. Only transient failures are retried:
//
//   - HTTP 429 (Rate Limit) — respects the Retry-After header
//   - HTTP 500, 502, 503, 504 (Server Errors)
//   - Network errors (connection refused, DNS failure, TLS timeout)
//
// Client errors (4xx except 429), context cancellation, and successful
// responses are never retried.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	// The total number of requests will be MaxRetries + 1.
	// Default: 3
	MaxRetries int

	// InitialBackoff is the duration to wait before the first retry.
	// Default: 1 * time.Second
	InitialBackoff time.Duration

	// MaxBackoff is the upper bound on backoff duration.
	// Default: 30 * time.Second
	MaxBackoff time.Duration

	// BackoffFactor is the multiplier applied to the backoff after each attempt.
	// Default: 2.0
	BackoffFactor float64

	// RetryableError, if provided, overrides the default error retryability check.
	// Return true to indicate the error is transient and should be retried.
	// The default check retries network errors (connection refused, DNS, TLS).
	RetryableError func(error) bool

	// RetryableStatus, if provided, overrides the default HTTP status code
	// retryability check. Return true to indicate the status code is transient.
	// The default check retries 429, 500, 502, 503, 504.
	RetryableStatus func(statusCode int) bool

	// OnRetry is called before each retry attempt with the attempt number
	// (0-based) and the error that triggered the retry. Use this for logging
	// or metrics.
	OnRetry func(attempt int, err error)
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults:
//   - MaxRetries: 3
//   - InitialBackoff: 1 second
//   - MaxBackoff: 30 seconds
//   - BackoffFactor: 2.0
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

// RetryTransport is an [http.RoundTripper] that automatically retries
// failed requests using exponential backoff with jitter.
//
// RetryTransport composes with the existing [Transport] (User-Agent injection)
// and any other [http.RoundTripper]. Typical transport chain:
//
//	RetryTransport → Transport → http.DefaultTransport
//
// Retry is only performed before a successful response is received.
// Once a 200 OK is returned (including for streaming/SSE responses),
// no mid-stream retry is attempted.
type RetryTransport struct {
	// Transport is the underlying [http.RoundTripper].
	// If nil, [http.DefaultTransport] is used.
	Transport http.RoundTripper

	// Config controls retry behavior. Must not be nil.
	Config *RetryConfig
}

// RoundTrip implements the [http.RoundTripper] interface.
// It retries the request according to the configured [RetryConfig].
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cfg := t.Config
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}

	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	ctx := req.Context()

	// Cache the request body for potential replays.
	bodyBytes, err := readBody(req)
	if err != nil {
		return nil, fmt.Errorf("retry: failed to read request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check if the context has been cancelled.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Clone the request, restoring the body.
		reqClone := cloneRequest(req, bodyBytes)

		resp, err := transport.RoundTrip(reqClone)
		if err != nil {
			lastErr = err
			if isRetryableError(err, cfg) && attempt < cfg.MaxRetries {
				fireOnRetry(cfg, attempt, err)
				if !wait(ctx, cfg.backoff(attempt)) {
					return nil, ctx.Err()
				}
				continue
			}
			return nil, err
		}

		// Check if the status code is retryable.
		if isRetryableStatus(resp.StatusCode, cfg) && attempt < cfg.MaxRetries {
			lastErr = fmt.Errorf("server returned %d: %s", resp.StatusCode, resp.Status)
			// Parse Retry-After header for 429 responses.
			waitDuration := cfg.backoff(attempt)
			if ra, ok := parseRetryAfter(resp); ok && ra > waitDuration {
				waitDuration = ra
			}
			resp.Body.Close()
			fireOnRetry(cfg, attempt, lastErr)
			if !wait(ctx, waitDuration) {
				return nil, ctx.Err()
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("retry: max retries (%d) exceeded, last error: %w", cfg.MaxRetries, lastErr)
}

// backoff calculates the wait duration for the given attempt using exponential
// backoff with random jitter.
func (c *RetryConfig) backoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.InitialBackoff
	}

	backoff := float64(c.InitialBackoff) * math.Pow(c.BackoffFactor, float64(attempt))

	// Cap at MaxBackoff, then apply jitter.
	if backoff > float64(c.MaxBackoff) {
		backoff = float64(c.MaxBackoff)
	}

	// Add random jitter in [0.5, 1.0) to ensure we never exceed MaxBackoff.
	jitter := 0.5 + rand.Float64()*0.5 //nolint:gosec // G404: jitter doesn't need crypto-rand
	return time.Duration(backoff * jitter)
}

// isRetryableError checks if the given error is transient and worth retrying.
func isRetryableError(err error, cfg *RetryConfig) bool {
	if cfg.RetryableError != nil {
		return cfg.RetryableError(err)
	}
	return defaultIsRetryableError(err)
}

// defaultIsRetryableError checks for network-level transient errors.
func defaultIsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are not retryable — the caller cancelled the request.
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Network errors (connection refused, DNS, TLS) are generally transient.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Check for common network error patterns via string matching.
	// Many HTTP client errors wrap underlying network errors without
	// implementing the net.Error interface.
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"TLS handshake",
		"no such host",
		"temporary",
		"i/o timeout",
		"EOF",
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// isRetryableStatus checks if the HTTP status code indicates a transient failure.
func isRetryableStatus(statusCode int, cfg *RetryConfig) bool {
	if cfg.RetryableStatus != nil {
		return cfg.RetryableStatus(statusCode)
	}
	return defaultIsRetryableStatus(statusCode)
}

// defaultIsRetryableStatus retries on 429 (rate limit) and 5xx (server errors).
func defaultIsRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// parseRetryAfter parses the Retry-After header from the response.
// It supports both seconds (integer) and HTTP date formats.
func parseRetryAfter(resp *http.Response) (time.Duration, bool) {
	val := resp.Header.Get("Retry-After")
	if val == "" {
		return 0, false
	}

	// Try parsing as integer seconds.
	if seconds, err := strconv.Atoi(val); err == nil {
		return time.Duration(seconds) * time.Second, true
	}

	// Try parsing as HTTP date.
	if t, err := http.ParseTime(val); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d, true
		}
	}

	return 0, false
}

// readBody reads and caches the request body so it can be replayed on retry.
func readBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	data, err := io.ReadAll(req.Body)
	req.Body.Close()
	return data, err
}

// cloneRequest creates a shallow clone of the request with a fresh body.
func cloneRequest(req *http.Request, body []byte) *http.Request {
	r := req.Clone(req.Context())
	if body != nil {
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
		r.ContentLength = int64(len(body))
	}
	return r
}

// wait blocks for the given duration or until the context is cancelled.
// Returns false if the context was cancelled.
func wait(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// fireOnRetry calls the OnRetry callback if configured.
func fireOnRetry(cfg *RetryConfig, attempt int, err error) {
	if cfg.OnRetry != nil {
		cfg.OnRetry(attempt, err)
	}
}
