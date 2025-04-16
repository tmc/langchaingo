// Package mongovector implements a vector store using MongoDB as the backend.
//
// The mongovector package provide a way for users to read and write to a
// MongoDB Atlas Database as a vector store using the MongoDB Go Driver and a
// supported embedding service.
//
// Project goals:
//   - Allows users to embed data using various services, including OpenAI, Ollama, Mistral, and others.
//   - Implement the VectorStore interface, providing methods to add documents and perform similarity searches.
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
// Installation:
//
//	go get github.com/tmc/langchaingo/vectorstores/mongovector@v0.1.13-pre.0
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
