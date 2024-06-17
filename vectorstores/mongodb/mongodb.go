package mongodb

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "github.com/tmc/langchaingo/embeddings"
	// "github.com/tmc/langchaingo/schema"
	// "github.com/tmc/langchaingo/vectorstores"
	// "google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
)

type Document struct {
	Text      string                 `bson:"text"`
	Embedding []float32              `bson:"embedding"`
	Metadata  map[string]interface{} `bson:",inline,omitempty"`
}

// Store is a wrapper around the mongodb client.
type Store struct {
	client           *mongo.Client
	connectionUri    string
	databaseName     string
	collectionName   string
	indexName        string
	textKey          string
	relevanceScoreFn string
	embeddingKey     string
	embedder         embeddings.Embedder
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

// AddDocuments creates vector embeddings from documents and adds them to the specified collection.
// We return the ids of the added documents.
func (s *Store) AddDocuments(ctx context.Context, docs []Document) (interface{}, error) {
	// specify the collection to insert into
	coll := s.client.Database(s.databaseName).Collection(s.collectionName)

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.Text)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}

	toInsert := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		doc.Embedding = vectors[i]
		toInsert = append(toInsert, doc)
	}

	result, err := coll.InsertMany(ctx, toInsert)
	if err != nil {
		return nil, err
	}

	return result.InsertedIDs, nil
}

func (s *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]interface{}, error) {
	return nil, nil
}

//mongodb+srv://langchaingotest:PVgzJSoESmj0Hzpv@langchaingotest.us7qrm4.mongodb.net/?retryWrites=true&w=majority&appName=langchaingotest
