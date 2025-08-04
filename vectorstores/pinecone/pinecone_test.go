package pinecone_test

import (
	"context"
	"os"
	"strings"
	"testing"
	
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/langchaingo/vectorstores/pinecone"
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

func TestPineconeStoreRest(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	storer, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(uuid.New().String()),
	)
	require.NoError(t, err)
	
	_, err = storer.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	})
	require.NoError(t, err)
	
	docs, err := storer.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
}

func TestPineconeStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	storer, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(uuid.New().String()),
	)
	require.NoError(t, err)
	
	_, err = storer.AddDocuments(context.Background(), []schema.Document{
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
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	storer, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
		pinecone.WithNameSpace(uuid.New().String()),
	)
	require.NoError(t, err)
	
	_, err = storer.AddDocuments(context.Background(), []schema.Document{
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

func TestPineconeAsRetriever(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	store, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)
	
	id := uuid.New().String()
	
	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(id),
	)
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

func TestPineconeAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	store, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)
	
	id := uuid.New().String()
	
	_, err = store.AddDocuments(
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
	
	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestPineconeAsRetrieverWithMetadataFilterEqualsClause(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	store, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)
	
	id := uuid.New().String()
	
	_, err = store.AddDocuments(
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

func TestPineconeAsRetrieverWithMetadataFilterInClause(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	store, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)
	
	id := uuid.New().String()
	
	_, err = store.AddDocuments(
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
	
	filter := make(map[string]any)
	filterValue := make(map[string]any)
	filterValue["$in"] = []string{"office", "kitchen"}
	filter["location"] = filterValue
	
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
	
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "orange", "expected orange in result")
}

func TestPineconeAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	store, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)
	
	id := uuid.New().String()
	
	_, err = store.AddDocuments(
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
	
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "blue", "expected blue in result")
	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "purple", "expected purple in result")
	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestPineconeAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()
	
	apiKey, host := getValues(t)
	
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	
	store, err := pinecone.New(
		pinecone.WithAPIKey(apiKey),
		pinecone.WithHost(host),
		pinecone.WithEmbedder(e),
	)
	require.NoError(t, err)
	
	id := uuid.New().String()
	
	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
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
		},
		vectorstores.WithNameSpace(id),
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				id), vectorstores.WithFilters(filter)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	
	require.Contains(t, result, "purple", "expected black in purple")
}
