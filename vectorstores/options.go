package vectorstores

import (
	"context"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
)

// Option is a function that configures an Options.
type Option func(*Options)

// Options is a set of options for similarity search and add documents.
type Options struct {
	NameSpace           string
	ScoreThreshold      float32
	Filters             any
	Embedder            embeddings.Embedder
	Deduplicater        func(context.Context, schema.Document) bool
	DocumentIDGenerater func(context.Context, schema.Document) string
}

// WithNameSpace returns an Option for setting the name space.
func WithNameSpace(nameSpace string) Option {
	return func(o *Options) {
		o.NameSpace = nameSpace
	}
}

func WithScoreThreshold(scoreThreshold float32) Option {
	return func(o *Options) {
		o.ScoreThreshold = scoreThreshold
	}
}

// WithFilters searches can be limited based on metadata filters. Searches with  metadata
// filters retrieve exactly the number of nearest-neighbors results that match the filters. In
// most cases the search latency will be lower than unfiltered searches
// See https://docs.pinecone.io/docs/metadata-filtering
func WithFilters(filters any) Option {
	return func(o *Options) {
		o.Filters = filters
	}
}

// WithEmbedder returns an Option for setting the embedder that could be used when
// adding documents or doing similarity search (instead the embedder from the Store context)
// this is useful when we are using multiple LLMs with single vectorstore.
func WithEmbedder(embedder embeddings.Embedder) Option {
	return func(o *Options) {
		o.Embedder = embedder
	}
}

// WithDeduplicater returns an Option for setting the deduplicater that could be used
// when adding documents. This is useful to prevent wasting time on creating an embedding
// when one already exists.
func WithDeduplicater(fn func(ctx context.Context, doc schema.Document) bool) Option {
	return func(o *Options) {
		o.Deduplicater = fn
	}
}

// WithIDGenerater returns an Option for setting to generate the IDS
func WithIDGenerater(fn func(ctx context.Context, doc schema.Document) string) Option {
	return func(o *Options) {
		o.DocumentIDGenerater = fn
	}
}

// generateDummyDoumentID generates a UUID
func (o Options) generateDummyDoumentID(_ context.Context) string {
	return uuid.NewString()
}

// GenerateDocumentID calls the provided ID generator or creates a new UUID if not provided or the generated ID is not unique
func (o Options) GenerateDocumentID(ctx context.Context, doc schema.Document, ids []string) string {
	if o.DocumentIDGenerater == nil {
		return o.generateDummyDoumentID(ctx)
	}
	id := o.DocumentIDGenerater(ctx, doc)
	if len(id) < 1 {
		slog.Warn("Document ID generator did not generate an id", "id", id)
		return o.generateDummyDoumentID(ctx)
	}
	if slices.Contains(ids, id) {
		slog.Warn("Document ID generator generated a non unique id", "id", id)
		return o.generateDummyDoumentID(ctx)
	}
	return id
}
