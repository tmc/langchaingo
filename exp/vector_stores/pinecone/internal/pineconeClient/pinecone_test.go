package pineconeClient_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/exp/vector_stores/pinecone/internal/pineconeClient"
)

func getEnvironmentAndApiKey(t *testing.T) (string, string) {
	t.Parallel()

	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		t.Skip("Must set PINECONE_API_KEY to run test")
	}

	environment := os.Getenv("PINECONE_ENVIRONMENT")
	if environment == "" {
		t.Skip("Must set PINECONE_ENVIRONMENT to run test")
	}

	return environment, apiKey
}

func TestUpsertAndQuery(t *testing.T) {
	environment, apiKey := getEnvironmentAndApiKey(t)

	c, err := pineconeClient.New(
		pineconeClient.WithApiKey(apiKey),
		pineconeClient.WithEnvironment(environment),
		pineconeClient.WithIndexName("database"),
		pineconeClient.WithDimensions(1563),
	)
	require.NoError(t, err)

	vector := make([]float64, 0)
	for i := 0; i < 1536; i++ {
		vector = append(vector, 0.1)
	}

	err = c.Upsert(
		context.Background(),
		pineconeClient.NewVectorsFromValues([][]float64{
			vector,
		}),
		"namespace",
	)
	require.NoError(t, err)

	/* queryResult, err := c.Query(
		context.Background(),
		queryVector,
		1,
		"namespace",
	)
	require.NoError(t, err)
	require.Len(t, queryResult.Matches, 1)

	closest := queryResult.Matches[0].Values
	assert.Equal(t, expectedClosest, closest) */
}
