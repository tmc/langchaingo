package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/0xDezzy/langchaingo/embeddings"
	"github.com/0xDezzy/langchaingo/llms/googleai"
	"github.com/0xDezzy/langchaingo/llms/googleai/vertex"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/util/cloudsqlutil"
	"github.com/0xDezzy/langchaingo/vectorstores"
	"github.com/0xDezzy/langchaingo/vectorstores/cloudsql"
)

func getEnvVariables() (string, string, string, string, string, string, string, string) {
	// Requires environment variable POSTGRES_USERNAME to be set.
	username := os.Getenv("POSTGRES_USERNAME")
	if username == "" {
		log.Fatal("env variable POSTGRES_USERNAME is empty")
	}
	// Requires environment variable POSTGRES_PASSWORD to be set.
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		log.Fatal("env variable POSTGRES_PASSWORD is empty")
	}
	// Requires environment variable POSTGRES_DATABASE to be set.
	database := os.Getenv("POSTGRES_DATABASE")
	if database == "" {
		log.Fatal("env variable POSTGRES_DATABASE is empty")
	}
	// Requires environment variable PROJECT_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("env variable PROJECT_ID is empty")
	}
	// Requires environment variable POSTGRES_REGION to be set.
	region := os.Getenv("POSTGRES_REGION")
	if region == "" {
		log.Fatal("env variable POSTGRES_REGION is empty")
	}
	// Requires environment variable POSTGRES_INSTANCE to be set.
	instance := os.Getenv("POSTGRES_INSTANCE")
	if instance == "" {
		log.Fatal("env variable POSTGRES_INSTANCE is empty")
	}
	// Requires environment variable POSTGRES_TABLE to be set.
	table := os.Getenv("POSTGRES_TABLE")
	if table == "" {
		log.Fatal("env variable POSTGRES_TABLE is empty")
	}

	// Requires environment variable GOOGLE_CLOUD_LOCATION to be set.
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		log.Fatal("env variable GOOGLE_CLOUD_LOCATION is empty")
	}

	return username, password, database, projectID, region, instance, table, location
}

func main() {
	// Requires the Environment variables to be set as indicated in the getEnvVariables function.
	username, password, database, projectID, region, instance, table, cloudLocation := getEnvVariables()
	ctx := context.Background()

	pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
		cloudsqlutil.WithUser(username),
		cloudsqlutil.WithPassword(password),
		cloudsqlutil.WithDatabase(database),
		cloudsqlutil.WithCloudSQLInstance(projectID, region, instance),
		cloudsqlutil.WithIPType("PUBLIC"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize table for the Vectorstore to use. You only need to do this the first time you use this table.
	vectorstoreTableoptions := cloudsqlutil.VectorstoreTableOptions{
		TableName:         table,
		VectorSize:        768,
		StoreMetadata:     true,
		OverwriteExisting: true,
		MetadataColumns: []cloudsqlutil.Column{
			{
				Name:     "area",
				DataType: "int",
			},
			{
				Name:     "population",
				DataType: "int",
			},
		},
	}
	err = pgEngine.InitVectorstoreTable(ctx, vectorstoreTableoptions)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize VertexAI LLM
	llm, err := vertex.New(ctx, googleai.WithCloudProject(projectID), googleai.WithCloudLocation(cloudLocation), googleai.WithDefaultModel("text-embedding-005"))
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Vectorstore
	vs, err := cloudsql.NewVectorStore(pgEngine, e, table, cloudsql.WithMetadataColumns([]string{"area", "population"}))
	if err != nil {
		log.Fatal(err)
	}

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
			PageContent: "Sao Paulo",
			Metadata: map[string]any{
				"population": 22.6,
				"area":       1523,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	docs, err := vs.SimilaritySearch(ctx, "Japan", 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Docs:", docs)
	filter := "\"area\" > 1500"
	filteredDocs, err := vs.SimilaritySearch(ctx, "Japan", 0, vectorstores.WithFilters(filter))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("FilteredDocs:", filteredDocs)
}
