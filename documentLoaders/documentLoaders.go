package documentLoaders

import (
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textSplitters"
)

type DocumentLoader interface {
	Load() ([]schema.Document, error)
	LoadAndSplit(textSplitters.TextSplitter)
}
