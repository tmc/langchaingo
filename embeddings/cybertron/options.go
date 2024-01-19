package cybertron

import (
	"github.com/nlpodyssey/cybertron/pkg/models/bert"
	"github.com/nlpodyssey/cybertron/pkg/tasks"
	"github.com/nlpodyssey/cybertron/pkg/tasks/textencoding"
)

const (
	_defaultModel           = "sentence-transformers/all-MiniLM-L6-v2"
	_defaultModelsDir       = "models"
	_defaultPoolingStrategy = bert.MeanPooling
)

// Option is a function type that can be used to modify the client.
type Option func(c *Cybertron)

// apply the option to the instance.
func (o Option) apply(c *Cybertron) {
	o(c)
}

// WithModel is an option for providing the model name to use. Default is
// "sentence-transformers/all-MiniLM-L6-v2". Note that not all embedding models
// are supported.
func WithModel(model string) Option {
	return func(c *Cybertron) {
		c.Model = model
	}
}

// WithModelsDir is an option for setting the directory to store downloaded models.
// Default is "models".
func WithModelsDir(dir string) Option {
	return func(c *Cybertron) {
		c.ModelsDir = dir
	}
}

// WithPoolingStrategy sets the pooling strategy. Default is mean pooling.
func WithPoolingStrategy(strategy bert.PoolingStrategyType) Option {
	return func(c *Cybertron) {
		c.PoolingStrategy = strategy
	}
}

// WithEncoder is an option for providing the Encoder.
func WithEncoder(encoder textencoding.Interface) Option {
	return func(c *Cybertron) {
		c.encoder = encoder
	}
}

func applyOptions(opts ...Option) (*Cybertron, error) {
	c := &Cybertron{
		Model:           _defaultModel,
		ModelsDir:       _defaultModelsDir,
		PoolingStrategy: _defaultPoolingStrategy,
		encoder:         nil,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	if c.encoder == nil {
		encoder, err := tasks.Load[textencoding.Interface](&tasks.Config{
			ModelsDir: c.ModelsDir,
			ModelName: c.Model,
		})
		if err != nil {
			return nil, err
		}

		c.encoder = encoder
	}

	return c, nil
}
