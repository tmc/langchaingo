package googleaierrors

import (
	"net/http"
	"strings"

	"github.com/vendasta/langchaingo/llms"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errorMapping represents a mapping from error patterns to error codes.
type errorMapping struct {
	patterns []string
	code     llms.ErrorCode
	message  string
}

// googleAIErrorMappings defines the error mappings for Google AI.
var googleAIErrorMappings = []errorMapping{
	{
		patterns: []string{"invalid api key", "api key not valid", "401"},
		code:     llms.ErrCodeAuthentication,
		message:  "Invalid or missing API key",
	},
	{
		patterns: []string{"quota exceeded", "rate limit", "429"},
		code:     llms.ErrCodeRateLimit,
		message:  "Rate limit or quota exceeded",
	},
	{
		patterns: []string{"model not found", "404"},
		code:     llms.ErrCodeResourceNotFound,
		message:  "Model not found",
	},
	{
		patterns: []string{"token limit", "context length"},
		code:     llms.ErrCodeTokenLimit,
		message:  "Token limit exceeded",
	},
	{
		patterns: []string{"harmful content", "safety"},
		code:     llms.ErrCodeContentFilter,
		message:  "Content blocked by safety filter",
	},
	{
		patterns: []string{"billing", "payment required"},
		code:     llms.ErrCodeQuotaExceeded,
		message:  "Billing quota exceeded",
	},
	{
		patterns: []string{"invalid request", "400"},
		code:     llms.ErrCodeInvalidRequest,
		message:  "Invalid request",
	},
	{
		patterns: []string{"service unavailable", "503"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "Google AI service temporarily unavailable",
	},
}

// MapError maps Google AI-specific errors to standardized error codes.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	var statusCode int
	errStatus, ok := status.FromError(err)
	if ok {
		statusCode = grpcErrorToHTTPStatusCode(errStatus.Code())
	}

	errStr := strings.ToLower(err.Error())

	// Check each error mapping
	for _, mapping := range googleAIErrorMappings {
		for _, pattern := range mapping.patterns {
			if strings.Contains(errStr, pattern) {
				return llms.NewError(mapping.code, "googleai", mapping.message, statusCode).WithCause(err)
			}
		}
	}

	// Use the generic error mapper for unrecognized errors
	mapper := llms.NewErrorMapper("googleai")
	return mapper.Map(err)
}

// StatusCodeToGRPCError converts a http error into a grpc error
func grpcErrorToHTTPStatusCode(statusCode codes.Code) int {
	switch statusCode {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.Internal:
		return http.StatusInternalServerError
	default:
		return 0
	}
}
