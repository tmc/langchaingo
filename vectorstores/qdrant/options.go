package qdrant

import (
	"errors"

	"github.com/tmc/langchaingo/embeddings"
	"google.golang.org/grpc"
)

const (
	DefaultCollectionName           = "langchain"
	DefaultPreDeleteCollection      = false
	DefaultEmbeddingStoreTableName  = "langchain_pg_embedding"
	DefaultCollectionStoreTableName = "langchain_pg_collection"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithGRPCConn is an option for setting the grpc connection. Use this or WithGRPCAddr.
func WithGRPCConn(conn *grpc.ClientConn) Option {
	return func(p *Store) {
		p.grpcConn = conn
	}
}

// WithGRPCAddr is an option for setting the grpc connection url.
func WithGRPCAddr(addr string) Option {
	return func(p *Store) {
		p.grpcAddr = addr
	}
}

// WithGRPCOptions is an option for setting the grpc connection options.
func WithGRPCOptions(opts ...grpc.DialOption) Option {
	return func(p *Store) {
		p.grpcOptions = opts
	}
}

// WithPreDeleteCollection is an option for setting if the collection should be deleted before creating.
// func WithPreDeleteCollection(preDelete bool) Option {
// 	return func(p *Store) {
// 		p.preDeleteCollection = preDelete
// 	}
// }

// WithCollectionName is an option for specifying the collection name.
func WithCollectionName(name string) Option {
	return func(p *Store) {
		p.collectionName = name
	}
}
