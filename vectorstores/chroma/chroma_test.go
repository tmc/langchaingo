package chroma_test

import (
	"context"
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

// TODO (noodnik2): update documentation to reflect the need to have a running Chroma server before running these tests
const chromaTestUrl = "http://localhost:8000"
const chromaTestCollection = "test-collection"

// TODO (noodnik2): add relevant tests from "weaviate_test.go" (these are based upon those found in "pinecone_test.go")
// TODO (noodnik2): consider refactoring out standard set of vectorstore unit tests to run across all implementations

func TestChromaGoStoreRest(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	storer, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithDistanceFunction(chromago.COSINE),
		chroma.WithResetChroma(true),
		//chroma.WithEmbedder(e),
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

func TestChromaStoreRestWithScoreThreshold(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	storer, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithDistanceFunction(chromago.COSINE),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	err = storer.AddDocuments(context.Background(), []schema.Document{
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
	docs, err := storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.8))
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	storer, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	err = storer.AddDocuments(context.Background(), []schema.Document{
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

	_, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(-0.8))
	require.Error(t, err)

	_, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(1.8))
	require.Error(t, err)
}

func TestChromaAsRetriever(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	store, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	id := uuid.New().String()

	err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(id),
	)
	require.NoError(t, err)

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
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestChromaAsRetrieverWithScoreThreshold(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	store, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	id := uuid.New().String()

	err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
		vectorstores.WithNameSpace(id),
	)
	require.NoError(t, err)

	llm, err := openai.New()
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				id), vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	//require.Contains(t, result, "orange", "expected orange in result") // TODO (noodnik2): clarify - WHY WOULD THIS BE EXPECTED?
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestChromaAsRetrieverWithMetadataFilterEqualsClause(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	store, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	id := uuid.New().String()

	err = store.AddDocuments(
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
		vectorstores.WithNameSpace(id),
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				id), vectorstores.WithFilters(filter)),
		),
		"What colors is the lamp?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestChromaAsRetrieverWithMetadataFilterInClause(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	store, newChromaErr := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, newChromaErr)

	id := uuid.New().String()

	addDocumentsErr := store.AddDocuments(
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
		vectorstores.WithNameSpace(id),
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				id), vectorstores.WithFilters(filter)),
		),
		"What color is the lamp in each room?",
	)
	require.Error(t, runChainErr)
	require.Equal(t, "500 Internal Server Error", runChainErr.Error())
	message := string(runChainErr.(*openapi.GenericOpenAPIError).Body())
	// Chroma doesn't (yet) support the `$in` operator
	require.Contains(t, message, "Expected where operator to be one of $gt, $gte, $lt, $lte, $ne, $eq, got $in")

	fmt.Printf("\n")
}

func TestChromaAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	store, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	id := uuid.New().String()

	err = store.AddDocuments(
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
		vectorstores.WithNameSpace(id),
	)
	require.NoError(t, err)

	llm, err := openai.New()
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				id)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	// TODO (noodnik2): hmm..., for this expected result (e.g., see analogous test for "Pinecone"),
	//  (how) does it connect a "location" to a "room"?
	//require.Contains(t, result, "black", "expected black in result")
	//require.Contains(t, result, "blue", "expected blue in result")
	//require.Contains(t, result, "orange", "expected orange in result")
	//require.Contains(t, result, "purple", "expected purple in result")
	//require.Contains(t, result, "yellow", "expected yellow in result")

	// Rather, I receive something like "I don't have enough information to answer the question."
	// which seems correct to me.  What am I missing?
	require.Equal(t, 1, len(strings.Split(result, "\n")))
	require.Contains(t, result, "don't")
	for _, color := range []string{"black", "blue", "orange", "purple", "yellow"} {
		require.NotContains(t, result, color)
	}
}

func TestChromaAsRetrieverWithMetadataFilters(t *testing.T) {
	//t.Parallel() // TODO (noodnik2): restore ability to run in parallel (e.g., using random collection names), removing "WithResetChroma"

	openaiApiKey := getValues(t)
	//e, err := openaiEmbeddings.NewOpenAI() // TODO (noodnik2): add ability to inject this
	//require.NoError(t, err)

	store, err := chroma.New(
		context.Background(),
		chroma.WithOpenAiApiKey(openaiApiKey),
		chroma.WithChromaUrl(chromaTestUrl),
		chroma.WithCollectionName(chromaTestCollection),
		chroma.WithResetChroma(true),
		//pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	id := uuid.New().String()

	err = store.AddDocuments(
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
		vectorstores.WithNameSpace(id),
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				id), vectorstores.WithFilters(filter)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	// TODO (noodnik2): hmm..., once received error: '"I don't know, as the context only mentions the color of the lamp beside
	//  the desk and doesn't provide information about the other rooms or their lamps." does not contain "purple"'
	require.Contains(t, result, "purple", "expected black in purple")
}

func getValues(t *testing.T) string {
	t.Helper()
	return os.Getenv("OPENAI_API_KEY")
}
