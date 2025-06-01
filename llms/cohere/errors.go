package cohere

import (
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// errorMapping represents a mapping from error patterns to error codes.
type errorMapping struct {
	patterns []string
	code     llms.ErrorCode
	message  string
}

// cohereErrorMappings defines the error mappings for Cohere.
var cohereErrorMappings = []errorMapping{
	{
		patterns: []string{"invalid api key", "unauthorized", "401"},
		code:     llms.ErrCodeAuthentication,
		message:  "Invalid or missing API key",
	},
	{
		patterns: []string{"rate limit", "too many requests", "429"},
		code:     llms.ErrCodeRateLimit,
		message:  "Rate limit exceeded",
	},
	{
		patterns: []string{"model not found", "invalid model"},
		code:     llms.ErrCodeResourceNotFound,
		message:  "Model not found",
	},
	{
		patterns: []string{"context length", "too many tokens"},
		code:     llms.ErrCodeTokenLimit,
		message:  "Token limit exceeded",
	},
	{
		patterns: []string{"invalid request", "400"},
		code:     llms.ErrCodeInvalidRequest,
		message:  "Invalid request",
	},
	{
		patterns: []string{"quota exceeded", "billing"},
		code:     llms.ErrCodeQuotaExceeded,
		message:  "API quota exceeded",
	},
	{
		patterns: []string{"service unavailable", "503"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "Cohere service temporarily unavailable",
	},
	{
		patterns: []string{"internal error", "500"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "Cohere service error",
	},
}

// MapError maps Cohere-specific errors to standardized error codes.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check each error mapping
	for _, mapping := range cohereErrorMappings {
		for _, pattern := range mapping.patterns {
			if strings.Contains(errStr, pattern) {
				return llms.NewError(mapping.code, "cohere", mapping.message).WithCause(err)
			}
		}
	}

	// Use the generic error mapper for unrecognized errors
	mapper := llms.NewErrorMapper("cohere")
	return mapper.Map(err)
}
