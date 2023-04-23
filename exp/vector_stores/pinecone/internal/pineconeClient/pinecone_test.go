package pineconeClient_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	t.Parallel()

	environment, apiKey := getEnvironmentAndApiKey(t)

	c, err := pineconeClient.New(
		pineconeClient.WithApiKey(apiKey),
		pineconeClient.WithEnvironment(environment),
		pineconeClient.WithIndexName("foo"),
		pineconeClient.WithDimensions(2),
	)
	require.NoError(t, err)

	queryVector := []float64{0.0, 0.0}
	expectedClosest := []float64{0.4, 0.5}
	expectedFarthest := []float64{10.1, 12.4}

	err = c.Upsert(
		context.Background(),
		pineconeClient.NewVectorsFromValues([][]float64{
			expectedClosest,
			expectedFarthest,
		}),
		"namespace",
	)
	require.NoError(t, err)

	queryResult, err := c.Query(
		context.Background(),
		queryVector,
		1,
		"namespace",
	)
	require.NoError(t, err)
	require.Len(t, queryResult.Matches, 1)

	closest := queryResult.Matches[0].Values
	assert.Equal(t, expectedClosest, closest)
}
