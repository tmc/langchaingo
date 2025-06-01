// Package testutil provides common utilities for tests in the langchaingo project.
//
// # Docker Environment Setup
//
// The Docker environment setup functions automatically detect if Docker is running
// with Colima and sets the appropriate environment variables for testcontainers.
//
// Usage in test files:
//
//	func TestMyFunction(t *testing.T) {
//	    testutil.SetupDockerEnvironment(t)
//	    // ... rest of your test that uses testcontainers
//	}
//
// For package-level initialization:
//
//	func init() {
//	    testutil.SetupDockerEnvironmentNoLog()
//	}
//
// The setup functions will:
//   - Check if Docker is available
//   - Run `docker context inspect` to detect the Docker host
//   - If Colima is detected (by checking for "colima" in the socket path),
//     set DOCKER_HOST and TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE
//   - Otherwise, set appropriate defaults
//
// This eliminates the need to manually set these environment variables
// in the Makefile or before running tests.
package testutil
