package mongodb

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvalidOptions = errors.New("invalid options")

var serverAPI = options.ServerAPI(options.ServerAPIVersion1)

const (
	DefaultIndexName        = "default"
	DefaultTextKey          = "text"
	DefaultEmbeddingKey     = "embedding"
	DefaultRelevanceScoreFn = "cosine"
)

type Option func(p *Store)

func WithConnectionUri(connectionUri string) Option {
	return func(p *Store) {
		p.connectionUri = connectionUri
	}
}

func WithDatabaseName(database string) Option {
	return func(p *Store) {
		p.databaseName = database
	}
}

func WithCollectionName(collection string) Option {
	return func(p *Store) {
		p.collectionName = collection
	}
}

func WithIndexName(indexName string) Option {
	return func(p *Store) {
		p.indexName = indexName
	}
}

func WithTextKey(textKey string) Option {
	return func(p *Store) {
		p.textKey = textKey
	}
}

func WithRelevanceScoreFn(relevanceScoreFn string) Option {
	return func(p *Store) {
		p.relevanceScoreFn = relevanceScoreFn
	}
}

func WithEmbeddingKey(embeddingKey string) Option {
	return func(p *Store) {
		p.embeddingKey = embeddingKey
	}
}

func WithEmbedder(embedder embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = embedder
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{} //todo add default values
	for _, opt := range opts {
		opt(o)
	}
	if o.connectionUri == "" {
		return Store{}, fmt.Errorf("%w: missing mongodb connection string", ErrInvalidOptions)
	}
	if o.databaseName == "" {
		return Store{}, fmt.Errorf("%w: missing mongodb database", ErrInvalidOptions)
	}
	if o.collectionName == "" {
		return Store{}, fmt.Errorf("%w: missing mongodb collection name", ErrInvalidOptions)
	}
	if o.indexName == "" {
		o.indexName = DefaultIndexName
	}
	if o.relevanceScoreFn == "" {
		o.relevanceScoreFn = DefaultRelevanceScoreFn
	}
	if o.embeddingKey == "" {
		o.embeddingKey = DefaultEmbeddingKey
	}
	if o.textKey == "" {
		o.textKey = DefaultTextKey
	}
	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedding model", ErrInvalidOptions)
	}
	o.clientOptions = options.Client().ApplyURI(o.connectionUri).SetServerAPIOptions(serverAPI)
	return *o, nil
}
