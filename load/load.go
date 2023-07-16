package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrInvalidPath = errors.New("invalid file path")

func (s *FileSerializer) FromFile(data any, path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	suffix := s.FileSystem.NormalizeSuffix(path)
	switch strings.ToLower(suffix) {
	case ".json":
		return s.fromJSON(data, path)
	case ".yaml", ".yml":
		return s.fromYAML(data, path)
	default:
		return fmt.Errorf("%w:%s", ErrInvalidPath, suffix)
	}
}

func (s *FileSerializer) fromJSON(data any, path string) error {
	byteData, err := s.FileSystem.Read(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteData, data)
	if err != nil {
		return fmt.Errorf("failed to deserialize JSON: %w", err)
	}

	return nil
}

func (s *FileSerializer) fromYAML(data any, path string) error {
	byteData, err := s.FileSystem.Read(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(byteData, data)
	if err != nil {
		return fmt.Errorf("failed to deserialize JSON: %w", err)
	}

	return nil
}
