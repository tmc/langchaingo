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
//	// JSONDebugClient pretty-prints JSON payloads with ANSI colors
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
package httputil
