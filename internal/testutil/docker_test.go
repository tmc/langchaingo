package testutil

import (
	"os"
	"testing"
)

func TestSetupDockerEnvironment(t *testing.T) {
	// Save original environment
	origDockerHost := os.Getenv("DOCKER_HOST")
	origSocketOverride := os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
	defer func() {
		// Restore original environment
		if origDockerHost != "" {
			os.Setenv("DOCKER_HOST", origDockerHost)
		} else {
			os.Unsetenv("DOCKER_HOST")
		}
		if origSocketOverride != "" {
			os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", origSocketOverride)
		} else {
			os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		}
	}()

	// Clear environment for test
	os.Unsetenv("DOCKER_HOST")
	os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")

	// Run the setup
	SetupDockerEnvironment(t)

	// Check if environment variables were set (if Docker is available)
	// We can't test the exact values as they depend on the system
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost != "" {
		t.Logf("DOCKER_HOST was set to: %s", dockerHost)
	}

	socketOverride := os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
	if socketOverride != "" {
		t.Logf("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE was set to: %s", socketOverride)
	}
}

func TestSetupDockerEnvironmentNoLog(t *testing.T) {
	// Save original environment
	origDockerHost := os.Getenv("DOCKER_HOST")
	origSocketOverride := os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
	defer func() {
		// Restore original environment
		if origDockerHost != "" {
			os.Setenv("DOCKER_HOST", origDockerHost)
		} else {
			os.Unsetenv("DOCKER_HOST")
		}
		if origSocketOverride != "" {
			os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", origSocketOverride)
		} else {
			os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		}
	}()

	// Clear environment for test
	os.Unsetenv("DOCKER_HOST")
	os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")

	// Run the setup
	SetupDockerEnvironmentNoLog()

	// Check if environment variables were set (if Docker is available)
	dockerHost := os.Getenv("DOCKER_HOST")
	socketOverride := os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
	
	// If Docker is available, at least DOCKER_HOST should be set
	if dockerHost == "" && socketOverride == "" {
		t.Log("No Docker environment variables were set (Docker might not be available)")
	} else {
		t.Logf("Docker environment configured: DOCKER_HOST=%s", dockerHost)
		if socketOverride != "" {
			t.Logf("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=%s", socketOverride)
		}
	}
}