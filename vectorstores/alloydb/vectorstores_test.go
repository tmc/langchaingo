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
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/alloydb"
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

func setEngine(t *testing.T) (alloydbutil.PostgresEngine, error) {
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

	return *pgEngine, nil
}

func setVectoreStore(t *testing.T) (alloydb.VectorStore, error) {
	pgEngine, err := setEngine(t)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
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
	vectorstoreTableoptions := &alloydbutil.VectorstoreTableOptions{
		TableName:  "table",
		VectorSize: 768,
	}

	if err != nil {
		log.Fatal(err)
	}

	err = pgEngine.InitVectorstoreTable(ctx, *vectorstoreTableoptions)

	vs, err := alloydb.NewVectorStore(ctx, pgEngine, e, "items")
	if err != nil {
		t.Fatal(err)
	}
	return vs, nil
}

func TestPingToDB(t *testing.T) {
	engine, err := setEngine(t)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	if err = engine.Pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestApplyVectorIndexAndDropIndex(t *testing.T) {
	vs, err := setVectoreStore(t)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	idx := vs.NewBaseIndex("testindex", "hnsw", alloydb.CosineDistance{}, []string{}, alloydb.HNSWOptions{})
	err = vs.ApplyVectorIndex(ctx, idx, "testindex", false, false)
	if err != nil {
		t.Fatal(err)
	}
	err = vs.DropVectorIndex(ctx, "testindex", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsValidIndex(t *testing.T) {
	vs, err := setVectoreStore(t)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	idx := vs.NewBaseIndex("testindex", "hnsw", alloydb.CosineDistance{}, []string{}, alloydb.HNSWOptions{})
	err = vs.ApplyVectorIndex(ctx, idx, "testindex", false, false)
	if err != nil {
		t.Fatal(err)
	}
	isValid, err := vs.IsValidIndex(ctx, "testindex")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(isValid)
	err = vs.DropVectorIndex(ctx, "testindex", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddDocuments(t *testing.T) {
	vs, err := setVectoreStore(t)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, err = vs.AddDocuments(ctx, []schema.Document{
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
}
