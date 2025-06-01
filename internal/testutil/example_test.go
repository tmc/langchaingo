package testutil_test

import (
	"testing"

	"github.com/tmc/langchaingo/internal/testutil"
)

// Example of using SetupDockerEnvironment in a test function
func TestExampleWithDocker(t *testing.T) {
	// This automatically sets up Docker environment variables
	testutil.SetupDockerEnvironment(t)

	// Your test code that uses testcontainers would go here
	t.Log("Docker environment has been configured")
}

// Example of package-level initialization for all tests in a package
func init() {
	// This sets up Docker environment for all tests in the package
	// Use this when you have multiple tests that need Docker
	testutil.SetupDockerEnvironmentNoLog()
}
