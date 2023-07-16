package prompts

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/tmc/langchaingo/load"
)

var (
	ErrInvalidWriteParams = errors.New("prompt template cannot be saved with partial variables")
	ErrInvalidSavePath    = errors.New("invalid save path")
)

type MockFileSystem struct {
	FS      load.FileSystem
	Storage map[string][]byte // map of file paths to file data
}

// uses the mock filesystem to write a file to the storage map.
func (f *MockFileSystem) Write(path string, data []byte) error {
	// return error if data is not nil or path is empty
	if data == nil || path == "" {
		return ErrInvalidWriteParams
	}
	f.Storage[path] = data
	return nil
}

func (f *MockFileSystem) NormalizeSuffix(path string) string {
	return filepath.Ext(path)
}

// lookup the file path in the storage map and return the file data.
func (f *MockFileSystem) Read(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("failed to read file: %w", ErrInvalidSavePath)
	}
	template, ok := f.Storage[path]
	if !ok {
		return nil, fmt.Errorf("failed to read file: %w", ErrInvalidSavePath)
	}
	return template, nil
}
