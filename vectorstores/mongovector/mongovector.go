package mongovector

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	defaultIndex               = "vector_index"
	pageContentName            = "pageContent"
	defaultPath                = "plot_embedding"
	metadataName               = "metadata"
	scoreName                  = "score"
	defaultNumCandidatesScalar = 10
)

var (
	// ErrEmbedderWrongNumberVectors is returned when the number of vectors from the embedder does not match the number of documents.
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	// ErrUnsupportedOptions is returned when unsupported options are provided.
	ErrUnsupportedOptions = errors.New("unsupported options")
	// ErrInvalidScoreThreshold is returned when an invalid score threshold is provided.
	ErrInvalidScoreThreshold = errors.New("score threshold must be between 0 and 1")
)

// Store represents a MongoDB-based vector store. It wraps a MongoDB collection
// for storing and searching vector embeddings.
type Store struct {
	coll          *mongo.Collection
	embedder      embeddings.Embedder
	index         string
	path          string
	numCandidates int
}

var _ vectorstores.VectorStore = &Store{}

// New creates a new MongoDB-based vector store with the given collection, embedder, and options.
// It initializes the store with default settings which can be overridden using the provided options.
func New(coll *mongo.Collection, embedder embeddings.Embedder, opts ...Option) Store {
	store := Store{
		coll:     coll,
		embedder: embedder,
		index:    defaultIndex,
		path:     defaultPath,
	}

	for _, opt := range opts {
		opt(&store)
	}

	return store
}

// AddDocuments creates embeddings for the given documents using the specified
// embedding model, then inserts the documents and their embeddings into the MongoDB collection.
// It returns the IDs of the added documents.
func (s *Store) AddDocuments(ctx context.Context, docs []schema.Document, opts ...vectorstores.Option) ([]string, error) {
	cfg, err := s.mergeAddOpts(opts...)
	if err != nil {
		return nil, err
	}

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := cfg.Embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}

	bdocs := make([]interface{}, len(docs))
	for i := range vectors {
		bdocs[i] = bson.D{
			{Key: pageContentName, Value: docs[i].PageContent},
			{Key: s.path, Value: vectors[i]},
			{Key: metadataName, Value: docs[i].Metadata},
		}
	}

	res, err := s.coll.InsertMany(ctx, bdocs)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(docs))
	for _, id := range res.InsertedIDs {
		id, ok := id.(fmt.Stringer)
		if !ok {
			return nil, fmt.Errorf("expected id for embedding to be a stringer")
		}

		ids = append(ids, id.String())
	}

	return ids, nil
}

// SimilaritySearch searches a vector store from the vector transformed from the
// query by the user-specified embedding model.
//
// This method searches the store-wrapped collection with an optionally
// provided index at instantiation, with a default index of "vector_index".
// Since multiple indexes can be defined for a collection, the options.NameSpace
// value can be used here to change the search index. The priority is
// options.NameSpace > Store.index > defaultIndex.
func (s *Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, opts ...vectorstores.Option) ([]schema.Document, error) {
	cfg, err := s.mergeSearchOpts(opts...)
	if err != nil {
		return nil, err
	}

	vector, err := cfg.Embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	numCandidates := defaultNumCandidatesScalar * numDocuments
	if s.numCandidates > 0 {
		numCandidates = s.numCandidates
	}

	stage := struct {
		Index         string    `bson:"index"`         // Name of Atlas Vector Search Index tied to Collection
		Path          string    `bson:"path"`          // Field in Collection containing embedding vectors
		QueryVector   []float32 `bson:"queryVector"`   // Query as vector
		NumCandidates int       `bson:"numCandidates"` // Number of nearest neighbors to use during the search.
		Limit         int       `bson:"limit"`         // Number of documents to return
		Filter        any       `bson:"filter"`        // MQL matching expression
	}{
		Index:         cfg.NameSpace,
		Path:          s.path,
		QueryVector:   vector,
		NumCandidates: numCandidates,
		Limit:         numDocuments,
		Filter:        cfg.Filters,
	}

	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: stage}},
		{{Key: "$set", Value: bson.D{{Key: scoreName, Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}}}}},
	}

	cur, err := s.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	found := []schema.Document{}
	for cur.Next(ctx) {
		var doc schema.Document
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}

		if doc.Score < cfg.ScoreThreshold {
			continue
		}

		found = append(found, doc)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return found, nil
}

func (s *Store) mergeAddOpts(opts ...vectorstores.Option) (*vectorstores.Options, error) {
	mopts := &vectorstores.Options{}
	for _, set := range opts {
		set(mopts)
	}

	if mopts.ScoreThreshold != 0 || mopts.Filters != nil || mopts.NameSpace != "" || mopts.Deduplicater != nil {
		return nil, ErrUnsupportedOptions
	}

	if mopts.Embedder == nil {
		mopts.Embedder = s.embedder
	}

	if mopts.Embedder == nil {
		return nil, fmt.Errorf("embedder is unset")
	}

	return mopts, nil
}

func (s *Store) mergeSearchOpts(opts ...vectorstores.Option) (*vectorstores.Options, error) {
	mopts := &vectorstores.Options{}
	for _, set := range opts {
		set(mopts)
	}

	if mopts.ScoreThreshold < 0 || mopts.ScoreThreshold > 1 {
		return nil, ErrInvalidScoreThreshold
	}

	if mopts.Deduplicater != nil {
		return nil, ErrUnsupportedOptions
	}

	if mopts.Embedder == nil {
		mopts.Embedder = s.embedder
	}

	if mopts.Embedder == nil {
		return nil, fmt.Errorf("embedder is unset")
	}

	if mopts.NameSpace == "" {
		mopts.NameSpace = s.index
	}

	if mopts.Filters == nil {
		mopts.Filters = bson.D{}
	}

	return mopts, nil
}
