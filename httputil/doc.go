// Package httputil provides HTTP transport and client utilities for LangChainGo.
//
// # User-Agent Management
//
// All transports automatically add a User-Agent header identifying
// LangChainGo, the calling program, and system information:
//
//	program/version langchaingo/version Go/version (GOOS GOARCH)
//
// Example:
//
//	openai-chat-example/devel langchaingo/v0.1.8 Go/go1.21.0 (darwin arm64)
//
// # Default Client
//
// DefaultClient includes the LangChainGo User-Agent:
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
// Use Transport to add the User-Agent to any HTTP client:
//
//	client := &http.Client{
//	    Transport: &httputil.Transport{
//	        Transport: myCustomTransport,
//	    },
//	}
//
// # Testing with httprr
//
// The transports work with httprr for HTTP record/replay.
// Use httputil.DefaultTransport for proper request interception.
package httputil
