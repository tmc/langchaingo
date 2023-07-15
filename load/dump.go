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

type Serializer interface {
	ToFile(d any, path string) error
}

type FileSerializer struct {
	FileSystem FileSystem
}

func NewSerializer(fs FileSystem) *FileSerializer {
	return &FileSerializer{
		FileSystem: fs,
	}
}

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
		return s.ToJSON(data, path)
	case ".yaml", ".yml":
		return s.ToYAML(data, path)
	case "":
		return s.ToJSON(data, path+".json")
	default:
		return fmt.Errorf("%w:%s", ErrInvalidFileSuffix, suffix)
	}
}

func (s *FileSerializer) ToJSON(d any, path string) error {
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

func (s *FileSerializer) ToYAML(d any, path string) error {
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
