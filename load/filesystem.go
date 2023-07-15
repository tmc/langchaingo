package load

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrWritingToFile   = errors.New("error writing to file")
	ErrInvalidSavePath = errors.New("invalid save path")
)

const filePermission = 0o600

type FileSystem interface {
	Write(path string, data []byte) error
	Read(path string) ([]byte, error)
	NormalizeSuffix(path string) string
}

type LocalFileSystem struct {
	FS FileSystem
}

func (f *LocalFileSystem) Write(path string, data []byte) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	err = f.makeDirectoriesIfNeeded(absPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(absPath, data, filePermission)
	if err != nil {
		return fmt.Errorf("failed writing to file: %w", err)
	}

	return nil
}

func (f *LocalFileSystem) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *LocalFileSystem) makeDirectoriesIfNeeded(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		dir := filepath.Dir(absPath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create path directories: %w", err)
		}
	}
	return nil
}

func (f *LocalFileSystem) NormalizeSuffix(path string) string {
	return filepath.Ext(path)
}
