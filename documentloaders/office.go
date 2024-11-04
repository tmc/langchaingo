package documentloaders

import (
	// "archive/zip"
	// "bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	// "github.com/richardlehane/mscfb"
	// "github.com/tealeg/xlsx"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

var _ Loader = Office{}

type Office struct {
	reader   io.ReaderAt
	size     int64
	fileType string
}

func NewOffice(reader io.ReaderAt, size int64, filename string) Office {
	return Office{
		reader:   reader,
		size:     size,
		fileType: strings.ToLower(filepath.Ext(filename)),
	}
}

func (loader Office) Load(ctx context.Context) ([]schema.Document, error) {
	switch loader.fileType {
	case ".doc", ".docx":
		return loader.loadWord()
	case ".xls", ".xlsx":
		return loader.loadExcel()
	case ".ppt", ".pptx":
		return loader.loadPowerPoint()
	default:
		return nil, fmt.Errorf("unsupported file type: %s", loader.fileType)
	}
}

func (loader Office) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}

func (loader Office) loadWord() ([]schema.Document, error) {
	return nil, nil
}

func (loader Office) loadExcel() ([]schema.Document, error) {
	return nil, nil
}

func (loader Office) loadPowerPoint() ([]schema.Document, error) {
	return nil, nil
}
