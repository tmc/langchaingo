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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
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

			// Only set DOCKER_HOST if using Colima (works around testcontainers bug)
			if dockerHost != "" && strings.Contains(dockerHost, "colima") {
				os.Setenv("DOCKER_HOST", dockerHost)
			}
		}
		// If docker context inspect fails, just continue without setting DOCKER_HOST
		// This is common when using standard Docker Desktop
	}

	// Set the testcontainers Docker socket override if not already set
	if os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE") == "" {
		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
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
}
