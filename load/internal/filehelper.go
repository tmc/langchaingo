package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrCreatingSavePath = errors.New("error creating prompt template save path")
	ErrWritingToFile    = errors.New("error writing to file")
)

type FileSystem interface {
	Write(path string, data []byte) error
	Read(path string) ([]byte, error)
}

type FileHelper struct {
	FileSystem FileSystem
}

func (f *FileHelper) Write(path string, data []byte) error {
	err := f.makeDirectoriesIfNeeded(path)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrWritingToFile, err)
	}

	return nil
}

func (f *FileHelper) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *FileHelper) makeDirectoriesIfNeeded(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		dir := filepath.Dir(absPath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("%s: %w", ErrCreatingSavePath, err)
		}
	}
	return nil
}
