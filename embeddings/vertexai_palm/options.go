package vertexai_palm

import (
	"github.com/tmc/langchaingo/llms/vertexai"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
)

// Option is a function type that can be used to modify the client.
type Option func(p *VertexAIPaLM)

// WithClient is an option for providing the LLM client.
func WithClient(client vertexai.LLM) Option {
	return func(p *VertexAIPaLM) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *VertexAIPaLM) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *VertexAIPaLM) {
		p.BatchSize = batchSize
	}
}

func applyClientOptions(opts ...Option) (*VertexAIPaLM, error) {
	v := &VertexAIPaLM{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.client == nil {
		client, err := vertexai.New()
		if err != nil {
			return nil, err
		}
		v.client = client
	}

	return v, nil
}
