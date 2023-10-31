package ollama

import (
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
)

// Option is a function type that can be used to modify the client.
type Option func(p *Ollama)

// WithClient is an option for providing the LLM client.
func WithClient(client ollama.LLM) Option {
	return func(p *Ollama) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *Ollama) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *Ollama) {
		p.BatchSize = batchSize
	}
}

func applyClientOptions(opts ...Option) (Ollama, error) {
	o := &Ollama{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := ollama.New()
		if err != nil {
			return Ollama{}, err
		}
		o.client = client
	}

	return *o, nil
}
