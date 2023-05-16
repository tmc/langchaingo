package pinecone_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

func getValues(t *testing.T) (string, string, string, string) {
	t.Helper()

	pineconeApiKey := os.Getenv("PINECONE_API_KEY")
	if pineconeApiKey == "" {
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

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	return environment, pineconeApiKey, indexName, projectName
}

/* func TestPineconeStoreGRPC(t *testing.T) {
	t.Parallel()

	environment, apiKey, indexName, projectName := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	storer, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithEnvironment(environment),
		pinecone.WithIndexName(indexName),
		pinecone.WithProjectName(projectName),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(uuid.New().String()),
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
} */

func TestPineconeStoreRest(t *testing.T) {
	t.Parallel()

	environment, apiKey, indexName, projectName := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	storer, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithEnvironment(environment),
		pinecone.WithIndexName(indexName),
		pinecone.WithProjectName(projectName),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(uuid.New().String()),
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

func TestPineconeAsRetriever(t *testing.T) {
	t.Parallel()

	environment, apiKey, indexName, projectName := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithEnvironment(environment),
		pinecone.WithIndexName(indexName),
		pinecone.WithProjectName(projectName),
		pinecone.WithEmbedder(e),
	)

	id := uuid.New().String()

	store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(id),
	)

	llm, err := openai.New()
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1, vectorstores.WithNameSpace(id)),
		),
		"What color is the desk?",
	)

	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}
