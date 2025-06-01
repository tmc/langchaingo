package llms_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestErrorMapperAddMatcher(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	// Add a custom matcher that takes precedence
	mapper.AddMatcher(llms.ErrorMatcher{
		Match: func(err error) bool {
			return err.Error() == "special custom error"
		},
		Code: llms.ErrCodeNotImplemented,
		Transform: func(err error) string {
			return "transformed: " + err.Error()
		},
	})

	// This error would normally match authentication patterns
	customErr := errors.New("special custom error")
	wrapped := mapper.WrapError(customErr)

	var stdErr *llms.Error
	if !errors.As(wrapped, &stdErr) {
		t.Fatal("Expected *llms.Error")
	}

	// Should match our custom matcher first, not the auth pattern
	if stdErr.Code != llms.ErrCodeNotImplemented {
		t.Errorf("Code = %v, want %v", stdErr.Code, llms.ErrCodeNotImplemented)
	}

	if stdErr.Message != "transformed: special custom error" {
		t.Errorf("Message = %v, want 'transformed: special custom error'", stdErr.Message)
	}
}

func TestErrorMapperNilError(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	wrapped := mapper.WrapError(nil)
	if wrapped != nil {
		t.Errorf("WrapError(nil) = %v, want nil", wrapped)
	}

	// Test Map alias as well
	wrapped = mapper.Map(nil)
	if wrapped != nil {
		t.Errorf("Map(nil) = %v, want nil", wrapped)
	}
}

func TestErrorMapperAlreadyWrapped(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	// Create an already wrapped error
	originalErr := llms.NewError(llms.ErrCodeAuthentication, "original", "auth failed")

	// Should return the same error without double-wrapping
	wrapped := mapper.WrapError(originalErr)
	if wrapped != originalErr {
		t.Error("Expected same error instance for already wrapped error")
	}
}

func TestErrorMapperProviderSpecificPatterns(t *testing.T) {
	tests := []struct {
		name         string
		mapper       *llms.ErrorMapper
		err          error
		expectedCode llms.ErrorCode
		checkMessage bool
		expectedMsg  string
	}{
		// OpenAI specific patterns
		{
			name:         "OpenAI invalid_api_key",
			mapper:       llms.OpenAIErrorMapper(),
			err:          errors.New("invalid_api_key: API key is invalid"),
			expectedCode: llms.ErrCodeAuthentication,
			checkMessage: true,
			expectedMsg:  "Invalid OpenAI API key. Please check your OPENAI_API_KEY environment variable.",
		},
		{
			name:         "OpenAI model_not_found",
			mapper:       llms.OpenAIErrorMapper(),
			err:          errors.New("model_not_found: gpt-5 does not exist"),
			expectedCode: llms.ErrCodeResourceNotFound,
			checkMessage: true,
			expectedMsg:  "Model not found. Please check the model name and your API access.",
		},
		// Anthropic specific patterns
		{
			name:         "Anthropic invalid_x_api_key",
			mapper:       llms.AnthropicErrorMapper(),
			err:          errors.New("invalid_x_api_key: Invalid API key provided"),
			expectedCode: llms.ErrCodeAuthentication,
			checkMessage: true,
			expectedMsg:  "Invalid Anthropic API key. Please check your ANTHROPIC_API_KEY environment variable.",
		},
		{
			name:         "Anthropic credit_balance",
			mapper:       llms.AnthropicErrorMapper(),
			err:          errors.New("credit_balance insufficient for request"),
			expectedCode: llms.ErrCodeQuotaExceeded,
			checkMessage: true,
			expectedMsg:  "Anthropic API credit balance exceeded.",
		},
		// Google AI specific patterns
		{
			name:         "Google AI API key not valid",
			mapper:       llms.GoogleAIErrorMapper(),
			err:          errors.New("API key not valid. Please pass a valid API key."),
			expectedCode: llms.ErrCodeAuthentication,
			checkMessage: true,
			expectedMsg:  "Invalid Google AI API key. Please check your GOOGLE_API_KEY environment variable.",
		},
		{
			name:         "Google AI RECITATION",
			mapper:       llms.GoogleAIErrorMapper(),
			err:          errors.New("Content blocked due to RECITATION"),
			expectedCode: llms.ErrCodeContentFilter,
			checkMessage: false,
		},
		{
			name:         "Google AI SAFETY",
			mapper:       llms.GoogleAIErrorMapper(),
			err:          errors.New("Response blocked by SAFETY filters"),
			expectedCode: llms.ErrCodeContentFilter,
			checkMessage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := tt.mapper.WrapError(tt.err)

			var stdErr *llms.Error
			if !errors.As(wrapped, &stdErr) {
				t.Fatal("Expected *llms.Error")
			}

			if stdErr.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", stdErr.Code, tt.expectedCode)
			}

			if tt.checkMessage && stdErr.Message != tt.expectedMsg {
				t.Errorf("Message = %v, want %v", stdErr.Message, tt.expectedMsg)
			}
		})
	}
}

func TestErrorMapperEdgeCases(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	tests := []struct {
		name         string
		err          error
		expectedCode llms.ErrorCode
	}{
		{
			name:         "empty error message",
			err:          errors.New(""),
			expectedCode: llms.ErrCodeUnknown,
		},
		{
			name:         "case insensitive matching",
			err:          errors.New("UNAUTHORIZED ACCESS"),
			expectedCode: llms.ErrCodeAuthentication,
		},
		{
			name:         "partial pattern matching",
			err:          errors.New("The request failed with 429 status"),
			expectedCode: llms.ErrCodeRateLimit,
		},
		{
			name:         "multiple pattern matches (first wins)",
			err:          errors.New("401 unauthorized rate limit exceeded"),
			expectedCode: llms.ErrCodeAuthentication, // auth patterns are checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := mapper.WrapError(tt.err)

			var stdErr *llms.Error
			if !errors.As(wrapped, &stdErr) {
				t.Fatal("Expected *llms.Error")
			}

			if stdErr.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", stdErr.Code, tt.expectedCode)
			}
		})
	}
}

func TestErrorMapperAllPatterns(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	// Test that all default patterns work correctly
	patterns := []struct {
		name         string
		errorMsg     string
		expectedCode llms.ErrorCode
	}{
		// Authentication patterns
		{"unauthorized", "Error: unauthorized access", llms.ErrCodeAuthentication},
		{"authentication", "authentication failed", llms.ErrCodeAuthentication},
		{"api key", "invalid api key provided", llms.ErrCodeAuthentication},
		{"401", "HTTP 401 Unauthorized", llms.ErrCodeAuthentication},

		// Rate limit patterns
		{"rate limit", "rate limit exceeded", llms.ErrCodeRateLimit},
		{"too many requests", "Error: too many requests", llms.ErrCodeRateLimit},
		{"429", "HTTP 429", llms.ErrCodeRateLimit},

		// Invalid request patterns
		{"invalid request", "invalid request format", llms.ErrCodeInvalidRequest},
		{"bad request", "bad request: missing parameter", llms.ErrCodeInvalidRequest},
		{"400", "HTTP 400 Bad Request", llms.ErrCodeInvalidRequest},

		// Not found patterns
		{"not found", "resource not found", llms.ErrCodeResourceNotFound},
		{"404", "HTTP 404", llms.ErrCodeResourceNotFound},

		// Quota patterns
		{"quota", "quota exceeded", llms.ErrCodeQuotaExceeded},
		{"limit exceeded", "monthly limit exceeded", llms.ErrCodeQuotaExceeded},
		{"insufficient", "insufficient credits", llms.ErrCodeQuotaExceeded},

		// Content filter patterns
		{"content filter", "blocked by content filter", llms.ErrCodeContentFilter},
		{"safety", "safety filter triggered", llms.ErrCodeContentFilter},
		{"blocked", "content blocked", llms.ErrCodeContentFilter},
		{"inappropriate", "inappropriate content detected", llms.ErrCodeContentFilter},

		// Token limit patterns
		{"token limit", "token limit reached", llms.ErrCodeTokenLimit},
		{"maximum context", "maximum context length exceeded", llms.ErrCodeTokenLimit},
		{"context length", "context length too large", llms.ErrCodeTokenLimit},
		{"too long", "input too long", llms.ErrCodeTokenLimit},

		// Unavailable patterns
		{"service unavailable", "service unavailable", llms.ErrCodeProviderUnavailable},
		{"503", "HTTP 503", llms.ErrCodeProviderUnavailable},
		{"500", "HTTP 500 Internal Server Error", llms.ErrCodeProviderUnavailable},
		{"internal server", "internal server error", llms.ErrCodeProviderUnavailable},

		// Not implemented patterns
		{"not implemented", "feature not implemented", llms.ErrCodeNotImplemented},
		{"not supported", "operation not supported", llms.ErrCodeNotImplemented},
		{"unsupported", "unsupported model", llms.ErrCodeNotImplemented},
	}

	for _, tt := range patterns {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := mapper.WrapError(errors.New(tt.errorMsg))

			var stdErr *llms.Error
			if !errors.As(wrapped, &stdErr) {
				t.Fatal("Expected *llms.Error")
			}

			if stdErr.Code != tt.expectedCode {
				t.Errorf("For pattern %q: Code = %v, want %v", tt.name, stdErr.Code, tt.expectedCode)
			}
		})
	}
}

// Custom network error for testing
type customNetError struct {
	timeout bool
}

func (e *customNetError) Error() string   { return "network error" }
func (e *customNetError) Timeout() bool   { return e.timeout }
func (e *customNetError) Temporary() bool { return true }

func TestErrorMapperNetworkErrors(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	// Test network timeout error
	netErr := &net.OpError{
		Op:  "dial",
		Err: &customNetError{timeout: true},
	}

	wrapped := mapper.WrapError(netErr)

	var stdErr *llms.Error
	if !errors.As(wrapped, &stdErr) {
		t.Fatal("Expected *llms.Error")
	}

	if stdErr.Code != llms.ErrCodeTimeout {
		t.Errorf("Code = %v, want %v", stdErr.Code, llms.ErrCodeTimeout)
	}
}

func TestErrorMapperTransform(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	// Add matcher with transform that modifies the message
	mapper.AddMatcher(llms.ErrorMatcher{
		Match: func(err error) bool {
			return err.Error() == "raw error"
		},
		Code: llms.ErrCodeUnknown,
		Transform: func(err error) string {
			return "Transformed: " + err.Error()
		},
	})

	err := errors.New("raw error")
	wrapped := mapper.WrapError(err)

	var stdErr *llms.Error
	if !errors.As(wrapped, &stdErr) {
		t.Fatal("Expected *llms.Error")
	}

	if stdErr.Message != "Transformed: raw error" {
		t.Errorf("Message = %v, want 'Transformed: raw error'", stdErr.Message)
	}
}

func TestErrorMapperContextErrors(t *testing.T) {
	mapper := llms.NewErrorMapper("test")

	tests := []struct {
		name         string
		err          error
		expectedCode llms.ErrorCode
	}{
		{
			name:         "context.Canceled",
			err:          context.Canceled,
			expectedCode: llms.ErrCodeCanceled,
		},
		{
			name:         "wrapped context.Canceled",
			err:          errors.Join(errors.New("operation failed"), context.Canceled),
			expectedCode: llms.ErrCodeCanceled,
		},
		{
			name:         "context.DeadlineExceeded",
			err:          context.DeadlineExceeded,
			expectedCode: llms.ErrCodeTimeout,
		},
		{
			name:         "wrapped context.DeadlineExceeded",
			err:          errors.Join(errors.New("request failed"), context.DeadlineExceeded),
			expectedCode: llms.ErrCodeTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := mapper.WrapError(tt.err)

			var stdErr *llms.Error
			if !errors.As(wrapped, &stdErr) {
				t.Fatal("Expected *llms.Error")
			}

			if stdErr.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", stdErr.Code, tt.expectedCode)
			}
		})
	}
}
