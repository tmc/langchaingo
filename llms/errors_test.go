package llms_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/vendasta/langchaingo/llms"
)

func TestStandardError(t *testing.T) {
	tests := []struct {
		name     string
		err      *llms.Error
		expected string
	}{
		{
			name:     "basic error",
			err:      llms.NewError(llms.ErrCodeAuthentication, "openai", "Invalid API key", 0),
			expected: "openai: authentication: Invalid API key",
		},
		{
			name:     "error without provider",
			err:      llms.NewError(llms.ErrCodeRateLimit, "", "Too many requests", 0),
			expected: "rate_limit: Too many requests",
		},
		{
			name: "error with cause",
			err: llms.NewError(llms.ErrCodeTimeout, "anthropic", "Request timed out", 0).
				WithCause(context.DeadlineExceeded),
			expected: "anthropic: timeout: Request timed out",
		},
		{
			name: "error with details",
			err: llms.NewError(llms.ErrCodeTokenLimit, "openai", "Token limit exceeded", 0).
				WithDetail("limit", 4096).
				WithDetail("used", 5000),
			expected: "openai: token_limit: Token limit exceeded",
		},
		{
			name: "error with details and status code",
			err: llms.NewError(llms.ErrCodeTokenLimit, "openai", "Token limit exceeded", 429).
				WithDetail("limit", 4096).
				WithDetail("used", 5000),
			expected: "openai: token_limit: Token limit exceeded (HTTP 429)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	authErr := llms.NewError(llms.ErrCodeAuthentication, "test", "auth failed", 0)
	cancelErr := llms.NewError(llms.ErrCodeCanceled, "test", "canceled", 0).WithCause(context.Canceled)

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "same error code",
			err:    authErr,
			target: llms.ErrAuthentication,
			want:   true,
		},
		{
			name:   "different error code",
			err:    authErr,
			target: llms.ErrRateLimit,
			want:   false,
		},
		{
			name:   "context canceled",
			err:    cancelErr,
			target: context.Canceled,
			want:   true,
		},
		{
			name:   "wrapped error",
			err:    fmt.Errorf("wrapped: %w", authErr),
			target: llms.ErrAuthentication,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check func(error) bool
		want  bool
	}{
		{
			name:  "IsAuthenticationError true",
			err:   llms.NewError(llms.ErrCodeAuthentication, "test", "auth failed", 0),
			check: llms.IsAuthenticationError,
			want:  true,
		},
		{
			name:  "IsAuthenticationError false",
			err:   llms.NewError(llms.ErrCodeRateLimit, "test", "rate limited", 0),
			check: llms.IsAuthenticationError,
			want:  false,
		},
		{
			name:  "IsRateLimitError wrapped",
			err:   fmt.Errorf("provider error: %w", llms.NewError(llms.ErrCodeRateLimit, "test", "rate limited", 0)),
			check: llms.IsRateLimitError,
			want:  true,
		},
		{
			name:  "IsTimeoutError",
			err:   llms.NewError(llms.ErrCodeTimeout, "test", "timeout", 0),
			check: llms.IsTimeoutError,
			want:  true,
		},
		{
			name:  "IsContentFilterError",
			err:   llms.NewError(llms.ErrCodeContentFilter, "test", "blocked", 0),
			check: llms.IsContentFilterError,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check(tt.err); got != tt.want {
				t.Errorf("check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMapper(t *testing.T) {
	mapper := llms.NewErrorMapper("test-provider")

	tests := []struct {
		name         string
		err          error
		expectedCode llms.ErrorCode
	}{
		{
			name:         "context canceled",
			err:          context.Canceled,
			expectedCode: llms.ErrCodeCanceled,
		},
		{
			name:         "context deadline",
			err:          context.DeadlineExceeded,
			expectedCode: llms.ErrCodeTimeout,
		},
		{
			name:         "network timeout",
			err:          &net.OpError{Op: "dial", Err: &timeoutError{}},
			expectedCode: llms.ErrCodeTimeout,
		},
		{
			name:         "unauthorized string",
			err:          errors.New("Unauthorized: Invalid API key"),
			expectedCode: llms.ErrCodeAuthentication,
		},
		{
			name:         "rate limit string",
			err:          errors.New("Error 429: Too Many Requests"),
			expectedCode: llms.ErrCodeRateLimit,
		},
		{
			name:         "not found",
			err:          errors.New("404: Model not found"),
			expectedCode: llms.ErrCodeResourceNotFound,
		},
		{
			name:         "token limit",
			err:          errors.New("Maximum context length exceeded"),
			expectedCode: llms.ErrCodeTokenLimit,
		},
		{
			name:         "content filter",
			err:          errors.New("Content blocked by safety filters"),
			expectedCode: llms.ErrCodeContentFilter,
		},
		{
			name:         "unknown error",
			err:          errors.New("Something went wrong"),
			expectedCode: llms.ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := mapper.WrapError(tt.err)

			var stdErr *llms.Error
			if !errors.As(wrapped, &stdErr) {
				t.Fatal("Expected wrapped error to be *llms.Error")
			}

			if stdErr.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", stdErr.Code, tt.expectedCode)
			}

			if stdErr.Provider != "test-provider" {
				t.Errorf("Provider = %v, want test-provider", stdErr.Provider)
			}

			if !errors.Is(wrapped, tt.err) {
				t.Error("Wrapped error should contain original error")
			}
		})
	}
}

func TestProviderSpecificMappers(t *testing.T) {
	t.Run("OpenAI mapper", func(t *testing.T) {
		mapper := llms.OpenAIErrorMapper()

		err := errors.New("invalid_api_key: Please check your API key")
		wrapped := mapper.WrapError(err)

		if !llms.IsAuthenticationError(wrapped) {
			t.Error("Expected authentication error")
		}

		if !errors.Is(wrapped, llms.ErrAuthentication) {
			t.Error("Expected to match ErrAuthentication")
		}
	})

	t.Run("Anthropic mapper", func(t *testing.T) {
		mapper := llms.AnthropicErrorMapper()

		err := errors.New("credit_balance insufficient")
		wrapped := mapper.WrapError(err)

		if !llms.IsQuotaExceededError(wrapped) {
			t.Error("Expected quota exceeded error")
		}
	})

	t.Run("GoogleAI mapper", func(t *testing.T) {
		mapper := llms.GoogleAIErrorMapper()

		err := errors.New("API key not valid. Please pass a valid API key")
		wrapped := mapper.WrapError(err)

		if !llms.IsAuthenticationError(wrapped) {
			t.Error("Expected authentication error")
		}
	})
}

func TestCustomMatcher(t *testing.T) {
	mapper := llms.NewErrorMapper("custom")

	// Add custom matcher
	mapper.AddMatcher(llms.ErrorMatcher{
		Match: func(err error) bool {
			return err.Error() == "custom error"
		},
		Code: llms.ErrCodeNotImplemented,
		Transform: func(_ error) string {
			return "This feature is not available"
		},
	})

	err := errors.New("custom error")
	wrapped := mapper.WrapError(err)

	var stdErr *llms.Error
	if !errors.As(wrapped, &stdErr) {
		t.Fatal("Expected *llms.Error")
	}

	if stdErr.Code != llms.ErrCodeNotImplemented {
		t.Errorf("Code = %v, want %v", stdErr.Code, llms.ErrCodeNotImplemented)
	}

	if stdErr.Message != "This feature is not available" {
		t.Errorf("Message = %v, want 'This feature is not available'", stdErr.Message)
	}
}

// timeoutError is a mock network timeout error.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
