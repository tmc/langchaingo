package milvus

import (
	"context"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/tei"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func getEmbedder() (embeddings.Embedder, error) {
	return tei.New(
		tei.WithAPIBaseURL("http://localhost:5500"),
	)
	// return hfe.NewHuggingface(
	// 	hfe.WithModel("thenlper/gte-large"),
	// )
}

func getNewStore(opts ...Option) (Store, error) {
	config := client.Config{
		Address: "http://localhost:19530",
	}
	e, err := getEmbedder()
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
	storer, err := getNewStore(WithDropOld())
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
		vectorstores.WithScoreThreshold(0.3))
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestMilvusSearch(t *testing.T) {
	t.Parallel()
	storer, err := getNewStore()
	require.NoError(t, err)

	require.NoError(t, err)
	// test with a score threshold of 0.8, expected 6 documents
	docs, err := storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.3),
	)
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = storer.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}
