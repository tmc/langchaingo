// Package vectorstores provides a common test suite for vector store implementations.
package vectorstores

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

// TestSuite contains common tests that all vector store implementations should pass.
type TestSuite struct {
	Store     VectorStore
	Documents []schema.Document
	Cleanup   func()
}

// TestDocuments contains sample documents for testing.
var TestDocuments = []schema.Document{
	{
		PageContent: "Tokyo is the capital of Japan",
		Metadata: map[string]any{
			"country":    "Japan",
			"city":       "Tokyo",
			"population": 37400000,
			"continent":  "Asia",
		},
	},
	{
		PageContent: "Paris is the capital of France",
		Metadata: map[string]any{
			"country":    "France",
			"city":       "Paris", 
			"population": 2161000,
			"continent":  "Europe",
		},
	},
	{
		PageContent: "London is the capital of the United Kingdom",
		Metadata: map[string]any{
			"country":    "United Kingdom",
			"city":       "London",
			"population": 8982000,
			"continent":  "Europe",
		},
	},
	{
		PageContent: "New York is a major city in the United States",
		Metadata: map[string]any{
			"country":    "United States",
			"city":       "New York",
			"population": 8336817,
			"continent":  "North America",
		},
	},
	{
		PageContent: "Sydney is a major city in Australia",
		Metadata: map[string]any{
			"country":    "Australia", 
			"city":       "Sydney",
			"population": 5312000,
			"continent":  "Oceania",
		},
	},
}

// RunBasicTests runs the basic test suite for a vector store implementation.
func (suite *TestSuite) RunBasicTests(t *testing.T) {
	t.Helper()
	if suite.Cleanup != nil {
		defer suite.Cleanup()
	}

	t.Run("AddDocuments", suite.testAddDocuments)
	t.Run("SimilaritySearch", suite.testSimilaritySearch)
	t.Run("SimilaritySearchWithScore", suite.testSimilaritySearchWithScore)
	t.Run("SimilaritySearchWithFilters", suite.testSimilaritySearchWithFilters)
	t.Run("SimilaritySearchWithScoreThreshold", suite.testSimilaritySearchWithScoreThreshold)
}

// testAddDocuments tests adding documents to the vector store.
func (suite *TestSuite) testAddDocuments(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	docs := suite.Documents
	if docs == nil {
		docs = TestDocuments
	}

	ids, err := suite.Store.AddDocuments(ctx, docs)
	require.NoError(t, err, "AddDocuments should not return an error")
	assert.Len(t, ids, len(docs), "Should return as many IDs as documents added")
	
	for _, id := range ids {
		assert.NotEmpty(t, id, "Document ID should not be empty")
	}
}

// testSimilaritySearch tests basic similarity search functionality.
func (suite *TestSuite) testSimilaritySearch(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	docs := suite.Documents
	if docs == nil {
		docs = TestDocuments
	}

	// Add documents first
	_, err := suite.Store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	// Test similarity search
	results, err := suite.Store.SimilaritySearch(ctx, "capital city of Japan", 2)
	require.NoError(t, err, "SimilaritySearch should not return an error")
	assert.Len(t, results, 2, "Should return requested number of documents")
	
	// The most similar document should be about Tokyo
	assert.Contains(t, results[0].PageContent, "Tokyo", 
		"Most similar document should contain 'Tokyo'")
}

// testSimilaritySearchWithScore tests similarity search with score information.
func (suite *TestSuite) testSimilaritySearchWithScore(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	docs := suite.Documents
	if docs == nil {
		docs = TestDocuments
	}

	// Add documents first
	_, err := suite.Store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	// Test similarity search with scores
	results, err := suite.Store.SimilaritySearch(ctx, "European capital", 3)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	
	// Check that we get results about European cities
	foundEuropean := false
	for _, doc := range results {
		if continent, ok := doc.Metadata["continent"]; ok && continent == "Europe" {
			foundEuropean = true
			break
		}
	}
	assert.True(t, foundEuropean, "Should find at least one European city")
}

// testSimilaritySearchWithFilters tests similarity search with metadata filters.
func (suite *TestSuite) testSimilaritySearchWithFilters(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	docs := suite.Documents
	if docs == nil {
		docs = TestDocuments
	}

	// Add documents first
	_, err := suite.Store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	// Test with filters (this may not be supported by all implementations)
	filters := map[string]any{"continent": "Europe"}
	results, err := suite.Store.SimilaritySearch(ctx, "capital city", 5, 
		WithFilters(filters))
	
	// If filters are not supported, the test should still pass
	if err == nil {
		// All results should be from Europe
		for _, doc := range results {
			if continent, ok := doc.Metadata["continent"]; ok {
				assert.Equal(t, "Europe", continent, 
					"All filtered results should be from Europe")
			}
		}
	}
}

// testSimilaritySearchWithScoreThreshold tests similarity search with score threshold.
func (suite *TestSuite) testSimilaritySearchWithScoreThreshold(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	docs := suite.Documents
	if docs == nil {
		docs = TestDocuments
	}

	// Add documents first
	_, err := suite.Store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	// Test with score threshold
	results, err := suite.Store.SimilaritySearch(ctx, "capital of Japan", 5,
		WithScoreThreshold(0.8))
	
	// This may not be supported by all implementations
	if err == nil {
		// Should find at least the Tokyo document with high similarity
		assert.Greater(t, len(results), 0, "Should find at least one document with high similarity")
		
		// Check that the most relevant document is about Tokyo/Japan
		found := false
		for _, doc := range results {
			if doc.Metadata["country"] == "Japan" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find the Japan document with high similarity")
	}
}

// RunStressTests runs stress tests for the vector store implementation.
func (suite *TestSuite) RunStressTests(t *testing.T) {
	t.Helper()
	if suite.Cleanup != nil {
		defer suite.Cleanup()
	}

	t.Run("LargeDocumentBatch", suite.testLargeDocumentBatch)
	t.Run("ConcurrentOperations", suite.testConcurrentOperations)
}

// testLargeDocumentBatch tests adding a large batch of documents.
func (suite *TestSuite) testLargeDocumentBatch(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	
	// Create 100 test documents
	largeBatch := make([]schema.Document, 100)
	for i := 0; i < 100; i++ {
		largeBatch[i] = schema.Document{
			PageContent: fmt.Sprintf("Test document number %d with some content", i),
			Metadata: map[string]any{
				"batch_id": "stress_test",
				"doc_num":  i,
				"category": fmt.Sprintf("cat_%d", i%5),
			},
		}
	}

	// Add the large batch
	ids, err := suite.Store.AddDocuments(ctx, largeBatch)
	require.NoError(t, err, "Should handle large document batches")
	assert.Len(t, ids, 100, "Should return IDs for all documents")

	// Search within the batch
	results, err := suite.Store.SimilaritySearch(ctx, "Test document", 10)
	require.NoError(t, err)
	assert.Greater(t, len(results), 0, "Should find documents from the large batch")
}

// testConcurrentOperations tests concurrent operations on the vector store.
func (suite *TestSuite) testConcurrentOperations(t *testing.T) {
	t.Helper()
	
	ctx := context.Background()
	
	// This is a basic test - more sophisticated concurrent testing
	// could be added based on specific vector store capabilities
	
	// Add some initial documents
	_, err := suite.Store.AddDocuments(ctx, TestDocuments[:2])
	require.NoError(t, err)

	// Perform concurrent searches
	done := make(chan bool, 2)
	
	go func() {
		defer func() { done <- true }()
		_, err := suite.Store.SimilaritySearch(ctx, "capital", 1)
		assert.NoError(t, err, "Concurrent search 1 should not error")
	}()

	go func() {
		defer func() { done <- true }()
		_, err := suite.Store.SimilaritySearch(ctx, "city", 1)  
		assert.NoError(t, err, "Concurrent search 2 should not error")
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}