package documentLoaders

import (
	"github.com/tmc/langchaingo/exp/schema"
	"github.com/tmc/langchaingo/exp/textSplitters"
)

type DocumentLoader interface {
	Load() ([]schema.Document, error)
	LoadAndSplit(textSplitters.TextSplitter)
}
