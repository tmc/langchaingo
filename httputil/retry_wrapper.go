package httputil

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// ErrorClassifier attempts to classify an error into a standardized [*llms.Error].
// Provider implementations should use their [MapError] functions for this.
// Returns nil if the error cannot be classified.
type ErrorClassifier func(error) error

// RetryOnError executes fn with retry on transient failures. It is the single
// unified retry mechanism that checks ALL retryable conditions in one pass:
//
//  1. Network errors (connection refused, DNS failure, TLS timeout)
//  2. HTTP status code patterns in error messages ("429", "500", "503", etc.)
//  3. Provider-specific error classification via the classifier (MapError)
//     — catches body-level errors like ERNIE's HTTP 200 + error_code:18
//
// If ANY condition indicates the error is retryable, the request is retried
// with exponential backoff. This replaces both [RetryTransport] and per-provider
// retry logic, ensuring a single retry budget is used.
//
// When the classified error carries a Retry-After hint (from [llms.Error.WithRetryAfter]
// or [*ResponseError]), the wait duration is max(computed_backoff, retryAfter).
//
// Usage:
//
//	err = httputil.RetryOnError(ctx, retryCfg, openai.MapError, func() error {
//	    result, err = client.CreateChat(ctx, req)
//	    return err
//	})
func RetryOnError(ctx context.Context, cfg *RetryConfig, classifier ErrorClassifier, fn func() error) error {
	if cfg == nil {
		return fn()
	}

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Classify the error once per iteration.
		var classified error
		if classifier != nil {
			classified = classifier(err)
		}

		// Exhausted retries — return classified error if available.
		if attempt >= cfg.MaxRetries {
			if classified != nil {
				return classified
			}
			return lastErr
		}

		// Check retryability: network error or classified error code.
		if !isProviderRetryable(err, cfg, classified) {
			if classified != nil {
				return classified
			}
			return err
		}

		// Calculate wait duration, respecting Retry-After if present.
		waitDuration := providerBackoff(cfg, attempt)
		if ra := extractRetryAfter(err, classified); ra > waitDuration {
			waitDuration = ra
		}

		fireOnRetry(cfg, attempt, err)
		if !wait(ctx, waitDuration) {
			return ctx.Err()
		}
	}

	return fmt.Errorf("retry: max retries (%d) exceeded, last error: %w", cfg.MaxRetries, lastErr)
}

// isProviderRetryable checks all retry conditions in one pass.
// Returns true if the error should be retried.
func isProviderRetryable(err error, cfg *RetryConfig, classified error) bool {
	// 1. Network errors (connection refused, DNS, TLS, etc.)
	if isRetryableError(err, cfg) {
		return true
	}

	// 2. Classified error is a retryable code
	//    (rate limit, provider unavailable, timeout).
	if classified != nil {
		if llms.IsRetryableError(classified) {
			return true
		}
	}

	return false
}

// retryAfterer is an interface for errors that carry Retry-After information.
type retryAfterer interface {
	RetryAfter() time.Duration
}

// extractRetryAfter checks both the classified error and the original error
// for Retry-After information. Returns 0 if neither carries it.
func extractRetryAfter(original, classified error) time.Duration {
	// Prefer classified error (MapError may have enriched it from ResponseError).
	if classified != nil {
		var ra retryAfterer
		if errors.As(classified, &ra) {
			if d := ra.RetryAfter(); d > 0 {
				return d
			}
		}
	}
	// Fall back to original error (direct *ResponseError from HTTP layer).
	var ra retryAfterer
	if errors.As(original, &ra) {
		return ra.RetryAfter()
	}
	return 0
}

// providerBackoff calculates backoff for provider-layer retries.
func providerBackoff(cfg *RetryConfig, attempt int) time.Duration {
	if attempt <= 0 {
		return cfg.InitialBackoff
	}

	backoff := float64(cfg.InitialBackoff) * math.Pow(cfg.BackoffFactor, float64(attempt))
	if backoff > float64(cfg.MaxBackoff) {
		backoff = float64(cfg.MaxBackoff)
	}

	jitter := 0.5 + rand.Float64()*0.5 //nolint:gosec // G404: jitter doesn't need crypto-rand
	return time.Duration(backoff * jitter)
}
