package textsplitter

// Option is a function that configures Options.
type Option func(*Options)

// Options is a set of options for configuring a text splitter.
type Options struct {
	Separators   []string
	ChunkSize    int
	ChunkOverlap int
}

// WithSeparators is an option for setting the separators.
func WithSeparators(separators []string) Option {
	return func(o *Options) {
		o.Separators = separators
	}
}

// WithChunkSize is an option for setting the chunk size.
func WithChunkSize(size int) Option {
	return func(o *Options) {
		o.ChunkSize = size
	}
}

// WithChunkSize is an option for setting the chunk overlap.
func WithChunkOverlap(overlap int) Option {
	return func(o *Options) {
		o.ChunkOverlap = overlap
	}
}
