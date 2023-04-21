package document_loaders

import (
	"github.com/tmc/langchaingo/exp/text_splitters"
	"github.com/tmc/langchaingo/schema"
)

// Document loader is the interface for loading and splitting documents from a source
type DocumentLoader interface {
	// Loads from source and returns documents
	Load() ([]schema.Document, error)
	// Loads from source and splits using a text splitter
	LoadAndSplit(text_splitters.TextSplitter)
}
