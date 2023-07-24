package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidFileSuffix = errors.New("invalid file suffix")
	ErrNoDataToSerialize = errors.New("no data to serialize")
	ErrInvalidSavePath   = errors.New("invalid save path")
)

const filePermission = 0o600

func ToFile(data any, path string) error {
	if path == "" {
		return ErrInvalidSavePath
	}

	if reflect.ValueOf(data).IsZero() {
		return ErrNoDataToSerialize
	}

	suffix := NormalizeSuffix(path)
	switch strings.ToLower(suffix) {
	case ".json":
		return toJSON(data, path)
	case ".yaml", ".yml":
		return toYAML(data, path)
	case "":
		return toJSON(data, path+".json")
	default:
		return fmt.Errorf("%w:%s", ErrInvalidFileSuffix, suffix)
	}
}

func toJSON(d any, path string) error {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	err = writeTo(path, data)
	if err != nil {
		return fmt.Errorf("failed writing to file: %w", err)
	}

	return nil
}

func toYAML(d any, path string) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return fmt.Errorf("failed to serialize YAML: %w", err)
	}

	err = writeTo(path, data)
	if err != nil {
		return fmt.Errorf("failed writing to file: %w", err)
	}

	return nil
}

func writeTo(path string, data []byte) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	err = makeDirectoriesIfNeeded(absPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(absPath, data, filePermission)
	if err != nil {
		return fmt.Errorf("failed writing to file: %w", err)
	}
	return nil
}

func makeDirectoriesIfNeeded(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		dir := filepath.Dir(absPath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create path directories: %w", err)
		}
	}
	return nil
}

func NormalizeSuffix(path string) string {
	return filepath.Ext(path)
}
