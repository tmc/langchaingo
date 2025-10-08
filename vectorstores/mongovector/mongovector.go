package mongovector

import (
	"context"
	"errors"
	"fmt"

	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/vectorstores"
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
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrUnsupportedOptions         = errors.New("unsupported options")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
)

// Store wraps a Mongo collection for writing to and searching an Atlas
// vector database.
type Store struct {
	coll          *mongo.Collection
	embedder      embeddings.Embedder
	index         string // Name of the Atlas Vector Search Index tied to Collection
	path          string // Field in Collection containing embedding vectors
	numCandidates int
}

var _ vectorstores.VectorStore = &Store{}

// New returns a Store that can read and write to the vector store.
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

func mergeAddOpts(store *Store, opts ...vectorstores.Option) (*vectorstores.Options, error) {
	mopts := &vectorstores.Options{}
	for _, set := range opts {
		set(mopts)
	}

	if mopts.ScoreThreshold != 0 || mopts.Filters != nil || mopts.NameSpace != "" || mopts.Deduplicater != nil {
		return nil, ErrUnsupportedOptions
	}

	if mopts.Embedder == nil {
		mopts.Embedder = store.embedder
	}

	if mopts.Embedder == nil {
		return nil, fmt.Errorf("embedder is unset")
	}

	return mopts, nil
}

// AddDocuments will create embeddings for the given documents using the
// user-specified embedding model, then insert that data into a vector store.
func (store *Store) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
	opts ...vectorstores.Option,
) ([]string, error) {
	cfg, err := mergeAddOpts(store, opts...)
	if err != nil {
		return nil, err
	}

	// Collect the page contents for embedding.
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

	bdocs := []bson.D{}
	for i := range vectors {
		bdocs = append(bdocs, bson.D{
			{Key: pageContentName, Value: docs[i].PageContent},
			{Key: store.path, Value: vectors[i]},
			{Key: metadataName, Value: docs[i].Metadata},
		})
	}

	res, err := store.coll.InsertMany(ctx, bdocs)
	if err != nil {
		return nil, err
	}

	// Since we don't allow user-defined ids, the InsertedIDs list will always
	// be primitive objects.
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

func mergeSearchOpts(store *Store, opts ...vectorstores.Option) (*vectorstores.Options, error) {
	mopts := &vectorstores.Options{}
	for _, set := range opts {
		set(mopts)
	}

	// Validate that the score threshold is in [0, 1]
	if mopts.ScoreThreshold > 1 || mopts.ScoreThreshold < 0 {
		return nil, ErrInvalidScoreThreshold
	}

	if mopts.Deduplicater != nil {
		return nil, ErrUnsupportedOptions
	}

	// Created an llm-specific-n-dimensional vector for searching the vector
	// space
	if mopts.Embedder == nil {
		mopts.Embedder = store.embedder
	}

	if mopts.Embedder == nil {
		return nil, fmt.Errorf("embedder is unset")
	}

	// If the user provides a method-level index, update.
	if mopts.NameSpace == "" {
		mopts.NameSpace = store.index
	}

	// If filters are unset, use an empty document.
	if mopts.Filters == nil {
		mopts.Filters = bson.D{}
	}

	return mopts, nil
}

// SimilaritySearch searches a vector store from the vector transformed from the
// query by the user-specified embedding model.
//
// This method searches the store-wrapped collection with an optionally
// provided index at instantiation, with a default index of "vector_index".
// Since multiple indexes can be defined for a collection, the options.NameSpace
// value can be used here to change the search index. The priority is
// options.NameSpace > Store.index > defaultIndex.
func (store *Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	opts ...vectorstores.Option,
) ([]schema.Document, error) {
	cfg, err := mergeSearchOpts(store, opts...)
	if err != nil {
		return nil, err
	}

	vector, err := cfg.Embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	numCandidates := defaultNumCandidatesScalar * numDocuments
	if store.numCandidates == 0 {
		numCandidates = numDocuments
	}

	// Create the pipeline for performing the similarity search.
	stage := struct {
		Index         string    `bson:"index"`         // Name of Atlas Vector Search Index tied to Collection
		Path          string    `bson:"path"`          // Field in Collection containing embedding vectors
		QueryVector   []float32 `bson:"queryVector"`   // Query as vector
		NumCandidates int       `bson:"numCandidates"` // Number of nearest neighbors to use during the search.
		Limit         int       `bson:"limit"`         // Number of docments to return
		Filter        any       `bson:"filter"`        // MQL matching expression
	}{
		Index:         cfg.NameSpace,
		Path:          store.path,
		QueryVector:   vector,
		NumCandidates: numCandidates,
		Limit:         numDocuments,
		Filter:        cfg.Filters,
	}

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$vectorSearch", Value: stage},
		},
		bson.D{
			{Key: "$set", Value: bson.D{{Key: scoreName, Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}}}},
		},
	}

	// Execute the aggregation.
	cur, err := store.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	found := []schema.Document{}
	for cur.Next(ctx) {
		doc := schema.Document{}
		err := cur.Decode(&doc)
		if err != nil {
			return nil, err
		}

		if doc.Score < cfg.ScoreThreshold {
			continue
		}

		found = append(found, doc)
	}

	return found, nil
}
