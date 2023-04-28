package documentloader

import (
	"os"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// Text loads text data from a text file.
type Text struct {
	filePath string
}

var _ Loader = Text{}

// NewText creates a new text loader from the filepath of the text file.
func NewText(filePath string) Text {
	return Text{
		filePath: filePath,
	}
}

// Load reads the text file and returns a single document with the text data. The
// document includes metadata about the file path in the "source" key.
func (l Text) Load() ([]schema.Document, error) {
	fileData, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, err
	}

	return []schema.Document{
		{
			PageContent: string(fileData),
			Metadata:    map[string]any{"source": l.filePath},
		},
	}, nil
}

// LoadAndSplit reads the text file and splits it using a text splitter.
func (l Text) LoadAndSplit(splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load()
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}
