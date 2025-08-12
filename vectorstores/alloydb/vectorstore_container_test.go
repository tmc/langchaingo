package alloydb_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/util/alloydbutil"
	"github.com/vendasta/langchaingo/vectorstores/alloydb"
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

func setEngineWithImage(t *testing.T) alloydbutil.PostgresEngine {
	t.Helper()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()
	myPool, err := pgxpool.New(ctx, pgvectorURL)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}
	// Call NewPostgresEngine to initialize the database connection
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithPool(myPool),
	)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}

	return pgEngine
}

func initVectorStore(t *testing.T) (alloydb.VectorStore, func() error) {
	t.Helper()
	pgEngine := setEngineWithImage(t)
	ctx := context.Background()
	vectorstoreTableoptions := alloydbutil.VectorstoreTableOptions{
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
	vs, err := alloydb.NewVectorStore(pgEngine, e, "my_test_table")
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

	if err := engine.Pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestContainerApplyVectorIndexAndDropIndex(t *testing.T) {
	t.Parallel()
	vs, cleanUpTableFn := initVectorStore(t)
	ctx := context.Background()
	idx := vs.NewBaseIndex("testindex", "hnsw", alloydb.CosineDistance{}, []string{}, alloydb.HNSWOptions{M: 4, EfConstruction: 16})
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
	ctx := context.Background()
	idx := vs.NewBaseIndex("testindex", "hnsw", alloydb.CosineDistance{}, []string{}, alloydb.HNSWOptions{M: 4, EfConstruction: 16})
	err := vs.ApplyVectorIndex(ctx, idx, "testindex", false)
	if err != nil {
		t.Fatal(err)
	}
	isValid, err := vs.IsValidIndex(ctx, "testindex")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(isValid)
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
	ctx := context.Background()
	vs, cleanUpTableFn := initVectorStore(t)
	t.Cleanup(func() {
		if err := cleanUpTableFn(); err != nil {
			t.Fatal("Cleanup failed:", err)
		}
	})

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
