package textsplitter

// Options is a struct that contains options for a text splitter.
type Options struct {
	ChunkSize         int
	ChunkOverlap      int
	Separators        []string
	ModelName         string
	EncodingName      string
	AllowedSpecial    []string
	DisallowedSpecial []string
	SecondSplitter    TextSplitter
}

// DefaultOptions returns the default options for all text splitter.
func DefaultOptions() Options {
	return Options{
		ChunkSize:    _defaultTokenChunkSize,
		ChunkOverlap: _defaultTokenChunkOverlap,
		Separators:   []string{"\n\n", "\n", " ", ""},

		ModelName:         _defaultTokenModelName,
		EncodingName:      _defaultTokenEncoding,
		AllowedSpecial:    []string{},
		DisallowedSpecial: []string{"all"},
	}
}

// Option is a function that can be used to set options for a text splitter.
type Option func(*Options)

// WithChunkSize sets the chunk size for a text splitter.
func WithChunkSize(chunkSize int) Option {
	return func(o *Options) {
		o.ChunkSize = chunkSize
	}
}

// WithChunkOverlap sets the chunk overlap for a text splitter.
func WithChunkOverlap(chunkOverlap int) Option {
	return func(o *Options) {
		o.ChunkOverlap = chunkOverlap
	}
}

// WithSeparators sets the separators for a text splitter.
func WithSeparators(separators []string) Option {
	return func(o *Options) {
		o.Separators = separators
	}
}

// WithModelName sets the model name for a text splitter.
func WithModelName(modelName string) Option {
	return func(o *Options) {
		o.ModelName = modelName
	}
}

// WithEncodingName sets the encoding name for a text splitter.
func WithEncodingName(encodingName string) Option {
	return func(o *Options) {
		o.EncodingName = encodingName
	}
}

// WithAllowedSpecial sets the allowed special tokens for a text splitter.
func WithAllowedSpecial(allowedSpecial []string) Option {
	return func(o *Options) {
		o.AllowedSpecial = allowedSpecial
	}
}

// WithDisallowedSpecial sets the disallowed special tokens for a text splitter.
func WithDisallowedSpecial(disallowedSpecial []string) Option {
	return func(o *Options) {
		o.DisallowedSpecial = disallowedSpecial
	}
}

// WithSecondSplitter sets the second splitter for a text splitter.
func WithSecondSplitter(secondSplitter TextSplitter) Option {
	return func(o *Options) {
		o.SecondSplitter = secondSplitter
	}
}
