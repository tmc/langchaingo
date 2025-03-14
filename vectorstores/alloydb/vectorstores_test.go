package alloydb_test

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/alloydb"
	"os"
	"testing"
)

func getEnvVariables(t *testing.T) (string, string, string, string, string, string, string) {
	t.Helper()

	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		t.Skip("ALLOYDB_USERNAME environment variable not set")
	}
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		t.Skip("ALLOYDB_PASSWORD environment variable not set")
	}
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		t.Skip("ALLOYDB_DATABASE environment variable not set")
	}
	projectID := os.Getenv("ALLOYDB_PROJECT_ID")
	if projectID == "" {
		t.Skip("ALLOYDB_PROJECT_ID environment variable not set")
	}
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		t.Skip("ALLOYDB_REGION environment variable not set")
	}
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		t.Skip("ALLOYDB_INSTANCE environment variable not set")
	}
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		t.Skip("ALLOYDB_CLUSTER environment variable not set")
	}

	return username, password, database, projectID, region, instance, cluster
}

func setEngine(t *testing.T) (alloydbutil.PostgresEngine, error) {
	username, password, database, projectID, region, instance, cluster := getEnvVariables(t)
	ctx := context.Background()
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser(username),
		alloydbutil.WithPassword(password),
		alloydbutil.WithDatabase(database),
		alloydbutil.WithAlloyDBInstance(projectID, region, cluster, instance),
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
	llmm, err := ollama.New(
		ollama.WithModel("llama3"),
	)
	if err != nil {
		t.Fatal(err)
	}
	e, err := embeddings.NewEmbedder(llmm)
	if err != nil {
		t.Fatal(err)
	}
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
	idx := vs.NewBaseIndex("testindex", "hnsw", alloydb.CosineDistance{}, []string{})
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
	idx := vs.NewBaseIndex("testindex", "hnsw", alloydb.CosineDistance{}, []string{})
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
