package openai

import "github.com/sashabaranov/go-openai"

var stringToEmbeddingModel = map[string]openai.EmbeddingModel{ // nolint:gochecknoglobals
	"text-similarity-ada-001":       openai.AdaSimilarity,
	"text-similarity-babbage-001":   openai.BabbageSimilarity,
	"text-similarity-curie-001":     openai.CurieSimilarity,
	"text-similarity-davinci-001":   openai.DavinciSimilarity,
	"text-search-ada-doc-001":       openai.AdaSearchDocument,
	"text-search-ada-query-001":     openai.AdaSearchQuery,
	"text-search-babbage-doc-001":   openai.BabbageSearchDocument,
	"text-search-babbage-query-001": openai.BabbageSearchQuery,
	"text-search-curie-doc-001":     openai.CurieSearchDocument,
	"text-search-curie-query-001":   openai.CurieSearchQuery,
	"text-search-davinci-doc-001":   openai.DavinciSearchDocument,
	"text-search-davinci-query-001": openai.DavinciSearchQuery,
	"code-search-ada-code-001":      openai.AdaCodeSearchCode,
	"code-search-ada-text-001":      openai.AdaCodeSearchText,
	"code-search-babbage-code-001":  openai.BabbageCodeSearchCode,
	"code-search-babbage-text-001":  openai.BabbageCodeSearchText,
	"text-embedding-ada-002":        openai.AdaEmbeddingV2,
}
