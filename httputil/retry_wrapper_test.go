package httputil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// --- ResponseError tests ---

func TestResponseError_Error(t *testing.T) {
	err := &ResponseError{
		StatusCode: 429,
		Message:    "API returned unexpected status code: 429: Rate limit exceeded",
		RetryAfter: 30 * time.Second,
	}
	if err.Error() != "API returned unexpected status code: 429: Rate limit exceeded" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestParseRetryAfterHeader_Integer(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "60")
	resp := &http.Response{Header: h}
	d := ParseRetryAfterHeader(resp)
	if d != 60*time.Second {
		t.Errorf("expected 60s, got %v", d)
	}
}

func TestParseRetryAfterHeader_Absent(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	d := ParseRetryAfterHeader(resp)
	if d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestNewResponseError(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "10")
	resp := &http.Response{
		StatusCode: 429,
		Header:     h,
	}
	err := NewResponseError(resp, "rate limited")
	if err.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", err.StatusCode)
	}
	if err.Message != "rate limited" {
		t.Errorf("unexpected message: %s", err.Message)
	}
	if err.RetryAfter != 10*time.Second {
		t.Errorf("expected 10s RetryAfter, got %v", err.RetryAfter)
	}
}

// mockRetryAfterClassifier simulates a provider MapError that extracts
// RetryAfter from a ResponseError and sets it on a real llms.Error.
func mockRetryAfterClassifier(err error) error {
	var respErr *ResponseError
	if errors.As(err, &respErr) && respErr.StatusCode == 429 {
		return llms.NewError(llms.ErrCodeRateLimit, "test", "rate limited").
			WithCause(err).
			WithRetryAfter(respErr.RetryAfter)
	}
	return nil
}

// --- RetryOnError tests ---

func TestRetryOnError_NilConfig_ExecutesOnce(t *testing.T) {
	callCount := 0
	err := RetryOnError(context.Background(), nil, nil, func() error {
		callCount++
		return fmt.Errorf("some error")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call with nil config, got %d", callCount)
	}
}

func TestRetryOnError_RetriesOnNetworkError(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	callCount := 0
	err := RetryOnError(context.Background(), cfg, nil, func() error {
		callCount++
		return fmt.Errorf("connection refused")
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if callCount != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryOnError_SucceedsAfterRetry(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	callCount := 0
	err := RetryOnError(context.Background(), cfg, nil, func() error {
		callCount++
		if callCount < 3 {
			return fmt.Errorf("connection refused")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryOnError_UsesClassifier(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	classifier := func(err error) error {
		if err != nil && err.Error() == "server busy" {
			return llms.NewError(llms.ErrCodeRateLimit, "test", "classified: rate_limit")
		}
		return nil
	}

	callCount := 0
	err := RetryOnError(context.Background(), cfg, classifier, func() error {
		callCount++
		return fmt.Errorf("server busy")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	// Should be the llms.Error from classifier
	var llmsErr *llms.Error
	if !errors.As(err, &llmsErr) {
		t.Fatalf("expected llms.Error, got: %T: %v", err, err)
	}
	if llmsErr.Message != "classified: rate_limit" {
		t.Errorf("expected classified message, got: %s", llmsErr.Message)
	}
}

func TestRetryOnError_NoRetryOnNonRetryableError(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	callCount := 0
	err := RetryOnError(context.Background(), cfg, nil, func() error {
		callCount++
		return fmt.Errorf("some non-retryable error")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestRetryOnError_ContextCancellation(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := RetryOnError(ctx, cfg, nil, func() error {
		return fmt.Errorf("connection refused")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestRetryOnError_RespectsRetryAfter(t *testing.T) {
	cfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	callCount := 0
	err := RetryOnError(context.Background(), cfg, mockRetryAfterClassifier, func() error {
		callCount++
		return &ResponseError{
			StatusCode: 429,
			Message:    "rate limited",
			RetryAfter: 50 * time.Millisecond,
		}
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if callCount != 3 { // 1 + 2 retries
		t.Errorf("expected 3 calls, got %d", callCount)
	}

	// Verify classified error carries RetryAfter
	var llmsErr *llms.Error
	if !errors.As(err, &llmsErr) {
		t.Fatal("expected llms.Error")
	}
	if llmsErr.RetryAfter() != 50*time.Millisecond {
		t.Errorf("expected RetryAfter 50ms, got %v", llmsErr.RetryAfter())
	}
}

func TestRetryOnError_OnRetryCallback(t *testing.T) {
	var mu sync.Mutex
	var attempts []int

	cfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
		OnRetry: func(attempt int, err error) {
			mu.Lock()
			attempts = append(attempts, attempt)
			mu.Unlock()
		},
	}

	_ = RetryOnError(context.Background(), cfg, nil, func() error {
		return fmt.Errorf("connection refused")
	})

	mu.Lock()
	defer mu.Unlock()
	if len(attempts) != 2 {
		t.Errorf("expected 2 OnRetry calls, got %d", len(attempts))
	}
}

// --- Integration: RetryOnError with real HTTP server ---

func TestRetryOnError_Integration(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"result":"ok"}`)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}

	err := RetryOnError(context.Background(), cfg, mockRetryAfterClassifier, func() error {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return NewResponseError(resp, fmt.Sprintf("status %d", resp.StatusCode))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}
