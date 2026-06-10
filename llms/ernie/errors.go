package ernie

import (
	"errors"
	"strings"

	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/llms"
)

// errorMapping represents a mapping from error patterns to error codes.
type errorMapping struct {
	patterns []string
	code     llms.ErrorCode
	message  string
}

// ernieErrorMappings defines the error mappings for ERNIE.
// ERNIE returns errors both as HTTP status codes and as body-level error_code fields.
// Mappings are checked in order; first match wins.
//
// Reference: https://ai.baidu.com/ai-doc/NLP/Bk6z52e59
var ernieErrorMappings = []errorMapping{
	{
		// error_code:18 = QPS rate limit (per-second) — retryable
		patterns: []string{"error_code:18", "qps limit"},
		code:     llms.ErrCodeRateLimit,
		message:  "ERNIE QPS limit exceeded",
	},
	{
		patterns: []string{"error_code:110", "error_code:111", "access token"},
		code:     llms.ErrCodeAuthentication,
		message:  "ERNIE authentication failed",
	},
	{
		// ERNIE invalid parameter errors
		patterns: []string{"error_code:1", "error_code:2", "error_code:3", "invalid parameter"},
		code:     llms.ErrCodeInvalidRequest,
		message:  "Invalid request parameter",
	},
	{
		// error_code:17 = daily request limit (per-day) — NOT retryable
		// error_code:19 = total request limit — NOT retryable
		patterns: []string{"error_code:17", "error_code:19", "quota", "limit exceeded"},
		code:     llms.ErrCodeQuotaExceeded,
		message:  "API quota exceeded",
	},
	{
		// Server-side errors
		patterns: []string{"error_code:500", "error_code:503", "internal error", "service unavailable"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "ERNIE service temporarily unavailable",
	},
}

// MapError maps ERNIE-specific errors to standardized error codes.
// It handles both HTTP status code errors and body-level error_code fields
// that ERNIE returns with HTTP 200 responses.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	for _, mapping := range ernieErrorMappings {
		for _, pattern := range mapping.patterns {
			if strings.Contains(errStr, strings.ToLower(pattern)) {
				classified := llms.NewError(mapping.code, "ernie", mapping.message).WithCause(err)
				transferRetryAfter(err, classified)
				return classified
			}
		}
	}

	// Fall back to generic error mapper.
	mapper := llms.NewErrorMapper("ernie")
	return mapper.Map(err)
}

// transferRetryAfter extracts the Retry-After value from an *httputil.ResponseError
// and sets it on the classified *llms.Error.
func transferRetryAfter(src error, dst *llms.Error) {
	var respErr *httputil.ResponseError
	if errors.As(src, &respErr) && respErr.RetryAfter > 0 {
		dst.WithRetryAfter(respErr.RetryAfter)
	}
}
