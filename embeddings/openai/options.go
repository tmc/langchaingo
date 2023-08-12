package openai

import (
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
)

// Option is a function type that can be used to modify the client.
type Option func(p *OpenAI)

// WithClient is an option for providing the LLM client.
func WithClient(client openai.LLM) Option {
	return func(p *OpenAI) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *OpenAI) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *OpenAI) {
		p.BatchSize = batchSize
	}
}

func applyClientOptions(opts ...Option) (OpenAI, error) {
	o := &OpenAI{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := openai.New()
		if err != nil {
			return OpenAI{}, err
		}
		o.client = client
	}

	return *o, nil
}
