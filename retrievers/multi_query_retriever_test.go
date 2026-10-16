package retrievers

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/tools/scraper"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// the test is similar to langchain python version in https://python.langchain.com/docs/modules/data_connection/retrievers/MultiQueryRetriever
func TestMultiQueryRetriever(t *testing.T) { //nolint:funlen
	t.Parallel()
	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("must set GENAI_API_KEY to run test")
	}
	ctx := context.Background()
	pgvectorURL := os.Getenv("PGVECTOR_CONNECTION_STRING")
	if pgvectorURL == "" {
		postgresContainer, err := postgres.RunContainer(ctx,
			testcontainers.WithImage("docker.io/pgvector/pgvector:pg16"),
			postgres.WithDatabase("db_test"),
			postgres.WithUsername("user"),
			postgres.WithPassword("passw0rd!"),
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
			require.NoError(t, postgresContainer.Terminate(context.Background()))
		})
		pgvectorURL, err = postgresContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)
	}

	llm, err := googleai.New(ctx, googleai.WithAPIKey(genaiKey))
	require.NoError(t, err)

	// get web page data
	webScraper, err := scraper.New()
	require.NoError(t, err)
	data, err := webScraper.Call(ctx, "https://lilianweng.github.io/posts/2023-06-23-agent/")
	require.NoError(t, err)

	// split into chunks
	spliter := textsplitter.NewRecursiveCharacter(textsplitter.WithChunkSize(500), textsplitter.WithChunkOverlap(0))
	docs, err := documentloaders.NewText(strings.NewReader(data)).LoadAndSplit(ctx, spliter)
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(pgvectorURL),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName("test-multi-query-retriever"),
	)
	require.NoError(t, err)

	defer func() {
		conn, err := pgx.Connect(ctx, pgvectorURL)
		require.NoError(t, err)

		tx, err := conn.Begin(ctx)
		require.NoError(t, err)

		require.NoError(t, store.RemoveCollection(ctx, tx))

		require.NoError(t, tx.Commit(ctx))
		require.NoError(t, store.Close(ctx))
	}()
	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	question := "What are the approaches to Task Decomposition?"
	retriever := NewMultiQueryRetrieverFromLLM(vectorstores.ToRetriever(store, 5), llm, nil, true)
	retriever.DelayTime = 2

	results, err := retriever.GetRelevantDocuments(ctx, question)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	t.Logf("results: %#v", results)
}
