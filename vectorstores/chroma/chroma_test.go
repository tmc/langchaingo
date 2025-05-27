package chroma_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	chromatypes "github.com/amikos-tech/chroma-go/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	tcchroma "github.com/testcontainers/testcontainers-go/modules/chroma"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

// TODO (noodnik2):
//  add relevant tests from "weaviate_test.go" (the initial tests are based upon those found in "pinecone_test.go")
//  consider refactoring out standard set of vectorstore unit tests to run across all implementations

//
// NOTE: View the 'getValues()' function to see which environment variables are required to run these tests.
// WARNING: When these values are not provided, the tests will not fail, but will be (silently) skipped.
//

func TestChromaGoStoreRest(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)
	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithDistanceFunction(chromatypes.COSINE),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)

		chromaGoStoreRest(s, t)
	}

}

func chromaGoStoreRest(s chroma.Store, t *testing.T) {
	_, err := s.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := s.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	country := docs[0].Metadata["country"]
	require.NoError(t, err)
	require.Equal(t, "japan", country)
}

func TestChromaStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)
	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithDistanceFunction(chromatypes.COSINE),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)

		chromaStoreRestWithScoreThreshold(s, t)
	}

}

func chromaStoreRestWithScoreThreshold(s chroma.Store, t *testing.T) {
	_, err := s.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London "},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := s.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.8))
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = s.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)
	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)

		similaritySearchWithInvalidScoreThreshold(s, t)
	}

}

func similaritySearchWithInvalidScoreThreshold(s chroma.Store, t *testing.T) {
	_, err := s.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London "},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	_, err = s.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(-0.8))
	require.Error(t, err)

	_, err = s.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(1.8))
	require.Error(t, err)
}

func TestChromaAsRetriever(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)

		chromaAsRetriever(s, llm, t)
	}

}

func chromaAsRetriever(s chroma.Store, llm *openai.LLM, t *testing.T) {
	_, err := s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 1),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestChromaAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithDistanceFunction(chromatypes.COSINE),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)

		chromaAsRetrieverWithScoreThreshold(s, llm, t)
	}

}

func chromaAsRetrieverWithScoreThreshold(s chroma.Store, llm *openai.LLM, t *testing.T) {
	_, err := s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5, vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	// TODO (noodnik2): clarify - WHY should we see "orange" in the result,
	//  as required by (expected in) the original "Pinecone" test??
	//   require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestChromaAsRetrieverWithMetadataFilterEqualsClause(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)

		chromaAsRetrieverWithMetadataFilterEqualsClause(s, llm, t)
	}
}

func chromaAsRetrieverWithMetadataFilterEqualsClause(s chroma.Store, llm *openai.LLM, t *testing.T) {
	_, err := s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{
				PageContent: "The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
	)
	require.NoError(t, err)

	filter := make(map[string]any)
	filterValue := make(map[string]any)
	filterValue["$eq"] = "patio"
	filter["location"] = filterValue

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5, vectorstores.WithFilters(filter)),
		),
		"What colors is the lamp?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestChromaAsRetrieverWithMetadataFilterInClause(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, newChromaErr := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, newChromaErr)

		defer cleanupTestArtifacts(t, s)

		chromaAsRetrieverWithMetadataFilterInClause(s, llm, t)
	}
}

func chromaAsRetrieverWithMetadataFilterInClause(s chroma.Store, llm *openai.LLM, t *testing.T) {
	ns := getTestNameSpace()
	_, addDocumentsErr := s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{
				PageContent: "The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
		vectorstores.WithNameSpace(ns),
	)
	require.NoError(t, addDocumentsErr)

	llm, newOpenaiErr := openai.New()
	require.NoError(t, newOpenaiErr)

	filter := make(map[string]any)
	filterValue := make(map[string]any)
	filterValue["$in"] = []string{"office", "kitchen"}
	filter["location"] = filterValue

	result, runChainErr := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5, vectorstores.WithNameSpace(ns),
				vectorstores.WithFilters(filter)),
		),
		"What color(s) was/were the lamp(s) beside the desk described as?",
	)
	require.NoError(t, runChainErr)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "orange", "expected orange in result")
}

func TestChromaAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)
		chromaAsRetrieverWithMetadataFilterNotSelected(s, llm, t)
	}
}

func chromaAsRetrieverWithMetadataFilterNotSelected(s chroma.Store, llm *openai.LLM, t *testing.T) {
	_, err := s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{
				PageContent: "The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5),
		),
		"What are all the colors of the lamps beside the desk?",
	)
	result = strings.ToLower(result)
	require.NoError(t, err)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "blue", "expected blue in result")
	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "purple", "expected purple in result")
	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestChromaAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	for _, v := range []string{chroma.ChromaV1, chroma.ChromaV2} {
		s, err := chroma.New(
			chroma.WithOpenAIAPIKey(openaiAPIKey),
			chroma.WithChromaURL(testChromaURL),
			chroma.WithNameSpace(getTestNameSpace()),
			chroma.WithEmbedder(e),
			chroma.WithChromaVersion(v),
		)
		require.NoError(t, err)

		defer cleanupTestArtifacts(t, s)
		chromaAsRetrieverWithMetadataFilters(s, llm, t)
	}

}

func chromaAsRetrieverWithMetadataFilters(s chroma.Store, llm *openai.LLM, t *testing.T) {
	_, err := s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{
				PageContent: "The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location":    "sitting room",
					"square_feet": 100,
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location":    "sitting room",
					"square_feet": 400,
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location":    "patio",
					"square_feet": 800,
				},
			},
		},
	)
	require.NoError(t, err)

	filter := map[string]interface{}{
		"$and": []map[string]interface{}{
			{
				"location": map[string]interface{}{
					"$eq": "sitting room",
				},
			},
			{
				"square_feet": map[string]interface{}{
					"$gte": 300,
				},
			},
		},
	}

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5, vectorstores.WithFilters(filter)),
		),
		"What color is the lamp beside the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "purple", "expected black in purple")
}

func getValues(t *testing.T) (string, string) {
	t.Helper()

	openaiAPIKey := os.Getenv(chroma.OpenAIAPIKeyEnvVarName)
	if openaiAPIKey == "" {
		t.Skipf("Must set %s to run test", chroma.OpenAIAPIKeyEnvVarName)
	}

	chromaURL := os.Getenv(chroma.ChromaURLKeyEnvVarName)
	if chromaURL == "" {
		chromaContainer, err := tcchroma.Run(context.Background(), "chromadb/chroma:0.4.24")
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, chromaContainer.Terminate(context.Background()))
		})

		chromaURL, err = chromaContainer.RESTEndpoint(context.Background())
		if err != nil {
			t.Skipf("Failed to get chroma container REST endpoint: %s", err)
		}
	}

	return chromaURL, openaiAPIKey
}

func cleanupTestArtifacts(t *testing.T, s chroma.Store) {
	t.Helper()
	require.NoError(t, s.RemoveCollection())
}

func getTestNameSpace() string {
	return fmt.Sprintf("test-namespace-%s", uuid.New().String())
}
