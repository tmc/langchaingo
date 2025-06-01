// nolint
package cloudsql_test

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/util/cloudsqlutil"
	"github.com/tmc/langchaingo/vectorstores/cloudsql"
	"os"
	"strings"
	"testing"
	"time"
)

func preCheckEnvSetting(t *testing.T) string {
	t.Helper()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	pgvectorURL := os.Getenv("PGVECTOR_CONNECTION_STRING")
	if pgvectorURL == "" {
		pgVectorContainer, err := tcpostgres.Run(
			t.Context(),
			"docker.io/pgvector/pgvector:pg16",
			tcpostgres.WithDatabase("db_test"),
			tcpostgres.WithUsername("user"),
			tcpostgres.WithPassword("passw0rd!"),
			testcontainers.WithWaitStrategy(
				wait.ForAll(
					wait.ForLog("database system is ready to accept connections").
						WithOccurrence(2).
						WithStartupTimeout(60*time.Second),
					wait.ForListeningPort("5432/tcp").
						WithStartupTimeout(60*time.Second),
				)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			if err := pgVectorContainer.Terminate(context.Background()); err != nil {
				t.Logf("Failed to terminate cloudsql container: %v", err)
			}
		})

		str, err := pgVectorContainer.ConnectionString(t.Context(), "sslmode=disable")
		require.NoError(t, err)

		pgvectorURL = str

		// Give the container a moment to fully initialize
		time.Sleep(2 * time.Second)
	}

	return pgvectorURL
}

func setEngineWithImage(t *testing.T) cloudsqlutil.PostgresEngine {
	t.Helper()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()
	myPool, err := pgxpool.New(ctx, pgvectorURL)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}
	// Call NewPostgresEngine to initialize the database connection
	pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
		cloudsqlutil.WithPool(myPool),
	)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}

	return pgEngine
}

func initVectorStore(t *testing.T) (cloudsql.VectorStore, func() error) {
	t.Helper()
	pgEngine := setEngineWithImage(t)
	ctx := t.Context()
	vectorstoreTableoptions := cloudsqlutil.VectorstoreTableOptions{
		TableName:         "my_test_table",
		OverwriteExisting: true,
		VectorSize:        1536,
		StoreMetadata:     true,
	}
	err := pgEngine.InitVectorstoreTable(ctx, vectorstoreTableoptions)
	if err != nil {
		t.Fatal(err)
	}
	// Initialize VertexAI LLM
	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	if err != nil {
		t.Fatal(err)
	}
	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		t.Fatal(err)
	}
	vs, err := cloudsql.NewVectorStore(pgEngine, e, "my_test_table")
	if err != nil {
		t.Fatal(err)
	}

	cleanUpTableFn := func() error {
		_, err := pgEngine.Pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", "my_test_table"))
		return err
	}
	return vs, cleanUpTableFn
}

func TestContainerPingToDB(t *testing.T) {
	t.Parallel()
	engine := setEngineWithImage(t)

	defer engine.Close()

	if err := engine.Pool.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestContainerApplyVectorIndexAndDropIndex(t *testing.T) {
	t.Parallel()
	vs, cleanUpTableFn := initVectorStore(t)
	ctx := t.Context()
	idx := vs.NewBaseIndex("testindex", "hnsw", cloudsql.CosineDistance{}, []string{}, cloudsql.HNSWOptions{M: 4, EfConstruction: 16})
	err := vs.ApplyVectorIndex(ctx, idx, "testindex", false)
	if err != nil {
		t.Fatal(err)
	}
	err = vs.DropVectorIndex(ctx, "testindex")
	if err != nil {
		t.Fatal(err)
	}
	err = cleanUpTableFn()
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainerIsValidIndex(t *testing.T) {
	t.Parallel()
	vs, cleanUpTableFn := initVectorStore(t)
	ctx := t.Context()
	idx := vs.NewBaseIndex("testindex", "hnsw", cloudsql.CosineDistance{}, []string{}, cloudsql.HNSWOptions{M: 4, EfConstruction: 16})
	err := vs.ApplyVectorIndex(ctx, idx, "testindex", false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vs.IsValidIndex(ctx, "testindex")
	if err != nil {
		t.Fatal(err)
	}
	err = vs.DropVectorIndex(ctx, "testindex")
	if err != nil {
		t.Fatal(err)
	}
	err = cleanUpTableFn()
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainerAddDocuments(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	vs, cleanUpTableFn := initVectorStore(t)

	_, err := vs.AddDocuments(ctx, []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population": 38,
				"area":       2190,
			},
		},
		{
			PageContent: "Paris",
			Metadata: map[string]any{
				"population": 11,
				"area":       105,
			},
		},
		{
			PageContent: "London",
			Metadata: map[string]any{
				"population": 9.5,
				"area":       1572,
			},
		},
		{
			PageContent: "Santiago",
			Metadata: map[string]any{
				"population": 6.9,
				"area":       641,
			},
		},
		{
			PageContent: "Buenos Aires",
			Metadata: map[string]any{
				"population": 15.5,
				"area":       203,
			},
		},
		{
			PageContent: "Rio de Janeiro",
			Metadata: map[string]any{
				"population": 13.7,
				"area":       1200,
			},
		},
		{
			PageContent: "Sao Paulo",
			Metadata: map[string]any{
				"population": 22.6,
				"area":       1523,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = cleanUpTableFn()
	if err != nil {
		t.Fatal(err)
	}
}
