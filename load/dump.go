package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidFileSuffix = errors.New("invalid file suffix")
	ErrNoDataToSerialize = errors.New("no data to serialize")
)

func (s *FileSerializer) ToFile(data any, path string) error {
	if path == "" {
		return ErrInvalidSavePath
	}

	if reflect.ValueOf(data).IsZero() {
		return ErrNoDataToSerialize
	}

	suffix := s.FileSystem.NormalizeSuffix(path)
	switch strings.ToLower(suffix) {
	case ".json":
		return s.toJSON(data, path)
	case ".yaml", ".yml":
		return s.toYAML(data, path)
	case "":
		return s.toJSON(data, path+".json")
	default:
		return fmt.Errorf("%w:%s", ErrInvalidFileSuffix, suffix)
	}
}

func (s *FileSerializer) toJSON(d any, path string) error {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	err = s.FileSystem.Write(path, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileSerializer) toYAML(d any, path string) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return fmt.Errorf("failed to serialize YAML: %w", err)
	}

	err = s.FileSystem.Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
