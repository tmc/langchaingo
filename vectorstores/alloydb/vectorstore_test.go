//nolint:paralleltest
package alloydb_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/util/alloydbutil"
	"github.com/yincongcyincong/langchaingo/vectorstores/alloydb"
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
	
	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		t.Skip("env variable ALLOYDB_USERNAME is empty")
	}
	// Requires environment variable ALLOYDB_PASSWORD to be set.
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		t.Skip("env variable ALLOYDB_PASSWORD is empty")
	}
	// Requires environment variable ALLOYDB_DATABASE to be set.
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		t.Skip("env variable ALLOYDB_DATABASE is empty")
	}
	// Requires environment variable PROJECT_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		t.Skip("env variable PROJECT_ID is empty")
	}
	// Requires environment variable ALLOYDB_REGION to be set.
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		t.Skip("env variable ALLOYDB_REGION is empty")
	}
	// Requires environment variable ALLOYDB_INSTANCE to be set.
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		t.Skip("env variable ALLOYDB_INSTANCE is empty")
	}
	// Requires environment variable ALLOYDB_CLUSTER to be set.
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		t.Skip("env variable ALLOYDB_CLUSTER is empty")
	}
	// Requires environment variable ALLOYDB_TABLE to be set.
	table := os.Getenv("ALLOYDB_TABLE")
	if table == "" {
		t.Skip("env variable ALLOYDB_TABLE is empty")
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

func setEngine(t *testing.T, envVariables EnvVariables) alloydbutil.PostgresEngine {
	t.Helper()
	ctx := context.Background()
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser(envVariables.Username),
		alloydbutil.WithPassword(envVariables.Password),
		alloydbutil.WithDatabase(envVariables.Database),
		alloydbutil.WithAlloyDBInstance(envVariables.ProjectID, envVariables.Region, envVariables.Cluster, envVariables.Instance),
	)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}
	
	return pgEngine
}

func vectorStore(t *testing.T, envVariables EnvVariables) (alloydb.VectorStore, func() error) {
	t.Helper()
	pgEngine := setEngine(t, envVariables)
	ctx := context.Background()
	vectorstoreTableoptions := alloydbutil.VectorstoreTableOptions{
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
		log.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		t.Fatal(err)
	}
	vs, err := alloydb.NewVectorStore(pgEngine, e, envVariables.Table)
	if err != nil {
		t.Fatal(err)
	}
	
	cleanUpTableFn := func() error {
		_, err := pgEngine.Pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", envVariables.Table))
		return err
	}
	return vs, cleanUpTableFn
}

func TestPingToDB(t *testing.T) {
	envVariables := getEnvVariables(t)
	engine := setEngine(t, envVariables)
	
	defer engine.Close()
	
	if err := engine.Pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestApplyVectorIndexAndDropIndex(t *testing.T) {
	envVariables := getEnvVariables(t)
	vs, cleanUpTableFn := vectorStore(t, envVariables)
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

func TestIsValidIndex(t *testing.T) {
	envVariables := getEnvVariables(t)
	vs, cleanUpTableFn := vectorStore(t, envVariables)
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

func TestAddDocuments(t *testing.T) {
	ctx := context.Background()
	envVariables := getEnvVariables(t)
	vs, cleanUpTableFn := vectorStore(t, envVariables)
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
