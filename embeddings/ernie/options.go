package ernie

import "github.com/tmc/langchaingo/llms/ernie"

const (
	// see: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu#body%E5%8F%82%E6%95%B0
	defaultBatchCount    = 16
	defaultBatchSize     = 384
	defaultStripNewLines = true
)

// Option is a function type that can be used to modify the client.
type Option func(p *Ernie)

// WithClient is an option for providing the LLM client.
func WithClient(client ernie.LLM) Option {
	return func(e *Ernie) {
		e.client = &client
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(e *Ernie) {
		e.batchSize = batchSize
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(e *Ernie) {
		e.stripNewLines = stripNewLines
	}
}
