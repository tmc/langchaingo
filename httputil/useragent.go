package httputil

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	userAgent     string
	userAgentOnce sync.Once
)

// UserAgent returns the default User-Agent string for LangChainGo HTTP clients.
// Format: program/version langchaingo/version Go/version (GOOS GOARCH)
// Example: "openai-chat-example/devel langchaingo/v0.1.8 Go/go1.21.0 (darwin arm64)"
func UserAgent() string {
	userAgentOnce.Do(func() {
		parts := []string{}

		// Get build info once
		if info, ok := debug.ReadBuildInfo(); ok {
			// Add program name if available
			if info.Main.Path != "" && info.Main.Path != "command-line-arguments" {
				name := info.Main.Path[strings.LastIndex(info.Main.Path, "/")+1:]
				parts = append(parts, name+"/devel")
			}

			// Add langchaingo version
			langchainVer := "devel"
			for _, dep := range info.Deps {
				if dep.Path == "github.com/tmc/langchaingo" {
					langchainVer = strings.Trim(dep.Version, "()")
					break
				}
			}
			parts = append(parts, "langchaingo/"+langchainVer)
		} else {
			// Fallback if no build info
			parts = append(parts, "langchaingo/devel")
		}

		// Add Go version and platform
		parts = append(parts, fmt.Sprintf("Go/%s", runtime.Version()))
		parts = append(parts, fmt.Sprintf("(%s %s)", runtime.GOOS, runtime.GOARCH))

		userAgent = strings.Join(parts, " ")
	})
	return userAgent
}
