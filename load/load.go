package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrInvalidPath = errors.New("invalid file path")

func FromFile(data any, path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	suffix := NormalizeSuffix(path)
	switch strings.ToLower(suffix) {
	case ".json":
		return fromJSON(data, path)
	case ".yaml", ".yml":
		return fromYAML(data, path)
	default:
		return fmt.Errorf("%w:%s", ErrInvalidPath, suffix)
	}
}

func fromJSON(data any, path string) error {
	byteData, err := readFrom(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, data)
	if err != nil {
		return fmt.Errorf("failed to deserialize JSON: %w", err)
	}

	return nil
}

func fromYAML(data any, path string) error {
	byteData, err := readFrom(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(byteData, data)
	if err != nil {
		return fmt.Errorf("failed to deserialize JSON: %w", err)
	}

	return nil
}

func readFrom(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	byteData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return byteData, nil
}
