package mariadb_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmariadb "github.com/testcontainers/testcontainers-go/modules/mariadb"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/llms/googleai"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/vectorstores"
	"github.com/vendasta/langchaingo/vectorstores/mariadb"
)

func preCheckEnvSetting(t *testing.T) string {
	t.Helper()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	mariadbURL := os.Getenv("MARIADB_CONNECTION_STRING")
	if mariadbURL == "" {
		mariadbContainer, err := tcmariadb.Run(
			context.Background(),
			"mariadb:11.7-rc", // supports vector types and functions
			tcmariadb.WithDatabase("db_test"),
			tcmariadb.WithUsername("user"),
			tcmariadb.WithPassword("passw0rd!"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("ready for connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, mariadbContainer.Terminate(context.Background()))
		})

		str, err := mariadbContainer.ConnectionString(context.Background())
		require.NoError(t, err)

		mariadbURL = str
	}

	return mariadbURL
}

func makeNewCollectionName() string {
	return fmt.Sprintf("test-collection-%s", uuid.New().String())
}

func cleanupTestArtifacts(ctx context.Context, t *testing.T, s mariadb.Store, mariadbURL string) {
	t.Helper()

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	require.NoError(t, s.RemoveDatabase(ctx, tx))

	require.NoError(t, tx.Commit())
}

func TestMariaDBStoreRest(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

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

func TestMariaDBStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
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

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0.8),
	)
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestMariaDBStoreSimilarityScore(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Tokyo is the capital city of Japan."},
		{PageContent: "Paris is the city of love."},
		{PageContent: "I like to visit London."},
	})
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := store.SimilaritySearch(
		ctx,
		"What is the capital city of Japan?",
		3,
		vectorstores.WithScoreThreshold(0.8),
	)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.True(t, docs[0].Score > 0.9)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
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

// note, we can also use same llm to show this test, but need imply
// openai embedding [dimensions](https://platform.openai.com/docs/api-reference/embeddings/create#embeddings-create-dimensions) args.
func TestSimilaritySearchWithDifferentDimensions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mariadbURL := preCheckEnvSetting(t)
	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("GENAI_API_KEY not set")
	}
	collectionName := makeNewCollectionName()

	// use Google embedding (now default model is embedding-001, with dimensions:768) to add some data to collection
	googleLLM, err := googleai.New(ctx, googleai.WithAPIKey(genaiKey))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(googleLLM)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(collectionName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Beijing"},
	})
	require.NoError(t, err)

	// use openai embedding (now default model is text-embedding-ada-002, with dimensions:1536) to add some data to same collection (same table)
	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err = embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err = mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(false),
		mariadb.WithDatabaseName(collectionName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
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
		5,
	)
	require.NoError(t, err)
	require.Len(t, docs, 5)
}

func TestMariaDBAsRetriever(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(
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

func TestMariaDBAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(
		context.Background(),
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
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestMariaDBAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{
				PageContent: "in kitchen, The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "in bedroom, The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "in office, The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "in sitting room, The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "in patio, The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	result = strings.ToLower(result)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "blue", "expected blue in result")
	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "purple", "expected purple in result")
	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestMariaDBAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(
		context.Background(),
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
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithVectorDimensions(1536),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
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
	t.Parallel()
	mariadbURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	require.NoError(t, err)
	db, err := sql.Open("mysql", mariadbURL)
	require.NoError(t, err)
	defer db.Close()

	store, err := mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
		mariadb.WithCollectionTableName("collection_table_name"),
		mariadb.WithEmbeddingTableName("embedding_table_name"),
		mariadb.WithDatabaseMetadata(map[string]any{
			"key": "value",
		}),
		mariadb.WithVectorDimensions(1536),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

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

	store, err = mariadb.New(
		ctx,
		mariadb.WithDB(db),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteDatabase(true),
		mariadb.WithDatabaseName(makeNewCollectionName()),
		mariadb.WithCollectionTableName("collection_table_name1"),
		mariadb.WithEmbeddingTableName("embedding_table_name1"),
		mariadb.WithDatabaseMetadata(map[string]any{
			"key": "value",
		}),
		mariadb.WithVectorDimensions(1536),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, mariadbURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err = store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}
