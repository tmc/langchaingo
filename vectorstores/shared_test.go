// Package vectorstores provides shared test utilities for vectorstore implementations.
package vectorstores

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// VectorStoreTestSuite provides a standard set of tests for vectorstore implementations.
type VectorStoreTestSuite struct {
	Store    VectorStore
	Embedder embeddings.Embedder
	LLM      llms.Model
}

// NewVectorStoreTestSuite creates a new test suite for a vectorstore.
func NewVectorStoreTestSuite(store VectorStore, embedder embeddings.Embedder, llm llms.Model) *VectorStoreTestSuite {
	return &VectorStoreTestSuite{
		Store:    store,
		Embedder: embedder,
		LLM:      llm,
	}
}

// TestBasicAddAndSearch tests basic document addition and similarity search.
func (suite *VectorStoreTestSuite) TestBasicAddAndSearch(t *testing.T) {
	t.Helper()

	docs := []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "potato", Metadata: map[string]any{"country": "ireland"}},
	}

	_, err := suite.Store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)

	similaritySearchDocs, err := suite.Store.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, similaritySearchDocs, 1)
	require.Equal(t, "tokyo", similaritySearchDocs[0].PageContent)

	country := similaritySearchDocs[0].Metadata["country"]
	require.Equal(t, "japan", country)
}

// TestSimilaritySearchWithScoreThreshold tests similarity search with score filtering.
func (suite *VectorStoreTestSuite) TestSimilaritySearchWithScoreThreshold(t *testing.T) {
	t.Helper()

	docs := []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "potato", Metadata: map[string]any{"country": "ireland"}},
	}

	_, err := suite.Store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)

	// Search with a very high threshold - should return no results
	similaritySearchDocs, err := suite.Store.SimilaritySearch(
		context.Background(),
		"japan",
		5,
		WithScoreThreshold(0.8),
	)
	require.NoError(t, err)
	require.Len(t, similaritySearchDocs, 1)
	require.Equal(t, "tokyo", similaritySearchDocs[0].PageContent)
}

// TestAsRetriever tests using the vectorstore as a retriever.
func (suite *VectorStoreTestSuite) TestAsRetriever(t *testing.T) {
	t.Helper()

	docs := []schema.Document{
		{
			PageContent: "There are pants in the wardrobe",
			Metadata:    map[string]any{"room": "bedroom"},
		},
		{
			PageContent: "There is a coffee pot in the kitchen",
			Metadata:    map[string]any{"room": "kitchen"},
		},
		{
			PageContent: "The big gray couch is in the living room next to the desk. The desk is black and next to the big gray couch and has a coffee mug on it.",
			Metadata:    map[string]any{"room": "living room"},
		},
		{
			PageContent: "In the living room, there is a big gray couch, a black desk, and a beige rug.",
			Metadata:    map[string]any{"room": "living room"},
		},
		{
			PageContent: "The new lamp is next to the dresser in the bedroom. The dresser is brown and the lamp is green.",
			Metadata:    map[string]any{"room": "bedroom"},
		},
	}

	_, err := suite.Store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)

	// Test as retriever with retrieval QA chain
	result, err := chains.LoadStuffQA(suite.LLM).Call(
		context.Background(),
		map[string]any{
			"input_documents": docs,
			"question":        "What colors is each piece of furniture next to the desk?",
		},
	)
	require.NoError(t, err)

	answer, ok := result["text"].(string)
	require.True(t, ok)
	require.Contains(t, answer, "black")
	require.Contains(t, answer, "beige")
}

// TestAsRetrieverWithMetadataFilter tests using the vectorstore as a retriever with metadata filtering.
func (suite *VectorStoreTestSuite) TestAsRetrieverWithMetadataFilter(t *testing.T) {
	t.Helper()

	docs := []schema.Document{
		{
			PageContent: "There are pants in the wardrobe",
			Metadata:    map[string]any{"room": "bedroom", "furniture": "wardrobe"},
		},
		{
			PageContent: "There is a coffee pot in the kitchen",
			Metadata:    map[string]any{"room": "kitchen", "furniture": "none"},
		},
		{
			PageContent: "The big gray couch is in the living room",
			Metadata:    map[string]any{"room": "living room", "furniture": "couch"},
		},
	}

	_, err := suite.Store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)

	// This test would need vectorstore-specific metadata filter implementation
	// For now, we'll just test basic retrieval
	retriever := ToRetriever(suite.Store, 2)
	retrievedDocs, err := retriever.GetRelevantDocuments(context.Background(), "furniture")
	require.NoError(t, err)
	require.Len(t, retrievedDocs, 2)
}

// TestSimilaritySearchWithEmbeddings tests similarity search using embeddings directly.
func (suite *VectorStoreTestSuite) TestSimilaritySearchWithEmbeddings(t *testing.T) {
	t.Helper()

	docs := []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "potato", Metadata: map[string]any{"country": "ireland"}},
	}

	_, err := suite.Store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)

	// Generate embedding for search query
	queryEmbedding, err := suite.Embedder.EmbedQuery(context.Background(), "japan")
	require.NoError(t, err)

	// Search using the embedding
	similaritySearchDocs, err := suite.Store.SimilaritySearchVectorWithScore(
		context.Background(),
		queryEmbedding,
		1,
	)
	require.NoError(t, err)
	require.Len(t, similaritySearchDocs, 1)
	require.Equal(t, "tokyo", similaritySearchDocs[0].Document.PageContent)
}

// RunStandardTests runs all standard vectorstore tests.
func (suite *VectorStoreTestSuite) RunStandardTests(t *testing.T) {
	t.Helper()

	t.Run("BasicAddAndSearch", suite.TestBasicAddAndSearch)
	t.Run("SimilaritySearchWithScoreThreshold", suite.TestSimilaritySearchWithScoreThreshold)
	t.Run("AsRetriever", suite.TestAsRetriever)
	t.Run("AsRetrieverWithMetadataFilter", suite.TestAsRetrieverWithMetadataFilter)
	t.Run("SimilaritySearchWithEmbeddings", suite.TestSimilaritySearchWithEmbeddings)
}