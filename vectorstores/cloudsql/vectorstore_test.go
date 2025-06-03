// nolint
package cloudsql_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/util/cloudsqlutil"
	"github.com/tmc/langchaingo/vectorstores/cloudsql"
)

type EnvVariables struct {
	Username  string
	Password  string
	Database  string
	ProjectID string
	Region    string
	Instance  string
	Cluster   string
	Table     string
}

func getEnvVariables(t *testing.T) EnvVariables {
	t.Helper()

	username := os.Getenv("CLOUDSQL_USERNAME")
	if username == "" {
		t.Skip("env variable CLOUDSQL_USERNAME is empty")
	}
	// Requires environment variable CLOUDSQL_PASSWORD to be set.
	password := os.Getenv("CLOUDSQL_PASSWORD")
	if password == "" {
		t.Skip("env variable CLOUDSQL_PASSWORD is empty")
	}
	// Requires environment variable CLOUDSQL_DATABASE to be set.
	database := os.Getenv("CLOUDSQL_DATABASE")
	if database == "" {
		t.Skip("env variable CLOUDSQL_DATABASE is empty")
	}
	// Requires environment variable CLOUDSQL_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		t.Skip("env variable PROJECT_ID is empty")
	}
	// Requires environment variable ALLOYDB_REGION to be set.
	region := os.Getenv("CLOUDSQL_REGION")
	if region == "" {
		t.Skip("env variable CLOUDSQL_REGION is empty")
	}
	// Requires environment variable ALLOYDB_INSTANCE to be set.
	instance := os.Getenv("CLOUDSQL_INSTANCE")
	if instance == "" {
		t.Skip("env variable CLOUDSQL_INSTANCE is empty")
	}
	// Requires environment variable CLOUDSQL_CLUSTER to be set.
	cluster := os.Getenv("CLOUDSQL_CLUSTER")
	if cluster == "" {
		t.Skip("env variable CLOUDSQL_CLUSTER is empty")
	}
	// Requires environment variable CLOUDSQL_TABLE to be set.
	table := os.Getenv("CLOUDSQL_TABLE")
	if table == "" {
		t.Skip("env variable CLOUDSQL_TABLE is empty")
	}

	envVariables := EnvVariables{
		Username:  username,
		Password:  password,
		Database:  database,
		ProjectID: projectID,
		Region:    region,
		Instance:  instance,
		Cluster:   cluster,
		Table:     table,
	}

	return envVariables
}

func setEngine(t *testing.T, envVariables EnvVariables) cloudsqlutil.PostgresEngine {
	t.Helper()
	ctx := context.Background()
	pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
		cloudsqlutil.WithUser(envVariables.Username),
		cloudsqlutil.WithPassword(envVariables.Password),
		cloudsqlutil.WithDatabase(envVariables.Database),
		cloudsqlutil.WithCloudSQLInstance(envVariables.ProjectID, envVariables.Region, envVariables.Instance),
	)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}

	return pgEngine
}

func vectorStore(t *testing.T, envVariables EnvVariables) (cloudsql.VectorStore, func() error) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping cloudsql tests in short mode")
	}
	pgEngine := setEngine(t, envVariables)
	ctx := context.Background()
	vectorstoreTableoptions := cloudsqlutil.VectorstoreTableOptions{
		TableName:         envVariables.Table,
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
	vs, err := cloudsql.NewVectorStore(pgEngine, e, envVariables.Table)
	if err != nil {
		t.Fatal(err)
	}

	cleanUpTableFn := func() error {
		_, err := pgEngine.Pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", envVariables.Table))
		return err
	}
	return vs, cleanUpTableFn
}

func TestApplyVectorIndexAndDropIndex(t *testing.T) {
	t.Parallel()
	envVariables := getEnvVariables(t)
	vs, cleanUpTableFn := vectorStore(t, envVariables)
	ctx := context.Background()
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

func TestIsValidIndex(t *testing.T) {
	t.Parallel()
	envVariables := getEnvVariables(t)
	vs, cleanUpTableFn := vectorStore(t, envVariables)
	ctx := context.Background()
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

func TestAddDocuments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	envVariables := getEnvVariables(t)
	vs, cleanUpTableFn := vectorStore(t, envVariables)

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
