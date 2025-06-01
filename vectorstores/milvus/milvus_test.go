package milvus

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/require"
	tcmilvus "github.com/testcontainers/testcontainers-go/modules/milvus"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func getEmbedding(model string, connectionStr ...string) (llms.Model, *embeddings.EmbedderImpl) {
	opts := []ollama.Option{ollama.WithModel(model)}
	if len(connectionStr) > 0 {
		opts = append(opts, ollama.WithServerURL(connectionStr[0]))
	}
	llm, err := ollama.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}
	return llms.Model(llm), e
}

func getNewStore(t *testing.T, opts ...Option) (Store, error) {
	t.Helper()
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		t.Skip("OLLAMA_HOST not set")
	}
	_, e := getEmbedding("gemma:2b")

	url := os.Getenv("MILVUS_URL")
	if url == "" {
		milvusContainer, err := tcmilvus.Run(t.Context(), "milvusdb/milvus:v2.4.0-rc.1-latest")
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			if err := milvusContainer.Terminate(context.Background()); err != nil {
				t.Logf("Failed to terminate milvus container: %v", err)
			}
		})

		url, err = milvusContainer.ConnectionString(t.Context())
		if err != nil {
			t.Skipf("Failed to get milvus container endpoint: %s", err)
		}
	}
	config := client.Config{
		Address: url,
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
		t.Context(),
		config,
		opts...,
	)
}

func TestMilvusConnection(t *testing.T) {
	t.Parallel()
	storer, err := getNewStore(t, WithDropOld(), WithCollectionName("test"))
	require.NoError(t, err)

	data := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"population": 9.7, "area": 622}},
		{PageContent: "Kyoto", Metadata: map[string]any{"population": 1.46, "area": 828}},
		{PageContent: "Hiroshima", Metadata: map[string]any{"population": 1.2, "area": 905}},
		{PageContent: "Kazuno", Metadata: map[string]any{"population": 0.04, "area": 707}},
		{PageContent: "Nagoya", Metadata: map[string]any{"population": 2.3, "area": 326}},
		{PageContent: "Toyota", Metadata: map[string]any{"population": 0.42, "area": 918}},
		{PageContent: "Fukuoka", Metadata: map[string]any{"population": 1.59, "area": 341}},
		{PageContent: "Paris", Metadata: map[string]any{"population": 11, "area": 105}},
		{PageContent: "London", Metadata: map[string]any{"population": 9.5, "area": 1572}},
		{PageContent: "Santiago", Metadata: map[string]any{"population": 6.9, "area": 641}},
		{PageContent: "Buenos Aires", Metadata: map[string]any{"population": 15.5, "area": 203}},
		{PageContent: "Rio de Janeiro", Metadata: map[string]any{"population": 13.7, "area": 1200}},
		{PageContent: "Sao Paulo", Metadata: map[string]any{"population": 22.6, "area": 1523}},
	}

	_, err = storer.AddDocuments(t.Context(), data)
	require.NoError(t, err)

	// search docs with filter
	filterRes, err := storer.SimilaritySearch(t.Context(),
		"Tokyo", 10,
		vectorstores.WithFilters("meta['area']==622"),
	)

	require.NoError(t, err)
	require.Len(t, filterRes, 1)

	japanRes, err := storer.SimilaritySearch(t.Context(),
		"Tokyo", 2,
		vectorstores.WithScoreThreshold(0.5))
	require.NoError(t, err)
	require.Len(t, japanRes, 1)
}
