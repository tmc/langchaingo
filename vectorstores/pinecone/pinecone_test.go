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

var (
	id = uuid.New().String()
)

func getValues(t *testing.T) (string, string) {
	t.Helper()

	pineconeAPIKey := os.Getenv("PINECONE_API_KEY")
	if pineconeAPIKey == "" {
		t.Skip("Must set PINECONE_API_KEY to run test")
	}

	pineconeHost := os.Getenv("PINECONE_HOST")
	if pineconeHost == "" {
		t.Skip("Must set PINECONE_HOST to run test")
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		t.Skip("Must set OPENAI_API_KEY to run test")
	}

	return pineconeAPIKey, pineconeHost
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

func TestPineconeStoreAddDocuments(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(id),
	)
	require.NoError(t, err)

	_, err = storer.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Artificial intelligence (AI), in its broadest sense, is intelligence exhibited by machines, particularly computer systems."},
		{PageContent: "It is a field of research in computer science that develops and studies methods and software which enable machines to perceive their environment and uses learning and intelligence to take actions that maximize their chances of achieving defined goals."},
		{PageContent: "Alan Turing was the first person to conduct substantial research in the field that he called machine intelligence."},
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
		{PageContent: "The color of the house is blue."},
		{PageContent: "The color of the car is red."},
		{PageContent: "The color of the desk is orange."},
		{PageContent: "The color of the lamp beside the desk is black."},
		{PageContent: "The color of the chair beside the desk is beige."},
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
		{
			PageContent: "The color of the lamp beside the desk is orange.",
			Metadata: map[string]any{
				"location":    "office",
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
	})
	require.NoError(t, err)
}

func TestPineconeAsRetriever(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Artificial intelligence (AI), in its broadest sense, is intelligence exhibited by machines, particularly computer systems."},
		{PageContent: "It is a field of research in computer science that develops and studies methods and software which enable machines to perceive their environment and uses learning and intelligence to take actions that maximize their chances of achieving defined goals."},
		{PageContent: "Alan Turing was the first person to conduct substantial research in the field that he called machine intelligence."},
	}, vectorstores.WithNameSpace(id))
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1, vectorstores.WithNameSpace(id)),
		),
		"who was the first person to conduct substantial research in AI?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Alan Turing"), "expected Alan Turing in result")
}

func TestPineconeStoreRestWithScoreThreshold(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(id),
	)
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan?", 10,
		vectorstores.WithScoreThreshold(0.8),
		vectorstores.WithNameSpace(id),
	)
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0),
		vectorstores.WithNameSpace(id),
	)
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(id),
	)
	require.NoError(t, err)

	_, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(-0.8),
		vectorstores.WithNameSpace(id),
	)
	require.Error(t, err)

	_, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(1.8),
		vectorstores.WithNameSpace(id),
	)
	require.Error(t, err)
}

func TestPineconeAsRetrieverWithScoreThreshold(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(id), vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestPineconeAsRetrieverWithMetadataFilterEqualsClause(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(id), vectorstores.WithFilters(filter)),
		),
		"What colors is the lamp?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestPineconeAsRetrieverWithMetadataFilterInClause(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	filter := make(map[string]any)
	filterValue := make(map[string]any)
	filterValue["$in"] = []string{"office", "kitchen"}
	filter["location"] = filterValue

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(id), vectorstores.WithFilters(filter)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "orange", "expected orange in result")
}

func TestPineconeAsRetrieverWithMetadataFilters(t *testing.T) {
	apiKey, pineconeHost := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := pinecone.New(
		context.Background(),
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)

	filter := map[string]interface{}{
		"$and": []map[string]interface{}{
			{
				"location": map[string]interface{}{
					"$in": []string{"office", "sitting room"},
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(id), vectorstores.WithFilters(filter)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "purple", "expected black in purple")
}
