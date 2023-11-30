package embeddings

const (
	defaultBatchSize     = 512
	defaultStripNewLines = true
)

type Option func(p *EmbedderImpl)

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *EmbedderImpl) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(p *EmbedderImpl) {
		p.BatchSize = batchSize
	}
}
