package vearch_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/vearch"
)


func TestVearchStoreRest(t *testing.T) {
	t.Parallel()

	apiKey, host := getValues(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	store, err := vearch.New(
		vearch.WithDbName("langchaingo_dbt"),
		vearch.WithSpaceName("langchaingo_t"),
		vearch.WithURL("your router url"),
		vearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
}
