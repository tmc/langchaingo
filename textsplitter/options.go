package textsplitter

import "unicode/utf8"

// Options is a struct that contains options for a text splitter.
type Options struct {
	ChunkSize         int
	ChunkOverlap      int
	Separators        []string
	LenFunc           func(string) int
	ModelName         string
	EncodingName      string
	AllowedSpecial    []string
	DisallowedSpecial []string
	SecondSplitter    TextSplitter
	CodeBlocks        bool
	ReferenceLinks    bool
}

// DefaultOptions returns the default options for all text splitter.
func DefaultOptions() Options {
	return Options{
		ChunkSize:    _defaultTokenChunkSize,
		ChunkOverlap: _defaultTokenChunkOverlap,
		Separators:   []string{"\n\n", "\n", " ", ""},
		LenFunc:      utf8.RuneCountInString,

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

// WithLenFunc sets the lenfunc for a text splitter.
func WithLenFunc(lenFunc func(string) int) Option {
	return func(o *Options) {
		o.LenFunc = lenFunc
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

// WithCodeBlocks sets whether indented and fenced codeblocks should be included
// in the output.
func WithCodeBlocks(renderCode bool) Option {
	return func(o *Options) {
		o.CodeBlocks = renderCode
	}
}

// WithReferenceLinks sets whether reference links (i.e. `[text][label]`)
// should be patched with the url and title from their definition. Note that
// by default reference definitions are dropped from the output.
//
// Caution: this also affects how other inline elements are rendered, e.g. all
// emphasis will use `*` even when another character (e.g. `_`) was used in the
// input.
func WithReferenceLinks(referenceLinks bool) Option {
	return func(o *Options) {
		o.ReferenceLinks = referenceLinks
	}
}
