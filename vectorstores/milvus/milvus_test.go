package milvus

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmilvus "github.com/testcontainers/testcontainers-go/modules/milvus"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func getEmbedder(t *testing.T) (embeddings.Embedder, error) {
	t.Helper()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	url := os.Getenv("OPENAI_BASE_URL")
	opts := []openai.Option{}
	if url != "" {
		opts = append(opts, openai.WithBaseURL(url))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)
	return embeddings.NewEmbedder(llm)
}

func getNewStore(t *testing.T, opts ...Option) (Store, error) {
	t.Helper()
	url := os.Getenv("MILVUS_URL")
	if url == "" {
		milvusContainer, err := tcmilvus.RunContainer(context.Background(), testcontainers.WithImage("milvusdb/milvus:v2.3.9"))
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, milvusContainer.Terminate(context.Background()))
		})

		url, err = milvusContainer.ConnectionString(context.Background())
		if err != nil {
			t.Skipf("Failed to get milvus container endpoint: %s", err)
		}
	}
	config := client.Config{
		Address: url,
	}
	e, err := getEmbedder(t)
	if err != nil {
		return Store{}, err
	}
	idx, err := entity.NewIndexAUTOINDEX(entity.L2)
	if err != nil {
		return Store{}, err
	}
	opts = append(
		opts,
		WithEmbedder(e),
		WithIndex(idx))
	return New(
		context.Background(),
		config,
		opts...,
	)
}

func TestMilvusConnection(t *testing.T) {
	t.Parallel()
	storer, err := getNewStore(t, WithDropOld())
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
	japanRes, err := storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.3))
	require.NoError(t, err)
	require.Len(t, japanRes, 6)

	// test with a score threshold of 0, expected all 10 documents
	euRes, err := storer.SimilaritySearch(context.Background(),
		"Which of these are cities are located in Europe?", 10,
		vectorstores.WithScoreThreshold(1),
	)
	require.NoError(t, err)
	require.Len(t, euRes, 10)
}
