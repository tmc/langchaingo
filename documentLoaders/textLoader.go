package documentLoaders

import (
	"os"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textSplitters"
)

type TextLoader struct {
	filePath string
	//Todo: blob equivalent
}

func NewTextLoaderFromFile(filePath string) TextLoader {
	return TextLoader{
		filePath: filePath,
	}
}

func (l TextLoader) Load() ([]schema.Document, error) {
	fileData, err := os.ReadFile(l.filePath)
	if err != nil {
		return []schema.Document{}, err
	}

	return []schema.Document{
		{
			PageContent: string(fileData),
			Metadata:    map[string]any{"source": l.filePath},
		},
	}, nil
}

func (l TextLoader) LoadAndSplit(splitter textSplitters.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load()
	if err != nil {
		return []schema.Document{}, err
	}

	return textSplitters.SplitDocuments(splitter, docs)
}
