package httputil

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ResponseError represents an HTTP error response that carries metadata
// useful for retry decisions, including the Retry-After header value.
//
// Provider internal clients should return *ResponseError when they receive
// non-200 HTTP responses so that [RetryOnError] can respect Retry-After.
type ResponseError struct {
	// StatusCode is the HTTP status code.
	StatusCode int

	// Message is the error message (typically includes status code and API body).
	Message string

	// RetryAfter is the duration indicated by the Retry-After response header.
	// Zero means the header was absent or could not be parsed.
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e *ResponseError) Error() string { return e.Message }

// ParseRetryAfterHeader extracts the Retry-After duration from an HTTP response.
// Returns 0 if the header is absent or cannot be parsed.
func ParseRetryAfterHeader(resp *http.Response) time.Duration {
	val := resp.Header.Get("Retry-After")
	if val == "" {
		return 0
	}

	// Try parsing as integer seconds.
	if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date.
	if t, err := http.ParseTime(val); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}

	return 0
}

// NewResponseError creates a ResponseError from an HTTP response.
// The body should be pre-read; this function only extracts the status code
// and Retry-After header.
func NewResponseError(resp *http.Response, message string) *ResponseError {
	return &ResponseError{
		StatusCode: resp.StatusCode,
		Message:    message,
		RetryAfter: ParseRetryAfterHeader(resp),
	}
}

// FormatHTTPError formats a standard HTTP error message with status code and optional body.
func FormatHTTPError(statusCode int, body string) string {
	if body != "" {
		return fmt.Sprintf("API returned unexpected status code: %d: %s", statusCode, body)
	}
	return fmt.Sprintf("API returned unexpected status code: %d", statusCode)
}
