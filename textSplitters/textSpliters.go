package textSplitters

import "github.com/tmc/langchaingo/schema"

type TextSplitter interface {
	SplitText(string) ([]string, error)
	SplitDocuments([]schema.Document) ([]schema.Document, error)
}
