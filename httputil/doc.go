// Package httputil provides HTTP transport and client utilities for LangChainGo.
//
// The package offers several key features:
//
// # User-Agent Management
//
// All HTTP clients and transports in this package automatically add a User-Agent
// header that identifies the LangChainGo library, the calling program, and
// system information. This helps API providers understand client usage patterns
// and aids in debugging.
//
// The User-Agent format is:
//
//	program/version langchaingo/version Go/version (GOOS GOARCH)
//
// For example:
//
//	openai-chat-example/devel langchaingo/v0.1.8 Go/go1.21.0 (darwin arm64)
//
// # Default HTTP Client
//
// The package provides DefaultClient, which is a pre-configured http.Client
// that includes the User-Agent header:
//
//	resp, err := httputil.DefaultClient.Get("https://api.example.com/data")
//
// # Logging and Debugging
//
// For development and debugging, the package provides logging clients:
//
//	// LoggingClient logs full HTTP requests and responses using slog
//	client := httputil.LoggingClient
//
//	// JSONDebugClient pretty-prints JSON payloads and SSE streams with ANSI colors
//	client := httputil.JSONDebugClient
//
// # Custom Transports
//
// The Transport type implements http.RoundTripper and can be used to add
// the LangChainGo User-Agent to any HTTP client:
//
//	client := &http.Client{
//	    Transport: &httputil.Transport{
//	        Transport: myCustomTransport,
//	    },
//	}
//
// # Integration with httprr
//
// The transports in this package are designed to work with the httprr
// HTTP record/replay system used in tests. When using httprr, pass
// httputil.DefaultTransport to ensure proper request interception.
//
// # Automatic Retry
//
// The package provides a unified retry mechanism for transient HTTP failures.
// It is opt-in: retry is only enabled when a RetryConfig is explicitly provided.
//
// ## What gets retried
//
// The following transient failures are automatically retried:
//
//   - HTTP 429 (Rate Limit) — respects the Retry-After response header
//   - HTTP 500, 502, 503, 504 (Server Errors)
//   - Network errors (connection refused, DNS failure, TLS timeout)
//   - Provider-specific body-level errors (e.g., ERNIE's HTTP 200 + error_code:18)
//
// The following are NOT retried:
//
//   - HTTP 400, 401, 403, 404 and other 4xx client errors
//   - Context cancellation (context.Canceled)
//   - Context deadline exceeded (context.DeadlineExceeded)
//   - Successful responses
//
// ## Basic usage
//
// All LLM providers support retry via the WithRetryConfig option:
//
//	import "github.com/tmc/langchaingo/httputil"
//
//	// Use sensible defaults: 3 retries, 1s initial backoff, 30s max backoff
//	llm, err := openai.New(
//	    openai.WithToken("sk-xxx"),
//	    openai.WithRetryConfig(httputil.DefaultRetryConfig()),
//	)
//
// The same pattern works for all providers:
//
//	llm, _ := anthropic.New(anthropic.WithRetryConfig(cfg), ...)
//	llm, _ := ernie.New(ernie.WithRetryConfig(cfg), ...)
//	llm, _ := ollama.New(ollama.WithRetryConfig(cfg), ...)
//	llm, _ := cohere.New(cohere.WithRetryConfig(cfg), ...)
//	llm, _ := cloudflare.New(cloudflare.WithRetryConfig(cfg), ...)
//	llm, _ := huggingface.New(huggingface.WithRetryConfig(cfg), ...)
//	llm, _ := llamafile.New(llamafile.WithRetryConfig(cfg), ...)
//	llm, _ := maritaca.New(maritaca.WithRetryConfig(cfg), ...)
//
// When WithRetryConfig is not provided, no retry is performed (the default
// behavior is unchanged).
//
// ## Custom configuration
//
//	llm, err := openai.New(
//	    openai.WithToken("sk-xxx"),
//	    openai.WithRetryConfig(&httputil.RetryConfig{
//	        MaxRetries:     5,
//	        InitialBackoff: 2 * time.Second,
//	        MaxBackoff:     60 * time.Second,
//	        BackoffFactor:  2.0,
//	    }),
//	)
//
// The backoff sequence with Factor 2.0 and InitialBackoff 2s is:
//
//	attempt 0: 2s   (initial)
//	attempt 1: 4s   (2s × 2.0)
//	attempt 2: 8s   (4s × 2.0)
//	attempt 3: 16s  (8s × 2.0)
//	attempt 4: 32s  (capped at MaxBackoff 60s)
//
// Random jitter (±50%) is applied to prevent thundering herd.
//
// ## Retry-After header support
//
// When a provider returns HTTP 429 with a Retry-After header, the retry
// mechanism waits the duration specified by the server instead of the
// computed backoff:
//
//	// Server returns: HTTP 429 + Retry-After: 60
//	// Wait duration: max(computed_backoff, 60s) = 60s
//
// This applies to OpenAI, Anthropic, and ERNIE providers.
//
// ## Logging retries
//
// Use the OnRetry callback for observability:
//
//	llm, err := openai.New(
//	    openai.WithToken("sk-xxx"),
//	    openai.WithRetryConfig(&httputil.RetryConfig{
//	        MaxRetries:     3,
//	        InitialBackoff: 1 * time.Second,
//	        MaxBackoff:     30 * time.Second,
//	        BackoffFactor:  2.0,
//	        OnRetry: func(attempt int, err error) {
//	            slog.Warn("retrying request",
//	                "attempt", attempt+1,
//	                "error", err,
//	            )
//	        },
//	    }),
//	)
//
// ## Custom retry conditions
//
// Override the default retryable status codes or error checks:
//
//	cfg := httputil.DefaultRetryConfig()
//
//	// Also retry HTTP 408 (Request Timeout)
//	cfg.RetryableStatus = func(code int) bool {
//	    return code == 408 || code == 429 || (code >= 500 && code <= 504)
//	}
//
//	// Custom network error check
//	cfg.RetryableError = func(err error) bool {
//	    return myCustomCheck(err)
//	}
//
// ## Two internal mechanisms
//
// The package provides two retry implementations. Users do not need to
// choose — the appropriate one is selected automatically by each provider:
//
//   - RetryOnError (application-layer): Used by OpenAI, Anthropic, and ERNIE.
//     Retries based on network errors, HTTP status codes, AND provider-specific
//     body-level errors (e.g., ERNIE returns HTTP 200 with error_code in the
//     response body). Also respects Retry-After headers.
//
//   - RetryTransport (transport-layer): Used by Ollama, Cohere, Cloudflare,
//     HuggingFace, Llamafile, and Maritaca. Retries based on HTTP status codes
//     and network errors at the http.RoundTripper level. Also respects
//     Retry-After headers.
//
// Both mechanisms share the same RetryConfig, backoff algorithm, and jitter
// strategy. The only difference is that RetryOnError can additionally detect
// body-level errors via provider-specific MapError functions.
package httputil
