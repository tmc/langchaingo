package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	// "encoding/json"
	// "errors"
	// "github.com/google/uuid"
	// "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "github.com/tmc/langchaingo/embeddings"
	// "github.com/tmc/langchaingo/schema"
	// "github.com/tmc/langchaingo/vectorstores"
	// "google.golang.org/protobuf/types/known/structpb"
)

// Store is a wrapper around the mongodb client.
// We want to wrap mongo's client options all in 1
type Store struct {
	client *mongo.Client
	connectionUri string
	clientOptions *options.ClientOptions
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




//mongodb+srv://langchaingotest:PVgzJSoESmj0Hzpv@langchaingotest.us7qrm4.mongodb.net/?retryWrites=true&w=majority&appName=langchaingotest

