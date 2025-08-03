package libsql_test

import (
	"context"
	"encoding/json"
	"fmt"

	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/libsql"
	_ "github.com/tursodatabase/go-libsql"
)

const (
	// Test index names and dimensions
	testIndexSize1536 = 1536                                     // Typical embedding size (e.g., OpenAI)
	testIndexSize3    = 3                                        // Small size for deterministic testing
	dsn               = "file:testx.db?mode=memory&cache=shared" // database for testing purpose using in memory and cached
)

type mockEmbedder struct{}

func (m *mockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return make([][]float32, len(texts)), nil
}

func (m *mockEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return make([]float32, 10), nil
}

func createOpenAIEmbedder(t *testing.T) *embeddings.EmbedderImpl {
	t.Helper()

	//t.Parallel() sqlite doesn't support parallel
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

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

func setupTestStore(t *testing.T) *libsql.Store {
	t.Helper()

	embedder := &mockEmbedder{}
	store, err := libsql.New(dsn, embedder, "", "", testIndexSize1536)
	require.NoError(t, err)

	require.NotNil(t, store)
	require.Equal(t, "langchain", store.TableName())
	require.Equal(t, "EMBEDDING_COLUMN", store.ColumnName())
	require.Equal(t, testIndexSize1536, store.VectorDim())

	return store
}

func TestStoreInitAndInsert(t *testing.T) {
	t.Parallel()
	store := setupTestStore(t)
	ctx := context.Background()

	id := "abc123"
	content := "LibSQL Go Vectorstore"
	meta := map[string]any{"source": "unit_test"}
	metaJSON, _ := json.Marshal(meta)
	embedding := make([]byte, 1536*4)

	insertQuery := fmt.Sprintf("INSERT INTO %s (id, content, metadata, %s) VALUES (?, ?, ?, ?);", store.TableName(), store.ColumnName())

	_, err := store.GetDB().ExecContext(ctx,
		insertQuery,
		id, content, string(metaJSON), embedding,
	)
	require.NoError(t, err)

	selectQuery := fmt.Sprintf("SELECT id, content, metadata, %s FROM %s WHERE id = ?;", store.ColumnName(), store.TableName())

	row := store.GetDB().QueryRowContext(ctx, selectQuery, id)
	var gotId, gotContent, gotMetadata string
	var gotEmbedding []byte
	err = row.Scan(&gotId, &gotContent, &gotMetadata, &gotEmbedding)
	require.NoError(t, err)

	require.Equal(t, id, gotId)
	require.Equal(t, content, gotContent)
	require.Equal(t, string(metaJSON), gotMetadata)
	require.Equal(t, embedding, gotEmbedding)
}

func TestAddDocumentsAndSimilaritySearch_WithOpenAI(t *testing.T) {
	ctx := context.Background()

	embedder := createOpenAIEmbedder(t)

	store, err := libsql.New(
		dsn,
		embedder,
		"similarity",
		"embedding",
		1536,
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Jakarta is the capital city of Indonesia", Metadata: map[string]any{"lang": "en"}},
		{PageContent: "Bandung is known for its universities and technology scene", Metadata: map[string]any{"lang": "en"}},
	}

	ids, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	sqlQuery := fmt.Sprintf("SELECT id, content, metadata FROM %s", store.TableName())
	rows, err := store.GetDB().Query(sqlQuery)
	require.NoError(t, err)
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, content, metadata string
		require.NoError(t, rows.Scan(&id, &content, &metadata))
		require.Contains(t, ids, id)
		require.Contains(t, []string{docs[0].PageContent, docs[1].PageContent}, content)
		var meta map[string]any
		require.NoError(t, json.Unmarshal([]byte(metadata), &meta))
		count++
	}
	require.Equal(t, 2, count)

	results, err := store.SimilaritySearch(ctx, "capital city of Indonesia", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Contains(t, results[0].PageContent, "Jakarta")
	require.InDelta(t, 0.0, results[0].Score, 1.0)
}

func TestDeleteByIDs(t *testing.T) {
	ctx := context.Background()
	embedder := createOpenAIEmbedder(t)

	store, err := libsql.New(
		dsn,
		embedder,
		"delete_by_id",
		"embedding",
		1536,
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Doc A", Metadata: map[string]any{"type": "a"}},
		{PageContent: "Doc B", Metadata: map[string]any{"type": "b"}},
	}
	ids, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	err = store.Delete(ctx, []string{ids[0]}, false)
	require.NoError(t, err)

	sqlQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", store.TableName())
	var count int
	require.NoError(t, store.GetDB().QueryRow(sqlQuery).Scan(&count))
	require.Equal(t, 1, count)

	err = store.Delete(ctx, []string{}, false)
	require.Error(t, err)
}

func TestDeleteAll(t *testing.T) {
	ctx := context.Background()
	embedder := createOpenAIEmbedder(t)

	store, err := libsql.New(
		dsn,
		embedder,
		"delete_all",
		"embedding",
		1536,
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Doc X", Metadata: map[string]any{"type": "x"}},
		{PageContent: "Doc Y", Metadata: map[string]any{"type": "y"}},
	}
	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	err = store.Delete(ctx, nil, true)
	require.NoError(t, err)

	sqlQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", store.TableName())
	var count int
	require.NoError(t, store.GetDB().QueryRow(sqlQuery).Scan(&count))
	require.Equal(t, 0, count)
}

func TestSimilaritySearchWithScore(t *testing.T) {
	ctx := context.Background()
	embedder := createOpenAIEmbedder(t)

	store, err := libsql.New(
		dsn,
		embedder,
		"similarity_score",
		"embedding",
		1536,
	)
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Jakarta is the capital city of Indonesia", Metadata: map[string]any{"lang": "en"}},
		{PageContent: "Bandung is known for its universities and technology scene", Metadata: map[string]any{"lang": "en"}},
	}
	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	results, err := store.SimilaritySearchWithScore(ctx, "capital of Indonesia", 1)

	require.NoError(t, err)
	require.NotEmpty(t, results)
	require.Contains(t, results[0].Doc.PageContent, "Jakarta")
	require.GreaterOrEqual(t, results[0].Score, float32(0))
}
