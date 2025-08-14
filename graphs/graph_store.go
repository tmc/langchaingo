package graphs

import "context"

// GraphStore defines the interface for graph database operations.
type GraphStore interface {
	// AddGraphDocuments adds graph documents to the store.
	AddGraphDocuments(ctx context.Context, docs []GraphDocument, options ...Option) error

	// Query executes a query against the graph store and returns results.
	Query(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error)

	// RefreshSchema refreshes the schema information from the graph database.
	RefreshSchema(ctx context.Context) error

	// GetSchema returns the current schema as a string representation.
	GetSchema() string

	// GetStructuredSchema returns the structured schema information.
	GetStructuredSchema() map[string]interface{}

	// Close closes the connection to the graph store.
	Close() error
}

// Option defines functional options for graph store operations.
type Option func(*Options)

// Options contains configuration options for graph store operations.
type Options struct {
	// IncludeSource indicates whether to include source document information
	IncludeSource bool
	// BatchSize specifies the batch size for bulk operations
	BatchSize int
	// Timeout specifies query timeout in milliseconds
	Timeout int
}

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return &Options{
		IncludeSource: false,
		BatchSize:     100,
		Timeout:       0, // No timeout by default
	}
}

// WithIncludeSource sets whether to include source document information.
func WithIncludeSource(include bool) Option {
	return func(o *Options) {
		o.IncludeSource = include
	}
}

// WithBatchSize sets the batch size for bulk operations.
func WithBatchSize(size int) Option {
	return func(o *Options) {
		o.BatchSize = size
	}
}

// WithTimeout sets the query timeout in milliseconds.
func WithTimeout(timeout int) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}
