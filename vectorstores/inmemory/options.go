package inmemory

import "github.com/tmc/langchaingo/embeddings"

const (
	defaultM              = 16
	defaultEfConstruction = 64
	defaultEfSearch       = 64

	defaultVectorSize = 128
	defaultSize       = 128
)

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithHNSWOptions is an option for specifying the HNSW index parameters.
// m: he max number of connections per layer (16 by default).
// efConstruction: the size of the dynamic candidate list for constructing the graph (64 by default).
// efSearch: the size of the dynamic candidate list for search (64 by default).
// vectorSize: the size of the vector (128 by default).
func WithHNSWOptions(m, efConstruction, efSearch int) Option {
	return func(s *Store) {
		s.m = m
		s.efConstruction = efConstruction
		s.efSearch = efSearch
	}
}

// WithSize is an option for setting the initial size of the store (128 by default).
func WithSize(limit int) Option {
	return func(s *Store) {
		s.sizeLimit = limit
	}
}

// WithVectorSize is an option for setting the size of the vector (128 by default).
func WithVectorSize(size int) Option {
	return func(s *Store) {
		s.vectorSize = size
	}
}

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(s *Store) {
		s.embedder = e
	}
}

func applyOptions(opts []Option) *Store {
	s := &Store{
		lastID:   0,
		embedder: nil,
		content:  make(map[uint32]string),
		meta:     make(map[uint32]map[string]any),

		m:              defaultM,
		efConstruction: defaultEfConstruction,
		efSearch:       defaultEfSearch,

		vectorSize: defaultVectorSize,
		sizeLimit:  defaultSize,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
