package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrJSONSerializationFailed = errors.New("failed to serialize to JSON")
	ErrYAMLSerializationFailed = errors.New("failed to serialize to YAML")
	ErrInvalidSavePath         = errors.New("invalid prompt template save path")
	ErrCreatingSavePath        = errors.New("error creating prompt template save path")
	ErrInvalidFileSuffix       = errors.New("invalid file suffix")
	ErrWritingToFile           = errors.New("error writing to file")
)

func ToFile(s any, path string) error {
	suffix := getFileSuffix(path)
	switch suffix {
	case ".json":
		return ToJSON(s, path)
	case ".yaml":
		return ToYAML(s, path)
	case "":
		return ToJSON(s, path)
	default:
		return ErrInvalidFileSuffix
	}
}

func ToJSON(s any, path string) error {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJSONSerializationFailed, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSavePath, err)
	}

	err = MakeDirectoriesIfNeeded(absPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWritingToFile, err)
	}

	return nil
}

func MakeDirectoriesIfNeeded(absPath string) error {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		dir := filepath.Dir(absPath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCreatingSavePath, err)
		}
	}
	return nil
}

func ToYAML(s any, path string) error {
	// TODO: implement
	return nil
}

func getFileSuffix(path string) string {
	index := strings.LastIndex(path, ".")
	if index != -1 && index != len(path)-1 {
		return path[index:]
	}
	return ""
}
