package huggingface

import (
	"github.com/tmc/langchaingo/llms/huggingface"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
	_defaultModel         = "sentence-transformers/all-mpnet-base-v2"
	_defaultTask          = "feature-extraction"
)

// Option is a function type that can be used to modify the client.
type Option func(p *Huggingface)

// WithModel is an option for providing the model name to use.
func WithModel(model string) Option {
	return func(p *Huggingface) {
		p.Model = model
	}
}

// WithTask is an option for providing the task to call the model with.
func WithTask(task string) Option {
	return func(p *Huggingface) {
		p.Task = task
	}
}

// WithClient is an option for providing the LLM client.
func WithClient(client huggingface.LLM) Option {
	return func(p *Huggingface) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *Huggingface) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *Huggingface) {
		p.BatchSize = batchSize
	}
}

func applyOptions(opts ...Option) (*Huggingface, error) {
	o := &Huggingface{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
		Model:         _defaultModel,
		Task:          _defaultTask,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := huggingface.New()
		if err != nil {
			return nil, err
		}
		o.client = client
	}

	return o, nil
}
