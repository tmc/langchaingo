package pgvector_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/averikitsch/langchaingo/chains"
	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/averikitsch/langchaingo/llms/googleai"
	"github.com/averikitsch/langchaingo/llms/openai"
	"github.com/averikitsch/langchaingo/schema"
	"github.com/averikitsch/langchaingo/vectorstores"
	"github.com/averikitsch/langchaingo/vectorstores/pgvector"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func preCheckEnvSetting(t *testing.T) string {
	t.Helper()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	pgvectorURL := os.Getenv("PGVECTOR_CONNECTION_STRING")
	if pgvectorURL == "" {
		pgVectorContainer, err := tcpostgres.RunContainer(
			context.Background(),
			testcontainers.WithImage("docker.io/pgvector/pgvector:pg16"),
			tcpostgres.WithDatabase("db_test"),
			tcpostgres.WithUsername("user"),
			tcpostgres.WithPassword("passw0rd!"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, pgVectorContainer.Terminate(context.Background()))
		})

		str, err := pgVectorContainer.ConnectionString(context.Background(), "sslmode=disable")
		require.NoError(t, err)

		pgvectorURL = str
	}

	return pgvectorURL
}

func makeNewCollectionName() string {
	return fmt.Sprintf("test-collection-%s", uuid.New().String())
}

func cleanupTestArtifacts(ctx context.Context, t *testing.T, s pgvector.Store, pgvectorURL string) {
	t.Helper()

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	tx, err := conn.Begin(ctx)
	require.NoError(t, err)

	require.NoError(t, s.RemoveCollection(ctx, tx))

	require.NoError(t, tx.Commit(ctx))
}

func TestPgvectorStoreRest(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

func TestPgvectorStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

func TestPgvectorStoreSimilarityScore(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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
	pgvectorURL := preCheckEnvSetting(t)
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

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(collectionName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

	store, err = pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(false),
		pgvector.WithCollectionName(collectionName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

func TestPgvectorAsRetriever(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

func TestPgvectorAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

func TestPgvectorAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

func TestPgvectorAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	require.NoError(t, err)
	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)
	defer conn.Close(ctx)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
		pgvector.WithCollectionTableName("collection_table_name"),
		pgvector.WithEmbeddingTableName("embedding_table_name"),
		pgvector.WithCollectionMetadata(map[string]any{
			"key": "value",
		}),
		pgvector.WithVectorDimensions(1536),
		pgvector.WithHNSWIndex(16, 64, "vector_l2_ops"),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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

	store, err = pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
		pgvector.WithCollectionTableName("collection_table_name1"),
		pgvector.WithEmbeddingTableName("embedding_table_name1"),
		pgvector.WithCollectionMetadata(map[string]any{
			"key": "value",
		}),
		pgvector.WithVectorDimensions(1536),
		pgvector.WithHNSWIndex(16, 64, "vector_l2_ops"),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

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
