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
	defaultIndex           = "vector_index"
	defaultPageContentName = "page_content"
	defaultPath            = "plot_embedding"
	metadataName           = "metadata"
	scoreName              = "score"
)

var (
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrUnsupportedOptions         = errors.New("unsupported options")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
)

type Store struct {
	coll            mongo.Collection
	embedder        embeddings.Embedder
	index           string
	pageContentName string
	path            string
}

var _ vectorstores.VectorStore = &Store{}

func New(coll mongo.Collection, embedder embeddings.Embedder, opts ...Option) Store {
	store := Store{
		coll:            coll,
		embedder:        embedder,
		index:           defaultIndex,
		pageContentName: defaultPageContentName,
		path:            defaultPath,
	}

	for _, opt := range opts {
		opt(&store)
	}

	return store
}

func (store *Store) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
	setters ...vectorstores.Option,
) ([]string, error) {
	opts := vectorstores.Options{}
	for _, set := range setters {
		set(&opts)
	}

	if opts.ScoreThreshold != 0 || opts.Filters != nil || opts.NameSpace != "" {
		return nil, ErrUnsupportedOptions
	}

	embedder := store.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	if embedder == nil {
		return nil, fmt.Errorf("embedder is unset")
	}

	// Collect the page contents for embedding.
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}

	//bdocs := make([]bson.D, 0, len(vectors))
	bdocs := []bson.D{}
	for i := 0; i < len(vectors); i++ {
		bdocs = append(bdocs, bson.D{
			{Key: store.pageContentName, Value: docs[i].PageContent},
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
		ids = append(ids, id.(bson.ObjectID).String())
	}

	return ids, nil
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
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
	setters ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := vectorstores.Options{}
	for _, set := range setters {
		set(&opts)
	}

	// Validate that the score threshold is in [0, 1]
	if opts.ScoreThreshold > 1 || opts.ScoreThreshold < 0 {
		return nil, ErrInvalidScoreThreshold
	}

	// Created an llm-specific-n-dimensional vector for searching the vector
	// space
	embedder := store.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	// If the user provides a method-level index, update.
	index := store.index
	if opts.NameSpace != "" {
		index = opts.NameSpace
	}

	vector, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	// Create the pipeline for performing the similarity search.
	stage := struct {
		Index         string    `bson:"index"`         // Name of Atlas Vector Search Index tied to Collection
		Path          string    `bson:"path"`          // Field in Collection containing embedding vectors
		QueryVector   []float32 `bson:"queryVector"`   // Query as vector
		NumCandidates int       `bson:"numCandidates"` // Number of nearest neighbors to use during the search.
		Limit         int       `bson:"limit"`         // Number of docments to return
	}{
		Index:         index,
		Path:          store.path,
		QueryVector:   vector,
		NumCandidates: 150,
		Limit:         numDocuments,
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
		doc := bson.M{}
		err := cur.Decode(&doc)
		if err != nil {
			return nil, err
		}

		schemaDoc := schema.Document{}

		score, ok := doc[scoreName].(float64)
		if ok {
			if score < float64(opts.ScoreThreshold) {
				continue
			}

			schemaDoc.Score = float32(score) // score ∈ [0, 1]
		}

		pageContent, ok := doc[store.pageContentName].(string)
		if ok {
			schemaDoc.PageContent = pageContent
		}

		metadata, ok := doc[metadataName].(map[string]any)
		if ok {
			schemaDoc.Metadata = metadata
		}

		found = append(found, schemaDoc)
	}

	return found, nil
}
