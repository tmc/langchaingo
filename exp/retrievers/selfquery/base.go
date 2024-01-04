package selfquery

import (
	"github.com/tmc/langchaingo/schema"
)

var _ schema.Retriever = SelfQueryRetriever{}

type SelfQueryRetriever struct {
}
