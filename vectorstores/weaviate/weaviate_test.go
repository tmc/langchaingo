package weaviate

import (
	"context"
	"fmt"
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
	"github.com/weaviate/weaviate/entities/models"
)

func getValues(t *testing.T) (string, string) {
	t.Helper()

	scheme := os.Getenv("WEAVIATE_SCHEME")
	if scheme == "" {
		t.Skip("Must set WEAVIATE_SCHEME to run test")
	}

	host := os.Getenv("WEAVIATE_HOST")
	if host == "" {
		t.Skip("Must set WEAVIATE_HOST to run test")
	}

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	return scheme, host
}

func randomizedCamelCaseClass() string {
	return "Test" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

func createTestClass(ctx context.Context, s Store) error {
	properties := make([]*models.Property, 0, len(s.queryAttrs))
	for _, v := range s.queryAttrs {
		properties = append(properties, &models.Property{
			Name:        v,
			Description: fmt.Sprintf("test property %s", v),
			DataType:    []string{"text"},
		})
	}
	return s.client.Schema().ClassCreator().WithClass(&models.Class{
		Class:       s.indexName,
		Description: "test class",
		VectorIndexConfig: map[string]any{
			"distance": "cosine",
		},
		ModuleConfig: map[string]any{},
		Properties:   properties,
	}).Do(ctx)
}

func TestWeaviateStoreRest(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"country"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, docs[0].PageContent, "tokyo")
	require.Equal(t, docs[0].Metadata["country"], "japan")
}

func TestWeaviateStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	err = store.AddDocuments(context.Background(), []schema.Document{
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
	docs, err := store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.9))
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	err = store.AddDocuments(context.Background(), []schema.Document{
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

	_, err = store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(-0.8))
	require.Error(t, err)

	_, err = store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(1.8))
	require.Error(t, err)
}

func TestWeaviateAsRetriever(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	nameSpace := uuid.New().String()

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	llm, err := openai.New()
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1, vectorstores.WithNameSpace(nameSpace)),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestWeaviateAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := embeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	nameSpace := randomizedCamelCaseClass()
	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	llm, err := openai.New()
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				nameSpace), vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}
