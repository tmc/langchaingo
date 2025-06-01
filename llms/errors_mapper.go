package llms

import (
	"context"
	"errors"
	"net"
	"strings"
)

// ErrorMapper helps map provider-specific errors to standardized errors.
type ErrorMapper struct {
	provider string
	matchers []ErrorMatcher
}

// ErrorMatcher matches an error and returns the appropriate error code.
type ErrorMatcher struct {
	// Match returns true if this matcher handles the error
	Match func(error) bool
	// Code is the error code to use
	Code ErrorCode
	// Transform optionally transforms the error message
	Transform func(error) string
}

// NewErrorMapper creates a new error mapper for a provider.
func NewErrorMapper(provider string) *ErrorMapper {
	return &ErrorMapper{
		provider: provider,
		matchers: defaultMatchers(),
	}
}

// defaultMatchers returns the default set of error matchers.
func defaultMatchers() []ErrorMatcher {
	matchers := []ErrorMatcher{}

	// Add context error matchers
	matchers = append(matchers, contextErrorMatchers()...)

	// Add string pattern matchers
	matchers = append(matchers, stringPatternMatchers()...)

	return matchers
}

// contextErrorMatchers returns matchers for context-related errors.
func contextErrorMatchers() []ErrorMatcher {
	return []ErrorMatcher{
		{
			Match: func(err error) bool {
				return errors.Is(err, context.Canceled)
			},
			Code: ErrCodeCanceled,
		},
		{
			Match: func(err error) bool {
				return errors.Is(err, context.DeadlineExceeded)
			},
			Code: ErrCodeTimeout,
		},
		{
			Match: func(err error) bool {
				var netErr net.Error
				return errors.As(err, &netErr) && netErr.Timeout()
			},
			Code: ErrCodeTimeout,
		},
	}
}

// stringPatternMatchers returns matchers based on error string patterns.
func stringPatternMatchers() []ErrorMatcher {
	// Define pattern groups for easier maintenance
	authPatterns := []string{"unauthorized", "authentication", "api key", "401"}
	rateLimitPatterns := []string{"rate limit", "too many requests", "429"}
	invalidPatterns := []string{"invalid request", "bad request", "400"}
	notFoundPatterns := []string{"not found", "404"}
	quotaPatterns := []string{"quota", "limit exceeded", "insufficient"}
	contentPatterns := []string{"content filter", "safety", "blocked", "inappropriate"}
	tokenPatterns := []string{"token limit", "maximum context", "context length", "too long"}
	unavailablePatterns := []string{"service unavailable", "503", "500", "internal server"}
	notImplPatterns := []string{"not implemented", "not supported", "unsupported"}

	return []ErrorMatcher{
		makeStringMatcher(authPatterns, ErrCodeAuthentication),
		makeStringMatcher(rateLimitPatterns, ErrCodeRateLimit),
		makeStringMatcher(invalidPatterns, ErrCodeInvalidRequest),
		makeStringMatcher(notFoundPatterns, ErrCodeResourceNotFound),
		makeStringMatcher(quotaPatterns, ErrCodeQuotaExceeded),
		makeStringMatcher(contentPatterns, ErrCodeContentFilter),
		makeStringMatcher(tokenPatterns, ErrCodeTokenLimit),
		makeStringMatcher(unavailablePatterns, ErrCodeProviderUnavailable),
		makeStringMatcher(notImplPatterns, ErrCodeNotImplemented),
	}
}

// makeStringMatcher creates an ErrorMatcher that checks for any of the given patterns.
func makeStringMatcher(patterns []string, code ErrorCode) ErrorMatcher {
	return ErrorMatcher{
		Match: func(err error) bool {
			s := strings.ToLower(err.Error())
			for _, pattern := range patterns {
				if strings.Contains(s, pattern) {
					return true
				}
			}
			return false
		},
		Code: code,
	}
}

// AddMatcher adds a custom error matcher.
func (m *ErrorMapper) AddMatcher(matcher ErrorMatcher) *ErrorMapper {
	// Prepend custom matchers so they take precedence
	m.matchers = append([]ErrorMatcher{matcher}, m.matchers...)
	return m
}

// WrapError wraps an error with standardized error information.
func (m *ErrorMapper) WrapError(err error) error {
	if err == nil {
		return nil
	}

	// Check if already wrapped
	var stdErr *Error
	if errors.As(err, &stdErr) {
		return err
	}

	// Find matching error code
	code := ErrCodeUnknown
	message := err.Error()

	for _, matcher := range m.matchers {
		if matcher.Match(err) {
			code = matcher.Code
			if matcher.Transform != nil {
				message = matcher.Transform(err)
			}
			break
		}
	}

	return NewError(code, m.provider, message).WithCause(err)
}

// Map is an alias for WrapError for consistency with provider error mappers.
func (m *ErrorMapper) Map(err error) error {
	return m.WrapError(err)
}

// OpenAIErrorMapper creates an error mapper with OpenAI-specific patterns.
func OpenAIErrorMapper() *ErrorMapper {
	mapper := NewErrorMapper("openai")

	// Add OpenAI-specific matchers
	mapper.AddMatcher(ErrorMatcher{
		Match: func(err error) bool {
			s := err.Error()
			return strings.Contains(s, "invalid_api_key")
		},
		Code: ErrCodeAuthentication,
		Transform: func(_ error) string {
			return "Invalid OpenAI API key. Please check your OPENAI_API_KEY environment variable."
		},
	})

	mapper.AddMatcher(ErrorMatcher{
		Match: func(err error) bool {
			s := err.Error()
			return strings.Contains(s, "model_not_found")
		},
		Code: ErrCodeResourceNotFound,
		Transform: func(_ error) string {
			return "Model not found. Please check the model name and your API access."
		},
	})

	return mapper
}

// AnthropicErrorMapper creates an error mapper with Anthropic-specific patterns.
func AnthropicErrorMapper() *ErrorMapper {
	mapper := NewErrorMapper("anthropic")

	// Add Anthropic-specific matchers
	mapper.AddMatcher(ErrorMatcher{
		Match: func(err error) bool {
			s := err.Error()
			return strings.Contains(s, "invalid_x_api_key")
		},
		Code: ErrCodeAuthentication,
		Transform: func(_ error) string {
			return "Invalid Anthropic API key. Please check your ANTHROPIC_API_KEY environment variable."
		},
	})

	mapper.AddMatcher(ErrorMatcher{
		Match: func(err error) bool {
			s := err.Error()
			return strings.Contains(s, "credit_balance")
		},
		Code: ErrCodeQuotaExceeded,
		Transform: func(_ error) string {
			return "Anthropic API credit balance exceeded."
		},
	})

	return mapper
}

// GoogleAIErrorMapper creates an error mapper with Google AI-specific patterns.
func GoogleAIErrorMapper() *ErrorMapper {
	mapper := NewErrorMapper("googleai")

	// Add Google AI-specific matchers
	mapper.AddMatcher(ErrorMatcher{
		Match: func(err error) bool {
			s := err.Error()
			return strings.Contains(s, "API key not valid")
		},
		Code: ErrCodeAuthentication,
		Transform: func(_ error) string {
			return "Invalid Google AI API key. Please check your GOOGLE_API_KEY environment variable."
		},
	})

	mapper.AddMatcher(ErrorMatcher{
		Match: func(err error) bool {
			s := err.Error()
			return strings.Contains(s, "RECITATION") || strings.Contains(s, "SAFETY")
		},
		Code: ErrCodeContentFilter,
	})

	return mapper
}
