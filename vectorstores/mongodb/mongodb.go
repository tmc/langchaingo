package mongodb

import (
	"context"
	"errors"
	"fmt"

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

// Document comes from https://www.mongodb.com/developer/products/mongodb/doc-modeling-vector-search/
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

// TODO: https://github.com/langchain-ai/langchain/blob/892bd4c29be34c0cc095ed178be6d60c6858e2ec/libs/partners/mongodb/langchain_mongodb/vectorstores.py#L187
func (s *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]interface{}, error) {
	opts := s.getOptions(options...)

	// specify the collection to query from
	coll := s.client.Database(s.databaseName).Collection(s.collectionName)

	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	queryResult, err := coll.Aggregate() //TODO
	if err != nil {
		return nil, err
	}
	if queryResult == nil {
		return []interface{}{}, nil
	}

	var doc Document
	err = queryResult.Decode(&doc)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, error(fmt.Errorf("score threshold must be between 0 and 1"))
	}
	return opts.ScoreThreshold, nil
}
