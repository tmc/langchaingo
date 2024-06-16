package mongodb

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"go.mongodb.org/mongo-driver/mongo"

	// "encoding/json"
	// "errors"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "github.com/tmc/langchaingo/embeddings"
	// "github.com/tmc/langchaingo/schema"
	// "github.com/tmc/langchaingo/vectorstores"
	// "google.golang.org/protobuf/types/known/structpb"
)

// Store is a wrapper around the mongodb client.
type Store struct {
	client           *mongo.Client
	connectionUri    string
	database         string
	collection       string
	indexName        string
	textKey          string
	relevanceScoreFn string
	embeddingKey     string
	embedding        embeddings.Embedder
	clientOptions    *options.ClientOptions
}

// New creates a new Store with options.
func New(ctx context.Context, opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}
	s.client, err = mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		return Store{}, err
	}
	return s, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []interface{}) error {
	return nil
}

func (s *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]interface{}, error) {
	return nil, nil
}

//mongodb+srv://langchaingotest:PVgzJSoESmj0Hzpv@langchaingotest.us7qrm4.mongodb.net/?retryWrites=true&w=majority&appName=langchaingotest
