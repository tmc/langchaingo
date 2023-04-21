package pineconeClient_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/exp/vector_stores/pinecone/internal/pineconeClient"
)

func getEnvironmentAndApiKey(t *testing.T) (string, string) {
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
		pineconeClient.WithIndexName("foo"),
		pineconeClient.WithDimensions(2),
	)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

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
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	queryResult, err := c.Query(
		context.Background(),
		queryVector,
		1,
		"namespace",
	)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}

	if len(queryResult.Matches) != 1 {
		t.Errorf("Unexpected length of matches: %v", queryResult.Matches)
		return
	}

	closest := queryResult.Matches[0].Values
	if !reflect.DeepEqual(expectedClosest, closest) {
		t.Errorf("Expected closest not closest. Got %v, Want: %v", closest, expectedClosest)
		return
	}
}
