package duckdb_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	duckdbstore "github.com/tmc/langchaingo/vectorstores/duckdb"
)

func preCheckEnvSetting(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

func makeNewCollectionName() string {
	return "test-collection-" + uuid.New().String()
}

func createOpenAIEmbedder(t *testing.T) *embeddings.EmbedderImpl {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	opts := []openai.Option{
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithHTTPClient(rr.Client()),
	}
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	return e
}

func createOpenAILLMAndEmbedder(t *testing.T) (llm *openai.LLM, e *embeddings.EmbedderImpl) {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	embeddingOpts := []openai.Option{
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithHTTPClient(rr.Client()),
	}
	if !rr.Recording() {
		embeddingOpts = append(embeddingOpts, openai.WithToken("test-api-key"))
	}

	embeddingLLM, err := openai.New(embeddingOpts...)
	require.NoError(t, err)

	e, err = embeddings.NewEmbedder(embeddingLLM)
	require.NoError(t, err)

	return llm, e
}

func newTestStore(ctx context.Context, t *testing.T, e embeddings.Embedder) duckdbstore.Store {
	t.Helper()

	store, err := duckdbstore.New(
		ctx,
		duckdbstore.WithConnectionURL(""),
		duckdbstore.WithEmbedder(e),
		duckdbstore.WithCollectionName(makeNewCollectionName()),
		duckdbstore.WithVectorDimensions(1536),
		duckdbstore.WithPreDeleteCollection(true),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	return store
}

func TestDuckDBStoreBasic(t *testing.T) {
	preCheckEnvSetting(t)

	e := createOpenAIEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}

func TestDuckDBStoreWithScoreThreshold(t *testing.T) {
	preCheckEnvSetting(t)

	e := createOpenAIEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0.6),
	)
	require.NoError(t, err)
	require.Greater(t, len(docs), 0)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0),
	)
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestDuckDBStoreSimilarityScore(t *testing.T) {
	preCheckEnvSetting(t)

	e := createOpenAIEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo is the capital city of Japan."},
		{PageContent: "Paris is the city of love."},
		{PageContent: "I like to visit London."},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(
		ctx,
		"What is the capital city of Japan?",
		3,
		vectorstores.WithScoreThreshold(0.85),
	)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.True(t, docs[0].Score > 0.9)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	preCheckEnvSetting(t)

	e := createOpenAIEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
	})
	require.NoError(t, err)

	_, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(-0.8),
	)
	require.Error(t, err)

	_, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(1.8),
	)
	require.Error(t, err)
}

func TestDuckDBAsRetriever(t *testing.T) {
	preCheckEnvSetting(t)

	llm, e := createOpenAILLMAndEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestDuckDBAsRetrieverWithScoreThreshold(t *testing.T) {
	preCheckEnvSetting(t)

	llm, e := createOpenAILLMAndEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.7)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestDuckDBAsRetrieverWithMetadataFilters(t *testing.T) {
	preCheckEnvSetting(t)

	llm, e := createOpenAILLMAndEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(
		ctx,
		[]schema.Document{
			{
				PageContent: "In office, the color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location":    "office",
					"square_feet": 100,
				},
			},
			{
				PageContent: "in sitting room, the color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location":    "sitting room",
					"square_feet": 400,
				},
			},
			{
				PageContent: "in patio, the color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location":    "patio",
					"square_feet": 800,
				},
			},
		},
	)
	require.NoError(t, err)

	filter := map[string]any{"location": "sitting room"}

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store,
				5,
				vectorstores.WithFilters(filter))),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	require.Contains(t, result, "purple", "expected purple in result")
	require.NotContains(t, result, "orange", "expected not orange in result")
	require.NotContains(t, result, "yellow", "expected not yellow in result")
}

func TestDeduplicater(t *testing.T) {
	preCheckEnvSetting(t)

	e := createOpenAIEmbedder(t)

	ctx := context.Background()
	store := newTestStore(ctx, t, e)

	_, err := store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"type": "city",
		}},
		{PageContent: "potato", Metadata: map[string]any{
			"type": "vegetable",
		}},
	}, vectorstores.WithDeduplicater(
		func(_ context.Context, doc schema.Document) bool {
			return doc.PageContent == "tokyo"
		},
	))
	require.NoError(t, err)

	docs, err := store.Search(ctx, 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "potato", docs[0].PageContent)
	require.Equal(t, "vegetable", docs[0].Metadata["type"])
}

func TestWithAllOptions(t *testing.T) {
	preCheckEnvSetting(t)

	e := createOpenAIEmbedder(t)

	ctx := context.Background()

	store, err := duckdbstore.New(
		ctx,
		duckdbstore.WithConnectionURL(""),
		duckdbstore.WithEmbedder(e),
		duckdbstore.WithPreDeleteCollection(true),
		duckdbstore.WithCollectionName(makeNewCollectionName()),
		duckdbstore.WithCollectionTableName("custom_collection"),
		duckdbstore.WithEmbeddingTableName("custom_embedding"),
		duckdbstore.WithCollectionMetadata(map[string]any{
			"key": "value",
		}),
		duckdbstore.WithVectorDimensions(1536),
		duckdbstore.WithDistanceMetric("cosine"),
	)
	require.NoError(t, err)
	defer store.Close()

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}
