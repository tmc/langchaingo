package jina

import (
	"net/http"
	"os"

	"github.com/0xDezzy/langchaingo/httputil"
)

const (
	_defaultStripNewLines = true
	_defaultModel         = "jina-embeddings-v2-small-en"
	_defaultTask          = "feature-extraction"
	SmallModel            = "jina-embeddings-v2-small-en"
	BaseModel             = "jina-embeddings-v2-base-en"
	LargeModel            = "jina-embeddings-v2-large-en"
	APIBaseURL            = "https://api.jina.ai/v1/embeddings"
)

// Option is a function type that can be used to modify the client.
type Option func(p *Jina)

// WithModel is an option for providing the model name to use.
func WithModel(model string) Option {
	return func(p *Jina) {
		p.Model = model
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *Jina) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *Jina) {
		p.BatchSize = batchSize
	}
}

// WithAPIBaseURL is an option for specifying the API base URL.
func WithAPIBaseURL(apiBaseURL string) Option {
	return func(p *Jina) {
		p.APIBaseURL = apiBaseURL
	}
}

// WithAPIKey is an option for specifying the API key.
func WithAPIKey(apiKey string) Option {
	return func(p *Jina) {
		p.APIKey = apiKey
	}
}

// WithClient is an option for providing a custom HTTP client.
func WithClient(client *http.Client) Option {
	return func(p *Jina) {
		p.client = client
	}
}

func applyOptions(opts ...Option) *Jina {
	_models := map[string]int{
		"jina-embeddings-v2-small-en": 512,
		"jina-embeddings-v2-base-en":  768,
		"jina-embeddings-v2-large-en": 1024,
	}

	o := &Jina{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _models[_defaultModel],
		Model:         _defaultModel,
		APIBaseURL:    APIBaseURL,
		APIKey:        os.Getenv("JINA_API_KEY"),
		client:        httputil.DefaultClient,
	}

	for _, opt := range opts {
		opt(o)
	}

	// verify if model exists in the map
	if _, ok := _models[o.Model]; ok {
		o.BatchSize = _models[o.Model]
	}

	return o
}
