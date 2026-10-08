// Package testctr provides utilities for setting up testcontainers in tests.
//
// Usage:
//
// 1. Add a TestMain function to your test package:
//
//	func TestMain(m *testing.M) {
//		code := testctr.EnsureTestEnv()
//		if code == 0 {
//			code = m.Run()
//		}
//		os.Exit(code)
//	}
//
// 2. In your test functions, check for Docker availability:
//
//	func TestWithContainers(t *testing.T) {
//		testctr.SkipIfDockerNotAvailable(t)
//		// Your test code here
//	}
//
// This approach ensures proper environment setup for testcontainers while
// maintaining compatibility with parallel tests.
package testctr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// EnsureTestEnv sets up the necessary environment variables for testcontainers
// at the process level. This should be called from TestMain before running tests.
// It returns an exit code that should be passed to os.Exit.
//
// This works around a testcontainers bug where it doesn't properly detect
// the Docker socket when using Colima or other non-standard Docker setups.
//
// Example usage:
//
//	func TestMain(m *testing.M) {
//		code := testctr.EnsureTestEnv()
//		if code == 0 {
//			code = m.Run()
//		}
//		os.Exit(code)
//	}
func EnsureTestEnv() int {
	verbose := os.Getenv("TESTCONTAINERS_VERBOSE") == "true"
	// Check if docker is available in PATH
	_, err := exec.LookPath("docker")
	if err != nil {
		// If DOCKER_HOST is set, assume Docker is available remotely
		if os.Getenv("DOCKER_HOST") == "" {
			fmt.Fprintf(os.Stderr, "WARNING: Docker not found in PATH and DOCKER_HOST not set\n")
			// Don't fail, just warn - tests will skip individually
			return 0
		}
		// Docker CLI not found but DOCKER_HOST is set, so continue
	}

	// Only set environment variables if they're not already set
	if os.Getenv("DOCKER_HOST") == "" && err == nil {
		// Get Docker host from docker context
		cmd := exec.Command("docker", "context", "inspect", "-f={{.Endpoints.docker.Host}}")
		output, err := cmd.CombinedOutput()
		if err == nil {
			dockerHost := strings.TrimSpace(string(output))

			// Set DOCKER_HOST if using non-standard Docker socket paths (Colima, Lima, etc.)
			// This works around testcontainers bug where it doesn't properly detect non-standard sockets
			if dockerHost != "" && (strings.Contains(dockerHost, "colima") ||
				strings.Contains(dockerHost, ".lima") ||
				!strings.Contains(dockerHost, "/var/run/docker.sock")) {
				os.Setenv("DOCKER_HOST", dockerHost)
				if verbose {
					fmt.Fprintf(os.Stderr, "testctr: Set DOCKER_HOST=%s\n", dockerHost)
				}
			} else if verbose && dockerHost != "" {
				fmt.Fprintf(os.Stderr, "testctr: Using standard Docker socket: %s\n", dockerHost)
			}
		}
		// If docker context inspect fails, just continue without setting DOCKER_HOST
		// This is common when using standard Docker Desktop
	}

	// Disable Ryuk reaper if not explicitly enabled to reduce resource usage
	// Ryuk is used for cleanup but can cause issues with limited Docker resources
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		if verbose {
			fmt.Fprintf(os.Stderr, "testctr: Disabled Ryuk reaper for resource efficiency\n")
		}
	}

	// Set the testcontainers Docker socket override if not already set
	// This tells testcontainers where to find the actual Docker socket
	if os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE") == "" {
		dockerSocket := "/var/run/docker.sock" // default

		// For Colima and other non-standard setups, extract socket path from DOCKER_HOST
		if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost != "" {
			if strings.HasPrefix(dockerHost, "unix://") {
				dockerSocket = strings.TrimPrefix(dockerHost, "unix://")
			}
		}

		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", dockerSocket)
		if verbose {
			fmt.Fprintf(os.Stderr, "testctr: Set TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=%s\n", dockerSocket)
		}
	}

	return 0
}

// SkipIfDockerNotAvailable checks if Docker is available and skips the test if not.
// This is safe to call from tests that use t.Parallel() as it doesn't use t.Setenv.
// You must ensure the environment variables are set before running the test (e.g., via EnsureTestEnv in TestMain).
func SkipIfDockerNotAvailable(t *testing.T) {
	t.Helper()

	// Check if docker is available in PATH
	_, err := exec.LookPath("docker")
	if err != nil {
		// If DOCKER_HOST is set, assume Docker is available remotely
		if os.Getenv("DOCKER_HOST") == "" {
			t.Skip("Docker not found in PATH and DOCKER_HOST not set, skipping test")
			return
		}
		// Docker CLI not found but DOCKER_HOST is set, so continue
	}

	if err := pingDaemon(); err != nil {
		t.Skipf("Docker daemon not reachable, skipping test: %v", err)
	}
}

// pingDaemon reports whether the Docker daemon answers on its local socket.
// Without this check, testcontainers panics in MustExtractDockerHost when the
// configured context points at a stopped daemon (for example, a stopped Colima VM).
// Non-unix hosts are not checked.
var pingDaemon = sync.OnceValue(func() error {
	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		host = "unix:///var/run/docker.sock"
		out, err := exec.Command("docker", "context", "inspect", "-f={{.Endpoints.docker.Host}}").Output()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			host = strings.TrimSpace(string(out))
		}
	}
	socket, ok := strings.CutPrefix(host, "unix://")
	if !ok {
		return nil
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socket)
			},
		},
	}
	resp, err := client.Get("http://docker/_ping")
	if err != nil {
		return fmt.Errorf("ping %s: %w", host, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping %s: %s", host, resp.Status)
	}
	return nil
})
