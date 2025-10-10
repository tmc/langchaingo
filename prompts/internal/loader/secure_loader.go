// Package loader provides secure filesystem access control for template engines.
//
// Security considerations:
// - Always validate custom fs.FS implementations before use
// - Ensure path traversal attacks are prevented at the fs.FS level
// - Consider implementing audit logging for all filesystem access
// - Use io/fs.Sub to create restricted filesystem views
package loader

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// ErrFilesystemAccessDisabled is the error returned when filesystem access is disabled for security.
var ErrFilesystemAccessDisabled = errors.New("template loading from filesystem disabled for security reasons")

// NilFSLoader is a template loader that provides no filesystem access.
// This prevents template injection attacks like {% include "/etc/passwd" %}.
type NilFSLoader struct{}

// Get always returns an error to prevent filesystem access.
func (nl *NilFSLoader) Get(path string) (io.Reader, error) {
	return nil, ErrFilesystemAccessDisabled
}

// Path always returns an error to prevent filesystem access.
func (nl *NilFSLoader) Path(path string) (string, error) {
	return "", ErrFilesystemAccessDisabled
}

// FSLoader wraps an fs.FS to provide controlled filesystem access.
// This is used internally when explicit filesystem access is needed.
//
// Security note: The provided fs.FS should be pre-validated to ensure:
// - It doesn't allow access outside intended boundaries
// - Path traversal attempts are blocked
// - Symbolic links don't escape the filesystem root
// - Consider using fs.Sub() to create restricted views
type FSLoader struct {
	filesystem fs.FS
}

// NewFSLoader creates a new FSLoader with the provided filesystem.
func NewFSLoader(filesystem fs.FS) *FSLoader {
	return &FSLoader{filesystem: filesystem}
}

// Get opens a file from the controlled filesystem.
// The path is passed directly to the underlying fs.FS, which should
// handle its own path validation and normalization.
func (fl *FSLoader) Get(path string) (io.Reader, error) {
	// Clean the path to prevent simple traversal attempts
	// Note: The fs.FS implementation should provide additional security
	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	return fl.filesystem.Open(path)
}

// Path validates that a path exists in the controlled filesystem.
func (fl *FSLoader) Path(path string) (string, error) {
	// Validate path before checking
	if err := validatePath(path); err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	_, err := fl.filesystem.Open(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

// validatePath performs basic path validation to prevent obvious attacks.
// The fs.FS implementation should provide additional security measures.
func validatePath(path string) error {
	// Reject paths with null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}
	// Reject absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Use filepath.Rel to ensure the path doesn't escape the base directory
	baseDir := "."
	cleanPath := filepath.Clean(path)
	relPath, err := filepath.Rel(baseDir, cleanPath)
	if err != nil || strings.HasPrefix(relPath, "..") || relPath == ".." {
		return fmt.Errorf("path traversal detected")
	}

	// Reject paths starting with /
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("paths must be relative")
	}
	return nil
}
