package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type lookPathTest struct {
	input          string
	expectedOutput string
	expectsError   bool
	errorString    string
}

var lookPathTests = []lookPathTest{
	// Full path
	{
		input:          "/bin/ls",
		expectedOutput: "/bin/ls",
		expectsError:   false,
	},

	// Environment variable expansion
	{
		input:          "$TEST_DIR/hello",
		expectedOutput: getCallerPath() + "/testdata/hello",
		expectsError:   false,
	},

	// Tilde expansion
	{
		input:          "~",
		expectedOutput: "",
		expectsError:   true,
		errorString:    "is a directory",
	},

	// Relative path
	{
		input:          "./testdata/hello",
		expectedOutput: "./testdata/hello",
		expectsError:   false,
	},

	// $PATH lookup
	{
		input:          "ls",
		expectedOutput: findExecutable("ls"),
		expectsError:   false,
	},

	// Invalid path
	{
		input:          "invalid/path/to/binary",
		expectedOutput: "",
		expectsError:   true,
		errorString:    "no such file or directory",
	},

	// Invalid environment variable
	{
		input:          "$INVALID_ENV_VAR/hello",
		expectedOutput: "",
		expectsError:   true,
		errorString:    "no such file or directory",
	},
}

func TestLookPath(t *testing.T) {
	// Set an environment variable for testing
	os.Setenv("TEST_DIR", getCallerPath()+"/testdata")

	for _, test := range lookPathTests {
		path, err := LookPath(test.input)

		if !test.expectsError && err != nil {
			t.Errorf("Unexpected error for input %s: %s", test.input, err.Error())
		}
		if test.expectsError && err == nil {
			t.Errorf("Expected error for input %s, but got none", test.input)
			if err.Error() != test.errorString {
				t.Errorf("Expected error string %s, got %s", test.errorString, err.Error())
			}
		}
		if path != test.expectedOutput {
			t.Errorf("LookPath(%s) returned %s, expected %s", test.input, path, test.expectedOutput)
		}
	}
}

func getCallerPath() string {
	// Get caller path
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func findExecutable(binary string) string {
	// Find the binary in $PATH
	path, err := exec.LookPath(binary)
	if err != nil {
		panic(err)
	}
	return path
}
