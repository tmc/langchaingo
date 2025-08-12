package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/vectorstores"
	"github.com/vendasta/langchaingo/vectorstores/mongovector"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	const (
		openAIEmbeddingModel = "text-embedding-3-small"
		openAIEmbeddingDim   = 1536
		similarityAlgorithm  = "dotProduct"
		indexDP1536          = "vector_index_dotProduct_1536"
		databaseName         = "langchaingo-test"
		collectionName       = "vstore"
	)

	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatalf("OPENAI_API_KEY required for this tutorial")
	}

	// First create a client and ensure that a vector search index that supports
	// OpenAI's embedding model exists on the example collection.
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI required and must point to an MongoDB Atlas Database")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to connect to server: %w", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("error disconnecting the client: %v", err)
		}
	}()

	coll := client.Database(databaseName).Collection(collectionName)

	if ok, _ := searchIndexExists(context.Background(), coll, indexDP1536); !ok {
		fields := []vectorField{
			{
				Type:          "vector",
				Path:          "plot_embedding", // Default path
				NumDimensions: openAIEmbeddingDim,
				Similarity:    similarityAlgorithm,
			},
			{
				Type: "filter",
				Path: "metadata.area",
			},
			{
				Type: "filter",
				Path: "metadata.population",
			},
		}

		// Create the vectorstore collection
		err = client.Database(databaseName).CreateCollection(context.Background(), collectionName)
		if err != nil {
			log.Fatalf("failed to create vector store collection: %w", err)
		}

		_, err = createVectorSearchIndex(context.Background(), coll, indexDP1536, fields...)
		if err != nil {
			log.Fatalf("failed to create index: %v", err)
		}
	}

	// Create an embeddings client using the OpenAI API. Requires environment
	// variable OPENAI_API_KEY to be set.
	llm, err := openai.New(openai.WithEmbeddingModel(openAIEmbeddingModel))
	if err != nil {
		log.Fatalf("failed to create an embedings client: %v", err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal("failed to create an embedder: %v", err)
	}

	// A Store is a wrapper for mongo.Collection, since adding and searching
	// vectors is collection-specific.
	store := mongovector.New(coll, embedder, mongovector.WithIndex(indexDP1536))

	// Add documents to the MongoDB Atlas Database vector store.
	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population": 38,
				"area":       2190,
			},
		},
		{
			PageContent: "Paris",
			Metadata: map[string]any{
				"population": 11,
				"area":       105,
			},
		},
		{
			PageContent: "London",
			Metadata: map[string]any{
				"population": 9.5,
				"area":       1572,
			},
		},
		{
			PageContent: "Santiago",
			Metadata: map[string]any{
				"population": 6.9,
				"area":       641,
			},
		},
		{
			PageContent: "Buenos Aires",
			Metadata: map[string]any{
				"population": 15.5,
				"area":       203,
			},
		},
		{
			PageContent: "Rio de Janeiro",
			Metadata: map[string]any{
				"population": 13.7,
				"area":       1200,
			},
		},
		{
			PageContent: "Sao Paulo",
			Metadata: map[string]any{
				"population": 22.6,
				"area":       1523,
			},
		},
	})
	if err != nil {
		log.Fatal("error adding documents: %v", err)
	}

	// Search for similar documents.
	docs, err := store.SimilaritySearch(context.Background(), "japan", 1)
	fmt.Println(docs)

	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(context.Background(), "South American cities", 4,
		vectorstores.WithScoreThreshold(0.7))
	fmt.Println(docs)

	// Search for similar documents using score threshold and metadata filter.
	filter := map[string]interface{}{
		"$and": []map[string]interface{}{
			{
				"metadata.area": map[string]interface{}{
					"$gte": 100,
				},
			},
			{
				"metadata.population": map[string]interface{}{
					"$gte": 15,
				},
			},
		},
	}

	docs, err = store.SimilaritySearch(context.Background(), "South American cities", 2,
		vectorstores.WithScoreThreshold(0.40),
		vectorstores.WithFilters(filter))
	fmt.Println(docs)
}

// vectorField defines the fields of an index used for vector search.
type vectorField struct {
	Type          string `bson:"type,omitempty"`
	Path          string `bson:"path,omityempty"`
	NumDimensions int    `bson:"numDimensions,omitempty"`
	Similarity    string `bson:"similarity,omitempty"`
}

// createVectorSearchIndex will create a vector search index on the "db.vstore"
// collection named "vector_index" with the provided field. This function blocks
// until the index has been created.
func createVectorSearchIndex(
	ctx context.Context,
	coll *mongo.Collection,
	idxName string,
	fields ...vectorField,
) (string, error) {
	def := struct {
		Fields []vectorField `bson:"fields"`
	}{
		Fields: fields,
	}

	view := coll.SearchIndexes()

	siOpts := options.SearchIndexes().SetName(idxName).SetType("vectorSearch")
	searchName, err := view.CreateOne(ctx, mongo.SearchIndexModel{Definition: def, Options: siOpts})
	if err != nil {
		return "", fmt.Errorf("failed to create the search index: %w", err)
	}

	// Await the creation of the index.
	var doc bson.Raw
	for doc == nil {
		cursor, err := view.List(ctx, options.SearchIndexes().SetName(searchName))
		if err != nil {
			return "", fmt.Errorf("failed to list search indexes: %w", err)
		}

		if !cursor.Next(ctx) {
			break
		}

		name := cursor.Current.Lookup("name").StringValue()
		queryable := cursor.Current.Lookup("queryable").Boolean()
		if name == searchName && queryable {
			doc = cursor.Current
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	return searchName, nil
}

// Check if the search index exists.
func searchIndexExists(ctx context.Context, coll *mongo.Collection, idx string) (bool, error) {
	view := coll.SearchIndexes()

	siOpts := options.SearchIndexes().SetName(idx).SetType("vectorSearch")
	cursor, err := view.List(ctx, siOpts)
	if err != nil {
		return false, fmt.Errorf("failed to list search indexes: %w", err)
	}

	if cursor == nil {
		return false, nil
	}

	if cursor.Current == nil {
		if ok := cursor.Next(ctx); !ok {
			return false, nil
		}
	}

	name := cursor.Current.Lookup("name").StringValue()
	queryable := cursor.Current.Lookup("queryable").Boolean()

	return name == idx && queryable, nil
}
