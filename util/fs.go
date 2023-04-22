package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LookPath expands the functionality of exec.LookPath() by additionally expanding
// environment variables and tilde in the path.
//
// The following formats are supported:
// /home/xxx/path/to/binary_name   - Full path.
// $VAR/path/to/binary_name        - Env variable expansion.
// ~/path/to/binary_name           - Tilde expansion.
// ./path/to/binary_name           - Relative path.
// binary_name                     - $PATH lookup.
func LookPath(path string) (string, error) {
	// Expand tilde in the path
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		path = filepath.Join(homeDir, strings.TrimPrefix(path, "~"))
	}

	// Expand environment variables in the path
	path = os.ExpandEnv(path)

	// Ensure the path is valid
	path, err := exec.LookPath(path)
	if err != nil {
		return "", err
	}

	return path, nil
}
