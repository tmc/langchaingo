package document_loaders

import (
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// Document loader is the interface for loading and splitting documents from a source.
type DocumentLoader interface {
	// Loads from source and returns documents
	Load() ([]schema.Document, error)
	// Loads from source and splits using a text splitter
	LoadAndSplit(textsplitter.TextSplitter)
}
