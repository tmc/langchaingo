// Package mongovector implements a vector store using MongoDB as the backend.
//
// Notice: this packge uses the 2.0.0-beta release of the MongoDB Go Driver.
//
// The mongovector package provides a way to store and retrieve document embeddings
// using MongoDB's vector search capabilities. It implements the VectorStore
// interface from the vectorstores package, allowing it to be used interchangeably
// with other vector store implementations.
//
// Key features:
//   - Store document embeddings in MongoDB
//   - Perform similarity searches on stored embeddings
//   - Configurable index and path settings
//   - Support for custom embedding functions
//
// Main types:
//   - Store: The main type that implements the VectorStore interface
//   - Option: A function type for configuring the Store
//
// Usage:
//
//	import (
//	    "github.com/tmc/langchaingo/vectorstores/mongovector"
//	    "go.mongodb.org/mongo-driver/mongo"
//	)
//
//	// Create a new Store
//	coll := // ... obtain a *mongo.Collection
//	embedder := // ... obtain an embeddings.Embedder
//	store := mongovector.New(coll, embedder)
//
//	// Add documents
//	docs := []schema.Document{
//	    {PageContent: "Document 1"},
//	    {PageContent: "Document 2"},
//	}
//	ids, err := store.AddDocuments(context.Background(), docs)
//
//	// Perform similarity search
//	results, err := store.SimilaritySearch(context.Background(), "query", 5)
//
// The package also provides options for customizing the Store:
//   - WithIndex: Set a custom index name
//   - WithPath: Set a custom path for the vector field
//   - WithNumCandidates: Set the number of candidates for similarity search
//
// For more detailed information, see the documentation for individual types and functions.
package mongovector
