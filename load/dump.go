package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/load/internal"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
)

var (
	ErrJSONSerializationFailed = errors.New("failed to serialize to JSON")
	ErrYAMLSerializationFailed = errors.New("failed to serialize to YAML")
	ErrInvalidSavePath         = errors.New("invalid prompt template save path")
	ErrInvalidFileSuffix       = errors.New("invalid file suffix")
)

type Serializer interface {
	ToFile(d any, path string) error
}

type FileSerializer struct {
	FileHelper *internal.FileHelper
}

func NewSerializer() *FileSerializer {
	return &FileSerializer{
		FileHelper: &internal.FileHelper{},
	}
}

func (s *FileSerializer) ToFile(d any, path string) error {
	suffix := getSuffix(path)
	switch suffix {
	case ".json":
		return s.ToJSON(d, path)
	case ".yaml":
		return s.ToYAML(d, path)
	case "":
		return s.ToJSON(d, path)
	default:
		return ErrInvalidFileSuffix
	}
}

func (s *FileSerializer) ToJSON(d any, path string) error {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrJSONSerializationFailed, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInvalidSavePath, err)
	}

	err = s.FileHelper.Write(absPath, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileSerializer) ToYAML(d any, path string) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrYAMLSerializationFailed, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInvalidSavePath, err)
	}

	err = s.FileHelper.Write(absPath, data)
	if err != nil {
		return err
	}

	return nil
}

func getSuffix(path string) string {
	index := strings.LastIndex(path, ".")
	if index != -1 && index != len(path)-1 {
		return path[index:]
	}
	return ""
}
