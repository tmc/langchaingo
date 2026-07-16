// This example shows you how to use a MongoDB Atlas vector store with kronk as
// the embedding model server. It follows the same flow as the OpenAI-based
// mongovector-vectorstore-example, but replaces the OpenAI embedding client
// with the shared kronk abstraction, which runs a local EmbeddingGemma model.
//
// The first time you run this program the kronk abstraction will download and
// install the embedding model and llama.cpp libraries.
//
// You need a MongoDB Atlas cluster. Set the MONGODB_URI environment variable to
// your Atlas connection string before running:
//
//	export MONGODB_URI="mongodb+srv://..."
//
// Run the example like this from the root of the project:
//
//	$ make example-mongovector-kronk
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/mongovector"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	const (
		// EmbeddingGemma-300m outputs 768-dimensional vectors.
		embeddingDim        = 768
		similarityAlgorithm = "dotProduct"
		indexName           = "vector_index_dotProduct_768"
		databaseName        = "langchaingo-test"
		collectionName      = "vstore"
	)

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017/?directConnection=true"
	}

	// -----------------------------------------------------------------------
	// Initialize the kronk embedding client.
	//
	// The first run downloads llama.cpp libraries and the EmbeddingGemma model.
	// This may take several minutes.

	fmt.Println("Initializing kronk embedding model...")

	client, err := kronk.New(context.Background(), kronk.Config{
		ModelSource: "ggml-org/embeddinggemma-300m-qat-Q8_0",
		AutoTune:    true,
	})
	if err != nil {
		log.Fatalf("failed to create kronk client: %v", err)
	}

	defer func() {
		fmt.Println("\nUnloading Kronk")
		if err := client.Unload(context.Background()); err != nil {
			fmt.Printf("failed to unload model: %v", err)
		}
	}()

	// Wrap the kronk client in a langchaingo embedder.
	embedder, err := embeddings.NewEmbedder(client)
	if err != nil {
		log.Fatalf("failed to create embedder: %v", err)
	}

	// -----------------------------------------------------------------------
	// Connect to MongoDB Atlas and ensure a vector search index exists.

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("error disconnecting the client: %v", err)
		}
	}()

	coll := mongoClient.Database(databaseName).Collection(collectionName)

	if ok, _ := searchIndexExists(context.Background(), coll, indexName); !ok {
		fields := []vectorField{
			{
				Type:          "vector",
				Path:          "plot_embedding", // Default path
				NumDimensions: embeddingDim,
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

		// Create the vectorstore collection.
		err = mongoClient.Database(databaseName).CreateCollection(context.Background(), collectionName)
		if err != nil {
			log.Fatalf("failed to create vector store collection: %v", err)
		}

		_, err = createVectorSearchIndex(context.Background(), coll, indexName, fields...)
		if err != nil {
			log.Fatalf("failed to create index: %v", err)
		}
	}

	// -----------------------------------------------------------------------
	// Create the vector store backed by MongoDB Atlas.

	store := mongovector.New(coll, embedder, mongovector.WithIndex(indexName))

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
		log.Fatalf("error adding documents: %v", err)
	}

	// Search for similar documents.
	docs, err := store.SimilaritySearch(context.Background(), "japan", 1)
	if err != nil {
		log.Fatalf("error searching: %v", err)
	}
	fmt.Println(docs)

	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(context.Background(), "South American cities", 4,
		vectorstores.WithScoreThreshold(0.7))
	if err != nil {
		log.Fatalf("error searching: %v", err)
	}
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
	if err != nil {
		log.Fatalf("error searching: %v", err)
	}
	fmt.Println(docs)
}

// vectorField defines the fields of an index used for vector search.
type vectorField struct {
	Type          string `bson:"type,omitempty"`
	Path          string `bson:"path,omityempty"`
	NumDimensions int    `bson:"numDimensions,omitempty"`
	Similarity    string `bson:"similarity,omitempty"`
}

// createVectorSearchIndex will create a vector search index on the collection
// with the provided name and fields. This function blocks until the index has
// been created.
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

// searchIndexExists checks if the search index exists and is queryable.
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
