package httputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.InitialBackoff != 1*time.Second {
		t.Errorf("expected InitialBackoff=1s, got %v", cfg.InitialBackoff)
	}
	if cfg.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff=30s, got %v", cfg.MaxBackoff)
	}
	if cfg.BackoffFactor != 2.0 {
		t.Errorf("expected BackoffFactor=2.0, got %f", cfg.BackoffFactor)
	}
}

func TestBackoff(t *testing.T) {
	cfg := &RetryConfig{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}

	tests := []struct {
		attempt  int
		min      time.Duration
		max      time.Duration
	}{
		{0, 1 * time.Second, 1 * time.Second},                         // exact InitialBackoff, no jitter for attempt 0
		{1, 1000 * time.Millisecond, 2000 * time.Millisecond},          // 2s * jitter [0.5, 1.0)
		{2, 2000 * time.Millisecond, 4000 * time.Millisecond},          // 4s * jitter [0.5, 1.0)
		{3, 4000 * time.Millisecond, 8000 * time.Millisecond},          // 8s * jitter [0.5, 1.0)
		{10, 15000 * time.Millisecond, 30000 * time.Millisecond},       // capped at MaxBackoff, jitter [0.5, 1.0)
	}

	for _, tt := range tests {
		got := cfg.backoff(tt.attempt)
		if got < tt.min || got > tt.max {
			t.Errorf("backoff(%d) = %v, want in [%v, %v]", tt.attempt, got, tt.min, tt.max)
		}
	}
}

func TestDefaultIsRetryableStatus(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	for _, code := range retryable {
		if !defaultIsRetryableStatus(code) {
			t.Errorf("expected status %d to be retryable", code)
		}
	}

	notRetryable := []int{200, 400, 401, 403, 404, 405, 408, 409, 422}
	for _, code := range notRetryable {
		if defaultIsRetryableStatus(code) {
			t.Errorf("expected status %d to NOT be retryable", code)
		}
	}
}

func TestDefaultIsRetryableError(t *testing.T) {
	tests := []struct {
		err       error
		retryable bool
	}{
		{nil, false},
		{context.Canceled, false},
		{context.DeadlineExceeded, false},
		{fmt.Errorf("connection refused"), true},
		{fmt.Errorf("connection reset by peer"), true},
		{fmt.Errorf("broken pipe"), true},
		{fmt.Errorf("TLS handshake timeout"), true},
		{fmt.Errorf("no such host"), true},
		{fmt.Errorf("i/o timeout"), true},
		{fmt.Errorf("unexpected EOF"), true},
		{fmt.Errorf("temporary failure"), true},
		{fmt.Errorf("some other error"), false},
	}

	for _, tt := range tests {
		got := defaultIsRetryableError(tt.err)
		if got != tt.retryable {
			t.Errorf("defaultIsRetryableError(%v) = %v, want %v", tt.err, got, tt.retryable)
		}
	}
}

func TestDefaultIsRetryableError_NetError(t *testing.T) {
	// net.Error should be retryable.
	netErr := &netOpError{Err: errors.New("network error")}
	if !defaultIsRetryableError(netErr) {
		t.Error("expected net.Error to be retryable")
	}
}

// netOpError implements net.Error for testing.
type netOpError struct {
	Err error
}

func (e *netOpError) Error() string   { return e.Err.Error() }
func (e *netOpError) Timeout() bool   { return false }
func (e *netOpError) Temporary() bool { return true }
func (e *netOpError) Unwrap() error   { return e.Err }

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		header    string
		hasResult bool
		min       time.Duration
	}{
		{"", false, 0},
		{"5", true, 4 * time.Second},  // 5 seconds
		{"0", true, -1 * time.Second}, // 0 seconds is valid but <=0
	}

	for _, tt := range tests {
		resp := &http.Response{Header: http.Header{"Retry-After": {tt.header}}}
		d, ok := parseRetryAfter(resp)
		if ok != tt.hasResult {
			t.Errorf("parseRetryAfter(%q) returned ok=%v, want %v", tt.header, ok, tt.hasResult)
		}
		if ok && d < tt.min {
			t.Errorf("parseRetryAfter(%q) = %v, want >= %v", tt.header, d, tt.min)
		}
	}
}

func TestRetryTransport_SuccessOnFirstAttempt(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetryTransport_RetriesOnServerError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryTransport_RetriesOnRateLimit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			w.Header().Set("Retry-After", "0") // immediate retry
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestRetryTransport_ExhaustsRetries(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	// After exhausting retries, the last 500 response is returned (not an error).
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// The last attempt returns the 500 response.
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
	// Total attempts = MaxRetries + 1 = 3
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryTransport_NoRetryOnClientError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestRetryTransport_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond, // longer than context timeout
		MaxBackoff:     500 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestRetryTransport_RequestBodyReplay(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))

		if len(bodies) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		server.URL,
		strings.NewReader(`{"model":"gpt-4","message":"hello"}`),
	)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if len(bodies) != 2 {
		t.Fatalf("expected 2 bodies, got %d", len(bodies))
	}
	for i, body := range bodies {
		expected := `{"model":"gpt-4","message":"hello"}`
		if body != expected {
			t.Errorf("body[%d] = %q, want %q", i, body, expected)
		}
	}
}

func TestRetryTransport_OnRetryCallback(t *testing.T) {
	callCount := 0
	var retryAttempts []int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		BackoffFactor:  2.0,
		OnRetry: func(attempt int, err error) {
			mu.Lock()
			retryAttempts = append(retryAttempts, attempt)
			mu.Unlock()
		},
	}
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	mu.Lock()
	defer mu.Unlock()
	if len(retryAttempts) != 2 {
		t.Errorf("expected 2 OnRetry calls, got %d", len(retryAttempts))
	}
}

func TestRetryTransport_CustomRetryableStatus(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return 408 (Request Timeout) — not in default retryable list
		w.WriteHeader(http.StatusRequestTimeout)
	}))
	defer server.Close()

	// Default config should NOT retry 408
	cfg := DefaultRetryConfig()
	cfg.InitialBackoff = 1 * time.Millisecond
	cfg.MaxBackoff = 5 * time.Millisecond
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, _ := client.Do(req)
	if callCount != 1 {
		t.Errorf("default config should NOT retry 408, got %d calls", callCount)
	}
	resp.Body.Close()

	// Custom config that retries 408
	callCount = 0
	cfg2 := DefaultRetryConfig()
	cfg2.InitialBackoff = 1 * time.Millisecond
	cfg2.MaxBackoff = 5 * time.Millisecond
	cfg2.RetryableStatus = func(code int) bool {
		return code == http.StatusRequestTimeout || defaultIsRetryableStatus(code)
	}

	client2 := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    cfg2,
		},
	}

	req2, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp2, err := client2.Do(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp2.Body.Close()
	if callCount != cfg2.MaxRetries+1 {
		t.Errorf("custom config should retry 408, got %d calls, want %d", callCount, cfg2.MaxRetries+1)
	}
}

func TestRetryTransport_NilConfigUsesDefault(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// RetryTransport with nil Config should use DefaultRetryConfig.
	client := &http.Client{
		Transport: &RetryTransport{
			Transport: http.DefaultTransport,
			Config:    nil,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestNewRetryClient(t *testing.T) {
	cfg := DefaultRetryConfig()
	client := NewRetryClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Verify the transport chain.
	rt, ok := client.Transport.(*RetryTransport)
	if !ok {
		t.Fatal("expected RetryTransport")
	}
	if rt.Config != cfg {
		t.Error("config mismatch")
	}
}

func TestNewClientWithRetry(t *testing.T) {
	cfg := DefaultRetryConfig()
	cfg.InitialBackoff = 1 * time.Millisecond
	cfg.MaxBackoff = 5 * time.Millisecond

	client := NewClientWithRetry(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Verify the transport chain: RetryTransport -> Transport -> http.DefaultTransport
	rt, ok := client.Transport.(*RetryTransport)
	if !ok {
		t.Fatal("expected RetryTransport as outer transport")
	}
	inner, ok := rt.Transport.(*Transport)
	if !ok {
		t.Fatal("expected Transport as inner transport")
	}
	if inner.Transport != http.DefaultTransport {
		t.Error("inner Transport should use http.DefaultTransport")
	}
}
