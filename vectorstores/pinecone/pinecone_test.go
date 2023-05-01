package pinecone

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
)

func getValues(t *testing.T) (string, string, string, string) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		t.Skip("Must set PINECONE_API_KEY to run test")
	}

	environment := os.Getenv("PINECONE_ENVIRONMENT")
	if environment == "" {
		t.Skip("Must set PINECONE_ENVIRONMENT to run test")
	}

	indexName := os.Getenv("PINECONE_INDEX")
	if environment == "" {
		t.Skip("Must set PINECONE_INDEX to run test")
	}

	projectName := os.Getenv("PINECONE_PROJECT")
	if environment == "" {
		t.Skip("Must set PINECONE_INDEX to run test")
	}

	return environment, apiKey, indexName, projectName
}

func TestPineconeStoreGRPC(t *testing.T) {
	t.Parallel()

	environment, apiKey, indexName, projectName := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	storer, err := New(
		WithApiKey(apiKey),
		WithEnvironment(environment),
		WithIndexName(indexName),
		WithProjectName(projectName),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		withGrpc(),
	)
	require.NoError(t, err)

	err = storer.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "yes"},
		{PageContent: "no"},
	})
	require.NoError(t, err)

	docs, err := storer.SimilaritySearch(context.Background(), "yeah", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, docs[0].PageContent, "yes")
}

func TestPineconeStoreRest(t *testing.T) {
	t.Parallel()

	environment, apiKey, indexName, projectName := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	storer, err := New(
		WithApiKey(apiKey),
		WithEnvironment(environment),
		WithIndexName(indexName),
		WithProjectName(projectName),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
	)
	require.NoError(t, err)

	err = storer.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := storer.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, docs[0].PageContent, "tokyo")
}
