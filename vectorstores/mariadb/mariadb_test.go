package mariadb_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	mariacontainer "github.com/testcontainers/testcontainers-go/modules/mariadb"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/mariadb"
)

func setupTest(t *testing.T) (string, embeddings.Embedder) {
	t.Helper()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	return setupMariaDBContainer(t), setupEmbedder(t)
}

func setupMariaDBContainer(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	mariadbURL := os.Getenv("MARIADB_CONNECTION_STRING")

	if mariadbURL != "" {
		return mariadbURL
	}

	container, err := mariacontainer.Run(
		context.Background(),
		"mariadb:11.7.1-ubi-rc",
		mariacontainer.WithDatabase("testdb"),
		mariacontainer.WithUsername("test"),
		mariacontainer.WithPassword("test"),
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
		require.NoError(t, container.Terminate(context.Background()))
	})

	dsn, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	return dsn
}

func setupEmbedder(t *testing.T) embeddings.Embedder {
	t.Helper()

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	return e
}

func makeNewCollectionName() string {
	return fmt.Sprintf("test-collection-%s", uuid.New().String())
}

func cleanup(ctx context.Context, t *testing.T, store *mariadb.Store, dsn string) {
	t.Helper()

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	err = store.RemoveCollection(ctx, tx)
	require.NoError(t, err)
}

func TestMariaDBStoreRestBasic(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)
	defer cleanup(ctx, t, store, dsn)

	docs := []schema.Document{
		{
			PageContent: "Vladivostok",
			Metadata:    map[string]any{"Country": "Russia"},
		},
		{
			PageContent: "Moscow",
			Metadata:    map[string]any{"Country": "Russia"},
		},
		{
			PageContent: "New York",
			Metadata:    map[string]any{"Country": "USA"},
		},
		{
			PageContent: "London",
			Metadata:    map[string]any{"Country": "England"},
		},
		{PageContent: "Potato"},
		{PageContent: "Cabbage"},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "England", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	doc := results[0]
	require.Equal(t, "London", doc.PageContent)
	require.Equal(t, "England", doc.Metadata["Country"])
}

func TestMariaDBEmptyCollection(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()

	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)
	defer cleanup(ctx, t, store, dsn)

	results, err := store.SimilaritySearch(ctx, "any query", 5)
	require.NoError(t, err)
	require.Empty(t, results, "expected no results for empty collection")
}

func TestMariaDBNonExistentFilter(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()

	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)
	defer cleanup(ctx, t, store, dsn)

	docs := []schema.Document{
		{PageContent: "Berlin", Metadata: map[string]any{"country": "germany"}},
		{PageContent: "Munich", Metadata: map[string]any{"country": "germany"}},
	}
	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	filter := map[string]any{"country": "france"}
	results, err := store.SimilaritySearch(ctx, "Berlin", 5, vectorstores.WithFilters(filter))
	require.NoError(t, err)
	require.Empty(t, results, "expected no results for non-matching filter")
}

func TestMariaDBStoreWithScoreThreshold(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()

	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "Yokohama", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "Osaka", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "Nagoya", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "Sapporo", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "Fukuoka", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "Dublin", Metadata: map[string]any{"country": "ireland"}},
		{PageContent: "Paris", Metadata: map[string]any{"country": "france"}},
		{PageContent: "London", Metadata: map[string]any{"country": "uk"}},
		{PageContent: "New York", Metadata: map[string]any{"country": "usa"}},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "Which of these are cities in Japan", 10, vectorstores.WithScoreThreshold(0.8))
	require.NoError(t, err)
	require.Len(t, results, 6)

	results, err = store.SimilaritySearch(ctx, "Which of these are cities in Japan", 10, vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, results, 10)
}

func TestLongDocumentHandling(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()

	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)
	defer cleanup(ctx, t, store, dsn)

	longText := strings.Repeat("This is a long document. ", 100)
	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: longText},
	})
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "long document", 1)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	require.Contains(t, results[0].PageContent, "long document")
}

func TestMariaDBStoreSimilarityScore(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Tokyo is the capital city of Japan."},
		{PageContent: "Paris is known as the city of love."},
		{PageContent: "I enjoy visiting London."},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "What is the capital city of Japan?", 3, vectorstores.WithScoreThreshold(0.8))
	require.NoError(t, err)
	require.Len(t, results, 1)

	require.True(t, results[0].Score > 0.9)
}

func TestMariaDBAsRetriever(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	llm, err := openai.New()
	require.NoError(t, err)
	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "The color of the house is blue."},
		{PageContent: "The color of the car is red."},
		{PageContent: "The color of the desk is orange."},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(llm, vectorstores.ToRetriever(store, 1)),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result), "orange")
}

func TestMariaDBDeduplicater(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{"type": "city"}},
		{PageContent: "potato", Metadata: map[string]any{"type": "vegetable"}},
	}, vectorstores.WithDeduplicater(func(_ context.Context, doc schema.Document) bool {
		return doc.PageContent == "tokyo"
	}))
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "any query", 2)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "potato", results[0].PageContent)
}

func TestMariaDBWithMetadataFilters(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "in kitchen, the lamp is black.", Metadata: map[string]any{"location": "kitchen"}},
		{PageContent: "in bedroom, the lamp is blue.", Metadata: map[string]any{"location": "bedroom"}},
		{PageContent: "in office, the lamp is orange.", Metadata: map[string]any{"location": "office"}},
		{PageContent: "in sitting room, the lamp is purple.", Metadata: map[string]any{"location": "sitting room"}},
		{PageContent: "in patio, the lamp is yellow.", Metadata: map[string]any{"location": "patio"}},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	filter := map[string]any{"location": "office"}
	results, err := store.SimilaritySearch(ctx, "What color is the lamp?", 5, vectorstores.WithFilters(filter))
	require.NoError(t, err)
	require.Len(t, results, 1)

	require.Contains(t, strings.ToLower(results[0].PageContent), "office")
}

func TestMariaDBWithAllOptions(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)
	ctx := context.Background()

	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
		mariadb.WithCollectionMetadata(map[string]any{"key": "value"}),
	)
	require.NoError(t, err)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{"country": "japan"}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "tokyo", results[0].PageContent)
	require.Equal(t, "japan", results[0].Metadata["country"])
}

func TestMariaDBUpdateDocument(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)
	defer cleanup(ctx, t, store, dsn)

	docs := []schema.Document{
		{
			PageContent: "Vladivostok",
			Metadata:    map[string]any{"Country": "Russia"},
		},
		{
			PageContent: "Moscow",
			Metadata:    map[string]any{"Country": "Russia"},
		},
	}

	ids, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	updatedDoc := schema.Document{
		PageContent: "Saint Petersburg",
		Metadata:    map[string]any{"Country": "Russia"},
	}

	err = store.UpdateDocument(ctx, ids[0], updatedDoc)
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "Russia", 2)
	require.NoError(t, err)
	require.Len(t, results, 2)

	found := false
	for _, doc := range results {
		if doc.PageContent == "Saint Petersburg" && doc.Metadata["Country"] == "Russia" {
			found = true
			break
		}
	}
	require.True(t, found, "updated document not found")
}

func TestMariaDBDeleteDocumentsByFilter(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)
	defer cleanup(ctx, t, store, dsn)

	docs := []schema.Document{
		{
			PageContent: "Vladivostok",
			Metadata:    map[string]any{"Country": "Russia"},
		},
		{
			PageContent: "Moscow",
			Metadata:    map[string]any{"Country": "Russia"},
		},
		{
			PageContent: "Paris",
			Metadata:    map[string]any{"Country": "France"},
		},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	results, err := store.SimilaritySearch(ctx, "Russia", 5, vectorstores.WithScoreThreshold(0.85))
	require.NoError(t, err)
	require.Len(t, results, 2, "expected 2 documents")

	filter := map[string]any{"Country": "Russia"}
	deletedCount, err := store.DeleteDocumentsByFilter(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, int64(2), deletedCount)

	results, err = store.SimilaritySearch(ctx, "Russia", 5, vectorstores.WithScoreThreshold(0.85))
	require.NoError(t, err)
	require.Empty(t, results, "expected no results for deleted documents")

	results, err = store.SimilaritySearch(ctx, "France", 5, vectorstores.WithScoreThreshold(0.85))
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Paris", results[0].PageContent)
	require.Equal(t, "France", results[0].Metadata["Country"])
}

func TestMariaDBSearch(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "in kitchen, the lamp is black.", Metadata: map[string]any{"location": "kitchen"}},
		{PageContent: "in bedroom, the lamp is blue.", Metadata: map[string]any{"location": "bedroom"}},
		{PageContent: "in office, the lamp is orange.", Metadata: map[string]any{"location": "office"}},
		{PageContent: "in sitting room, the lamp is purple.", Metadata: map[string]any{"location": "sitting room"}},
		{PageContent: "in patio, the lamp is yellow.", Metadata: map[string]any{"location": "patio"}},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	filter := map[string]any{"location": "office"}
	results, err := store.Search(ctx, filter, 5)
	require.NoError(t, err)
	require.Len(t, results, 1)

	require.Contains(t, strings.ToLower(results[0].PageContent), "office")
}

func TestMariaDBSearchWithAdvancedFilters(t *testing.T) {
	t.Parallel()
	dsn, e := setupTest(t)

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN(dsn),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
		mariadb.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Document 1", Metadata: map[string]any{"value": 10, "category": "A"}},
		{PageContent: "Document 2", Metadata: map[string]any{"value": 20, "category": "B"}},
		{PageContent: "Document 3", Metadata: map[string]any{"value": 30, "category": "A"}},
		{PageContent: "Document 4", Metadata: map[string]any{"value": 40, "category": "C"}},
		{PageContent: "Document 5", Metadata: map[string]any{"value": 50, "category": "B"}},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	tests := []struct {
		name     string
		filter   map[string]any
		expected int
	}{
		{
			name:     "Equal filter",
			filter:   map[string]any{"category": "A"},
			expected: 2,
		},
		{
			name:     "Greater than filter",
			filter:   map[string]any{"value >": 30},
			expected: 2,
		},
		{
			name:     "Less than filter",
			filter:   map[string]any{"value <": 30},
			expected: 2,
		},
		{
			name:     "Not equal filter",
			filter:   map[string]any{"category !=": "A"},
			expected: 3,
		},
		{
			name:     "Combined filters",
			filter:   map[string]any{"category": "B", "value >": 20},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			results, err := store.Search(ctx, tt.filter, 5)
			require.NoError(t, err)
			require.Len(t, results, tt.expected)
		})
	}
}
