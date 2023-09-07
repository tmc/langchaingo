package chromadb

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	openaiEmbeddings "github.com/tmc/langchaingo/embeddings/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func getValues(t *testing.T) (string, string) {
	t.Helper()

	scheme := os.Getenv("CHROMA_SCHEME")
	if scheme == "" {
		t.Skip("Must set CHROMA_SCHEME to run test")
	}

	host := os.Getenv("CHROMA_HOST")
	if host == "" {
		t.Skip("Must set CHROMA_HOST to run test")
	}

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	return scheme, host
}

func TestChromaStoreRest(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := openaiEmbeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithDistanceFunc(chroma.COSINE),
	)
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

	country, err := strconv.Unquote(fmt.Sprintf("%s", docs[0].Metadata["country"]))
	require.NoError(t, err)
	require.Equal(t, country, "japan")
}

func TestChromaStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	e, err := openaiEmbeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithDistanceFunc(chroma.COSINE),
	)
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
		vectorstores.WithScoreThreshold(0.6))
	require.NoError(t, err)
	require.Len(t, docs, 2)

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
	e, err := openaiEmbeddings.NewOpenAI()
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
	)
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
