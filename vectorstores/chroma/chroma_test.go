package chroma_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	chromago "github.com/amikos-tech/chroma-go"
	openapi "github.com/amikos-tech/chroma-go/swagger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
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
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithDistanceFunction(chromago.COSINE),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
		// chroma.WithEmbedder(e),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := s.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, docs[0].PageContent, "tokyo")
}

func TestChromaStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)
	// e , err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithDistanceFunction(chromago.COSINE),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(context.Background(), []schema.Document{
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
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(context.Background(), []schema.Document{
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
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
	)
	require.NoError(t, err)

	llm, err := openai.New()
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
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithDistanceFunction(chromago.COSINE),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(
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

	llm, err := openai.New()
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
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(
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

	llm, err := openai.New()
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
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, newChromaErr := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, newChromaErr)

	defer cleanupTestArtifacts(t, s)

	addDocumentsErr := s.AddDocuments(
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
	require.NoError(t, addDocumentsErr)

	llm, newOpenaiErr := openai.New()
	require.NoError(t, newOpenaiErr)

	filter := make(map[string]any)
	filterValue := make(map[string]any)
	filterValue["$in"] = []string{"office", "kitchen"}
	filter["location"] = filterValue

	_, runChainErr := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5, vectorstores.WithFilters(filter)),
		),
		"What color is the lamp in each room?",
	)
	require.Error(t, runChainErr)
	require.Equal(t, "500 Internal Server Error", runChainErr.Error())
	var apiError *openapi.GenericOpenAPIError
	require.True(t, errors.As(runChainErr, &apiError))
	message := string(apiError.Body())
	// Chroma doesn't (yet) support the `$in` operator
	require.Contains(t, message, "Expected where operator to be one of $gt, $gte, $lt, $lte, $ne, $eq, got $in")
}

func TestChromaAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(
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

	llm, err := openai.New()
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(s, 5),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	// TODO (noodnik2): hmm..., for this expected result (e.g., see analogous test for "Pinecone"),
	//  (how) does it connect a "location" to a "room"?
	//   require.Contains(t, result, "black", "expected black in result")
	//   require.Contains(t, result, "blue", "expected blue in result")
	//   require.Contains(t, result, "orange", "expected orange in result")
	//   require.Contains(t, result, "purple", "expected purple in result")
	//   require.Contains(t, result, "yellow", "expected yellow in result")
	//  Rather, I observe something like "I don't have enough information to answer the question."
	//  which seems correct to me; otherwise, perhaps I am missing something?
	require.Equal(t, 1, len(strings.Split(result, "\n")))
	require.Contains(t, result, "don't")
	for _, color := range []string{"black", "blue", "orange", "purple", "yellow"} {
		require.NotContains(t, result, color)
	}
}

func TestChromaAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()

	testChromaURL, openaiAPIKey := getValues(t)
	// e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	// require.NoError(t, err)

	s, err := chroma.New(
		chroma.WithOpenAiAPIKey(openaiAPIKey),
		chroma.WithChromaURL(testChromaURL),
		chroma.WithNameSpace(getTestNameSpace()),
		chroma.WithCollectionName(getTestCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(t, s)

	err = s.AddDocuments(
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

	llm, err := openai.New()
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
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	// TODO (noodnik2): hmm..., once received error: '"I don't know, as the context only
	//  mentions the color of the lamp beside the desk and doesn't provide information
	//  about the other rooms or their lamps." does not contain "purple"'
	require.Contains(t, result, "purple", "expected black in purple")
}

func getValues(t *testing.T) (string, string) {
	t.Helper()

	chromaURL := os.Getenv("CHROMA_URL")
	if chromaURL == "" {
		t.Skip("Must set CHROMA_URL to run test")
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		t.Skip("Must set OPENAI_API_KEY to run test")
	}

	return chromaURL, openaiAPIKey
}

func cleanupTestArtifacts(t *testing.T, s chroma.Store) {
	t.Helper()
	require.NoError(t, s.RemoveCollection())
}

func getTestCollectionName() string {
	return fmt.Sprintf("test-collection-%s", uuid.New().String())
}

func getTestNameSpace() string {
	return fmt.Sprintf("test-namespace-%s", uuid.New().String())
}
