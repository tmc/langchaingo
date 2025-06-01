// Package testutil provides utilities for testing.
package testutil

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// SetupDockerEnvironment detects if Docker is present and if it's using Colima,
// then sets the appropriate environment variables for testcontainers.
// This is automatically called in test init functions.
func SetupDockerEnvironment(t *testing.T) {
	t.Helper()

	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		t.Logf("Docker not found in PATH, skipping Docker environment setup")
		return
	}

	// Run docker context inspect
	cmd := exec.Command("docker", "context", "inspect", "-f", "{{.Endpoints.docker.Host}}")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If docker context inspect fails, try the default
		t.Logf("Failed to inspect docker context: %v, using defaults", err)
		setDefaultDockerEnv()
		return
	}

	dockerHost := strings.TrimSpace(out.String())
	if dockerHost == "" {
		setDefaultDockerEnv()
		return
	}

	// Check if it's Colima based on the socket path
	if strings.Contains(dockerHost, "colima") || strings.Contains(dockerHost, ".colima") {
		t.Logf("Detected Colima Docker setup, configuring environment")
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
		t.Logf("Set DOCKER_HOST=%s", dockerHost)
		t.Logf("Set TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock")
	} else if dockerHost != "" {
		t.Logf("Detected Docker with host: %s", dockerHost)
		os.Setenv("DOCKER_HOST", dockerHost)
		// For non-Colima setups, we might not need the socket override
		if dockerHost != "unix:///var/run/docker.sock" {
			os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
		}
	}
}

// SetupDockerEnvironmentNoLog is the same as SetupDockerEnvironment but without logging.
// Useful for init() functions where we don't have access to testing.T.
func SetupDockerEnvironmentNoLog() {
	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		return
	}

	// Run docker context inspect
	cmd := exec.Command("docker", "context", "inspect", "-f", "{{.Endpoints.docker.Host}}")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		setDefaultDockerEnv()
		return
	}

	dockerHost := strings.TrimSpace(out.String())
	if dockerHost == "" {
		setDefaultDockerEnv()
		return
	}

	// Check if it's Colima based on the socket path
	if strings.Contains(dockerHost, "colima") || strings.Contains(dockerHost, ".colima") {
		os.Setenv("DOCKER_HOST", dockerHost)
		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
	} else if dockerHost != "" {
		os.Setenv("DOCKER_HOST", dockerHost)
		if dockerHost != "unix:///var/run/docker.sock" {
			os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
		}
	}
}

func setDefaultDockerEnv() {
	// Only set if not already set
	if os.Getenv("DOCKER_HOST") == "" {
		os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")
	}
}
